package flash_http

const (
	HTTP_CONTENT_TYPE     = "Content-type"
	HTTP_ACCEPT           = "Accept"
	HTTP_APPLICATION_JSON = "application/json"
	HTTP_CONTENT_FORM_URL  = "application/x-www-form-urlencoded"
	HTTP_CONTENT_TEXT_PLAIN = "text/plain"
)

const (
	HTTP_REQUEST_TYPE_GET    = "GET"
	HTTP_REQUEST_TYPE_POST   = "POST"
	HTTP_REQUEST_TYPE_PUT    = "PUT"
	HTTP_REQUEST_TYPE_DELETE = "DELETE"
)

const (
	AUTH_KEY_USERNAME  = "username"
	AUTH_KEY_PASSWORD  = "password"
)

const (
	ERROR_TIMEOUT         = "hystrix_timeout"
	ERROR_CIRCUIT_OPEN    = "hystrix_circuit_open"
	ERROR_MAX_CONCURRENCY = "hystrix_max_concurrency"
	ERROR_FLASH_OTHER     = "flash_unknown"
	ERROR_OTHER           = "unknown"
)