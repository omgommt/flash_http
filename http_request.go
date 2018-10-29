package flash_http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/valyala/fasthttp"
)

const FLASH_HTTP = "flash_http"


type LOG func(isError bool, args ...interface{})
var log LOG = func(isError bool, args ...interface{}) {
	if isError{
		fmt.Sprint("Error", args)
	} else {
		fmt.Sprint("Info", args)
	}
}

func SetLogger(logger LOG){
	log = logger
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
	log(false,"httpHeaders ", httpHeaders.String(), "  ")
	return httpRequest
}

func doClient(httpRequest *fasthttp.Request, httpResponse *fasthttp.Response, proxy string, timeout time.Duration) error {
	defer func() {
		log(false,"Response headers ", httpResponse.Header.String(), "End")
	}()
	if proxy == "" {
		client := &fasthttp.Client{}
		return client.DoTimeout(httpRequest, httpResponse, timeout)
	} else {
		log(false, "proxy", proxy)
		client := &fasthttp.HostClient{
			Addr: proxy,
		}
		return client.DoTimeout(httpRequest, httpResponse, timeout)
	}
}

func getBodyBytes(httpRequest *fasthttp.Request, httpResponse *fasthttp.Response) (res []byte) {
	log(false, "Response Headers ", string(httpResponse.Header.Header()))
	if httpRequest.Header.HasAcceptEncoding("gzip") {
		res, _ = httpResponse.BodyGunzip()
	} else {
		res = httpResponse.Body()
	}
	return res
}

type MetricUpdateFuncType func(urlStr string, startTime time.Time, responseStatus int, args ...interface{}) error
var metricUpdateFunction MetricUpdateFuncType

func SetMetricUpdateFunc(metricUpdateFunc MetricUpdateFuncType) {
	metricUpdateFunction = metricUpdateFunc
}

func updateMetrics(urlStr string, startTime time.Time, responseStatus int, args ...interface{}) error{
	return metricUpdateFunction(urlStr, startTime, responseStatus, args)
}

// Common http service for all external calls, in sync
func DoFlashHttp(request *HTTPRequest) (responseObject *HTTPResponse, err error) {
	var responseData []byte
	startTime := time.Now()
	log(false,"FLASH HTTP REQUEST BODY, ", string(request.Body))
	httpRequest := request.prepareFastHttpRequest()

	responseObject = &HTTPResponse{}
	if httpRequest != nil {
		httpResponse := fasthttp.AcquireResponse()
		proxy := request.Proxy
		var respData []byte
		hystrixKey := request.GetHystrixCommand()
		if hystrixKey != "" {
			log(false,"Hystrix Command=", hystrixKey)
			err = hystrix.Do(hystrixKey, func() error {
				log(false,"Hystrix hit -> ", request.URL)
				err = doClient(httpRequest, httpResponse, proxy, request.GetTimeOut())
				if err != nil {
					log(false,"hystrix.Do error1 ", err)
					return err
				}
				if httpResponse.StatusCode() >= http.StatusInternalServerError {
					return errors.New(fmt.Sprintf("HttpStatus code == %d", httpResponse.StatusCode()))
				}
				respData = getBodyBytes(httpRequest, httpResponse)
				return nil
			}, func(e error) error {
				respData = nil
				err = e
				log(false,"hystrix.Do error2", e, request.URL)
				return nil
			})
			if err != nil {
				log(false,"hystrix.Do init error", err)
			}
		} else {
			log(false,"Non-Hystrix hit -> ", request.URL)
			err = doClient(httpRequest, httpResponse, proxy, request.GetTimeOut())
			if err != nil {
				responseObject.HttpStatus = http.StatusGatewayTimeout
				log(false,"client.Do error ", err)
			}
			respData = getBodyBytes(httpRequest, httpResponse)
		}
		if responseObject.HttpStatus == 0 {
			responseObject.HttpStatus = httpResponse.StatusCode()
		}
		responseData = respData
	}

	responseObject.Body = responseData
	log(false, "FLASH HTTP RESPONSE BODY, ", string(responseData))
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
