variable "project_name" {
  description = "Name prefix applied to the API and its resources."
  type        = string
}

variable "allowed_origins" {
  description = <<-EOT
    Origins allowed to call the API from a browser. Each entry must carry the
    scheme, the host and the port with no trailing slash, because the browser
    compares them literally: "http://localhost:5173" matches, "localhost:5173"
    and "http://localhost:5173/" do not.
  EOT
  type        = list(string)
}

variable "lambda_invoke_arn" {
  description = "invoke_arn of the API Lambda, used as the proxy integration target."
  type        = string
}

variable "lambda_function_name" {
  description = "Name of the API Lambda, needed to grant API Gateway permission to invoke it."
  type        = string
}

variable "cors_max_age" {
  description = <<-EOT
    Seconds a browser may cache the preflight result. Without it every request
    carrying an Authorization header pays for an extra OPTIONS round trip.
  EOT
  type        = number
  default     = 3600
}
