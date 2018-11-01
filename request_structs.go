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
	Timeout 	   int
	TimeoutInMs    int // timeout in milli second
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
	if r.Timeout != 0{
		return time.Duration(r.Timeout)*time.Second
	} else {
		return time.Duration(r.TimeoutInMs)*time.Millisecond
	}
}

func NewHTTPRequest() *HTTPRequest {
	request := HTTPRequest{}
	request.TimeoutInMs = 2*1000
	return &request
}

type HTTPResponse struct {
	Body           []byte
	HttpStatus	   int
}