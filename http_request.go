package flash_http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/valyala/fasthttp"
)

const FLASH_HTTP = "flash_http"


type LOG func(errType string, args ...interface{})
var log LOG = func(errType string, args ...interface{}) {
	if errType !=  "" {
		fmt.Println("Error", args)
	} else {
		fmt.Println("Info", args)
	}
}

func SetLogger(logger LOG){
	log = logger
}

func logData(skipLog bool, isError bool, args ...interface{}){
	if skipLog {
		return
	}
	errType := ""
	if isError {
		errType = ERROR_OTHER
	}
	log(errType, args)
}


func (request *HTTPRequest) prepareFastHttpRequest() *fasthttp.Request {
	httpRequest := fasthttp.AcquireRequest()
	httpHeaders := &httpRequest.Header

	httpRequest.SetRequestURI(request.URL)
	httpRequest.Header.SetMethod(request.RequestType)
	if request.RequestType != http.MethodGet{
		httpRequest.SetBody(request.Body)
	}
	if httpRequest != nil {
		for k, v := range request.Headers {
			httpHeaders.Set(k, v)
		}
		if request.AuthType == BASIC_AUTH {
			username, _ := request.AuthData[AUTH_KEY_USERNAME]
			password, _ := request.AuthData[AUTH_KEY_PASSWORD]
			token := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
			httpHeaders.Set("Authorization", token)
		}
	}
	logData(request.GetSkipLogs(), false,"httpHeaders ", httpHeaders.String(), "  ")
	return httpRequest
}

func doClient(httpRequest *fasthttp.Request, httpResponse *fasthttp.Response, proxy string, timeout time.Duration) error {
	defer func() {
		logData(false,false,"Response headers ", httpResponse.Header.String(), "End")
	}()
	if proxy == "" {
		client := &fasthttp.Client{}
		return client.DoTimeout(httpRequest, httpResponse, timeout)
	} else {
		logData(false, false, "proxy", proxy)
		client := &fasthttp.HostClient{
			Addr: proxy,
		}
		return client.DoTimeout(httpRequest, httpResponse, timeout)
	}
}

func getBodyBytes(httpRequest *fasthttp.Request, httpResponse *fasthttp.Response, skipLogs bool) (res []byte) {
	logData(skipLogs, false, "Response Headers ", string(httpResponse.Header.Header()))
	if httpRequest.Header.HasAcceptEncoding("gzip") {
		res, _ = httpResponse.BodyGunzip()
	} else {
		res = httpResponse.Body()
	}
	return res
}

type MetricUpdateFuncType func(urlStr string, startTime time.Time, responseStatus int, args ...interface{}) error
var metricUpdateFunction MetricUpdateFuncType = func(urlStr string, startTime time.Time, responseStatus int, args ...interface{}) error {
	return nil
}

func SetMetricUpdateFunc(metricUpdateFunc MetricUpdateFuncType) {
	metricUpdateFunction = metricUpdateFunc
}

func updateMetrics(urlStr string, startTime time.Time, responseStatus int, args ...interface{}) error{
	return metricUpdateFunction(urlStr, startTime, responseStatus, args)
}

func handleFlashError(hystrixKey string, url string, body string, err error, skipError bool){
	if skipError {
		return
	}

	if log != nil {
		errMsg := err.Error()
		errType := ERROR_FLASH_OTHER
		if strings.Contains(errMsg,"circuit") {
			errType = ERROR_CIRCUIT_OPEN
		} else if strings.Contains(errMsg,"max concurrency"){
			errType = ERROR_MAX_CONCURRENCY
		} else if strings.Contains(errMsg,"timeout"){
			errType = ERROR_TIMEOUT
		}
		log(errType, fmt.Sprintf("hystrixKey=%s, URL=%s, Error=%s, Body=%s", hystrixKey, url, errMsg, body))
	}
}

// Common http service for all external calls, in sync
func DoFlashHttp(request *HTTPRequest) (responseObject *HTTPResponse, err error) {
	var responseData []byte
	startTime := time.Now()
	logData(request.GetSkipLogs(),false,"FLASH HTTP REQUEST BODY, ", string(request.Body))
	httpRequest := request.prepareFastHttpRequest()

	responseObject = &HTTPResponse{}
	if httpRequest != nil {
		httpResponse := fasthttp.AcquireResponse()
		proxy := request.Proxy
		var respData []byte
		hystrixKey := request.GetHystrixCommand()
		if hystrixKey != "" {
			logData(request.GetSkipLogs(),false,"Hystrix Command=", hystrixKey)
			err = hystrix.Do(hystrixKey, func() error {
				logData(request.GetSkipLogs(),false,"Hystrix hit -> ", request.URL)
				err = doClient(httpRequest, httpResponse, proxy, request.GetTimeOut())
				if err != nil {
					logData(request.GetSkipLogs(),true,"hystrix.Do error1 ", err)
					//handleFlashError(hystrixKey, request.URL, err, request.SkipErrorHandler)
					responseObject.HttpStatus = http.StatusGatewayTimeout
					return err
				}
				respData = getBodyBytes(httpRequest, httpResponse, request.GetSkipLogs())
				if httpResponse.StatusCode() >= http.StatusInternalServerError {
					return errors.New(fmt.Sprintf("HttpStatus code == %d", httpResponse.StatusCode()))
				}
				return nil
			}, func(e error) error {
				logData(request.GetSkipLogs(),true,"hystrix.Do error2", e, request.URL)
				handleFlashError(hystrixKey, request.URL, string(respData), e, request.GetSkipLogs())
				return e
			})
			if err != nil {
				logData(request.GetSkipLogs(),true,"hystrix.Do init error", err)
			}
		} else {
			logData(request.GetSkipLogs(),false,"Non-Hystrix hit -> ", request.URL)
			err = doClient(httpRequest, httpResponse, proxy, request.GetTimeOut())
			if err != nil {
				responseObject.HttpStatus = http.StatusGatewayTimeout
				logData(request.GetSkipLogs(),true,"client.Do error ", err)
			}
			respData = getBodyBytes(httpRequest, httpResponse, request.GetSkipLogs())
		}
		if responseObject.HttpStatus == 0 {
			responseObject.HttpStatus = httpResponse.StatusCode()
		}
		responseData = respData
	}

	responseObject.Body = responseData
	logData(request.GetSkipLogs(),false, "FLASH HTTP RESPONSE BODY, ", string(responseData))
	go updateMetrics(request.URL, startTime, responseObject.HttpStatus)
	return responseObject, err
}

//// GoFlashHttp Common http service to communicate with external calls in async with goroutines. Use go prefix.
//func GoFlashHttp(subResponseChannel chan []byte, request *HTTPRequest) []byte {
//	responseData := DoFlashHttp(request)
//	fmt.Println("FLASH HTTP REQUEST BODY, ", string(responseData))
//	if subResponseChannel != nil {
//		subResponseChannel <- responseData
//	} else {
//		return responseData
//	}
//	return nil
//}
