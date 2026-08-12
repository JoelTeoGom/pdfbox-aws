locals {
  # These strings are matched verbatim against req.RouteKey in the Go router,
  # path parameter names included: the handler reads PathParameters["fileID"],
  # so renaming the placeholder here silently breaks the lookup.
  routes = [
    "POST /files",
    "GET /files",
    "GET /files/{fileID}",
    "DELETE /files/{fileID}",
  ]
}

resource "aws_apigatewayv2_api" "http" {
  name          = "${var.project_name}-api"
  protocol_type = "HTTP"

  # With this block API Gateway answers preflight OPTIONS requests itself and
  # attaches the CORS headers to the responses of the routes below, so the
  # Lambda never sees them. That matters here: a preflight carries no
  # Authorization header by design, so it would fail the token check the router
  # runs before dispatching. For the same reason no OPTIONS route is declared
  # below — an explicit route would take precedence over this handling.
  cors_configuration {
    allow_origins = var.allowed_origins
    allow_methods = ["GET", "POST", "DELETE", "OPTIONS"]

    # Authorization is what makes every request non-simple and triggers the
    # preflight in the first place. Content-Type covers the JSON body of POST.
    allow_headers = ["Authorization", "Content-Type"]

    # Only for cookie based sessions. This API authenticates with a bearer
    # token in a header, and enabling it would also forbid a wildcard origin.
    allow_credentials = false

    max_age = var.cors_max_age
  }
}

resource "aws_apigatewayv2_integration" "lambda" {
  api_id           = aws_apigatewayv2_api.http.id
  integration_type = "AWS_PROXY"
  integration_uri  = var.lambda_invoke_arn

  # Always POST for a proxy integration: it describes how API Gateway talks to
  # Lambda, not the method the client used.
  integration_method = "POST"

  # 2.0 is the event shape the handler unmarshals into
  # events.APIGatewayV2HTTPRequest. Format 1.0 carries no RouteKey and would
  # send every request to the router's default branch.
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "routes" {
  for_each = toset(local.routes)

  api_id    = aws_apigatewayv2_api.http.id
  route_key = each.value
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id = aws_apigatewayv2_api.http.id

  # The $default stage serves the API without a stage prefix in the path.
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowInvokeFromApiGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "apigateway.amazonaws.com"

  # Scoped to this API, but open across its stages and routes so adding a route
  # does not require a matching permission.
  source_arn = "${aws_apigatewayv2_api.http.execution_arn}/*/*"
}
