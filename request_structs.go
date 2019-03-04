package flash_http

import (
	"time"
	"sync"
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
	Timeout 	   int // depreciate
	TimeoutInMs    int // timeout in milli second
	AuthType       int
	Proxy          string
	SkipErrorHandler bool
	SkipLogs		 bool
	AuthData       map[string]string
}

var defaultTimeOutInMs = 2000
var muxDefaultTimeOutInMs sync.Mutex

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

func (r *HTTPRequest) GetSkipLogs() bool {
	return r.SkipLogs
}

func SetDefaultTimeOut(timeout int ){
	muxDefaultTimeOutInMs.Lock()
	defer muxDefaultTimeOutInMs.Unlock()
	defaultTimeOutInMs = timeout
}

func NewHTTPRequest() *HTTPRequest {
	request := HTTPRequest{}
	request.TimeoutInMs = defaultTimeOutInMs
	return &request
}

type HTTPResponse struct {
	Body           []byte
	HttpStatus	   int
}