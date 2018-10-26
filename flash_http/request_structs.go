package flash_http

import (
	"time"
)

const (
	BASIC_AUTH = 1
)

type HTTPRequest struct {
	URL            string
	RequestType    string
	Body           []byte
	Headers        map[string]string
	HystrixCommand string
	Timeout        int
	AuthType       int
	Proxy          string
	AuthData       map[string]string
}

func (r *HTTPRequest) GetHystrixCommand() string{
	if len(r.HystrixCommand) > 0 {
		return r.HystrixCommand
	}
	key := GetHystrixAutoKey(r.URL)
	if len(key) > 0{
		return key
	}
	key = GetHystrixDefaultCommand()
	return key
}

func (r *HTTPRequest) GetTimeOut() time.Duration {
	return time.Duration(r.Timeout)*time.Second
}

func NewHTTPRequest() *HTTPRequest {
	request := HTTPRequest{}
	request.Timeout = 2
	return &request
}

type HTTPResponse struct {
	Body           []byte
	HttpStatus	   int
}