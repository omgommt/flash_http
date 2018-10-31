package flash_http

import (
	"github.com/afex/hystrix-go/hystrix"
	"net/url"
	"sync"
)

const (
	//Timeout is how long to wait for command to complete, in milliseconds
	HYSTRIX_TIMEOUT_KEY = "timeout"
	//MaxConcurrent is how many commands of the same type can run at the same time
	HYSTRIX_CONCURRENCY_KEY = "MaxConcurrentRequests"
	//VolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	HYSTRIX_MIN_REQUESTS_KEY = "RequestVolumeThreshold"
	//ErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	HYSTRIX_ERROR_PERCENT_KEY = "ErrorPercentThreshold"
	//SleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	HYSTRIX_RECHECK_TIME_KEY = "SleepWindow"

	HYSTRIX_TIMEOUT_DEFAULT       = 2 * 1000
	HYSTRIX_CONCURRENCY_DEFAULT   = 200
	HYSTRIX_MIN_REQUESTS_DEFAULT  = 100
	HYSTRIX_ERROR_PERCENT_DEFAULT = 33
	HYSTRIX_RECHECK_TIME_DEFAULT  = 5 * 60000
)

var muxEnableAutoHystrix sync.Mutex
var enableAutoHystrix bool
var autoHystrixConfig map[string]int

const hystrixDefault = "hystrixDefaultCommand"

var muxHystrixCommandMap sync.Mutex
var hystrixCommandMap = map[string]bool

// SetHystixCommand
func SetHystixCommand(commandName string, configDetail map[string]int) {
	hysConfig := hystrix.CommandConfig{
		Timeout:                HYSTRIX_TIMEOUT_DEFAULT,
		MaxConcurrentRequests:  HYSTRIX_CONCURRENCY_DEFAULT,
		RequestVolumeThreshold: HYSTRIX_MIN_REQUESTS_DEFAULT,
		ErrorPercentThreshold:  HYSTRIX_ERROR_PERCENT_DEFAULT,
		SleepWindow:            HYSTRIX_RECHECK_TIME_DEFAULT,
	}
	if hTimeout, ok := configDetail[HYSTRIX_TIMEOUT_KEY]; ok {
		hysConfig.Timeout = hTimeout
	}
	if hConcurrency, ok := configDetail[HYSTRIX_CONCURRENCY_KEY]; ok {
		hysConfig.MaxConcurrentRequests = hConcurrency
	}
	if minRequests, ok := configDetail[HYSTRIX_MIN_REQUESTS_KEY]; ok {
		hysConfig.RequestVolumeThreshold = minRequests
	}
	if errorPercent, ok := configDetail[HYSTRIX_ERROR_PERCENT_KEY]; ok {
		hysConfig.ErrorPercentThreshold = errorPercent
	}
	if recheckTime, ok := configDetail[HYSTRIX_RECHECK_TIME_KEY]; ok {
		hysConfig.SleepWindow = recheckTime
	}

	muxHystrixCommandMap.Lock()
	defer muxHystrixCommandMap.Unlock()
	if _, ok := hystrixCommandMap[commandName]; !ok{
		hystrix.ConfigureCommand(commandName, hysConfig)
		hystrixCommandMap[commandName] = true
		log(false,"New SetHystixCommand=", commandName)
	}
}

func SetHystrixDefaultCommandConfig(configDetail map[string]int)  {
	SetHystixCommand(hystrixDefault, configDetail)
}

func EnableAutoHystrixDefault(enable bool, configDetail map[string]int)  {
	muxEnableAutoHystrix.Lock()
	defer muxEnableAutoHystrix.Unlock()
	enableAutoHystrix = enable
	autoHystrixConfig = configDetail
}

func GetHystrixDefaultCommand() string{
	return hystrixDefault
}

func GetHystrixAutoKey(urlPath string) string {
	if !enableAutoHystrix{
		return ""
	}
	u, err := url.Parse(urlPath)
	if err != nil {
		return ""
	}
	key := u.Hostname()+ ":" + u.Port()
	if _, ok := hystrixCommandMap[key]; !ok{
		SetHystixCommand(key, autoHystrixConfig)
	}
	return key
}

