output "api_endpoint" {
  description = "Base URL of the API. The $default stage adds no path prefix."
  value       = aws_apigatewayv2_api.http.api_endpoint
}

output "api_id" {
  value = aws_apigatewayv2_api.http.id
}

output "execution_arn" {
  description = "Execution ARN, for scoping IAM policies or extra invoke permissions."
  value       = aws_apigatewayv2_api.http.execution_arn
}
