package flash_http

import (
	"net/http"
	"testing"
)

func TestFlashHttp(t *testing.T) {
	request := NewHTTPRequest()
	request.URL = "https://www.google.com"
	res, err := DoFlashHttp(request, 0)
	if err != nil {
		t.Errorf("Flashhttp error %v ", err)
	} else if res == nil {
		t.Errorf("Flashhttp blnk response")
	} else if res.HttpStatus != http.StatusOK {
		t.Errorf("Flashhttp status = %v", res.HttpStatus)
	} else if len(res.Body) == 0 {
		t.Errorf("Flashhttp blank body")
	}

}

func TestFlashHttpSkipLogs(t *testing.T) {
	request := NewHTTPRequest()
	request.URL = "https://www.google.com"
	request.SkipLogs = true
	res, err := DoFlashHttp(request, 0)
	if err != nil {
		t.Errorf("Flashhttp error %v ", err)
	} else if res == nil {
		t.Errorf("Flashhttp blnk response")
	} else if res.HttpStatus != http.StatusOK {
		t.Errorf("Flashhttp status = %v", res.HttpStatus)
	} else if len(res.Body) == 0 {
		t.Errorf("Flashhttp blank body")
	}
}

func TestFlashHttpErrorHandled(t *testing.T) {
	request := NewHTTPRequest()
	request.URL = "https://www.google.unknown.com"
	res, err := DoFlashHttp(request, 0)
	if err == nil {
		t.Errorf("Flashhttp no error ")
	} else if res.HttpStatus == http.StatusOK {
		t.Errorf("Flashhttp status = %v", res.HttpStatus)
	}
}

func TestFlashHttpWithReDirect(t *testing.T) {
	request := NewHTTPRequest()
	request.URL = "https://www.google.com"
	res, err := DoFlashHttp(request, 1)
	if err != nil {
		t.Errorf("Flashhttp error %v ", err)
	} else if res == nil {
		t.Errorf("Flashhttp blnk response")
	} else if res.HttpStatus != http.StatusOK {
		t.Errorf("Flashhttp status = %v", res.HttpStatus)
	} else if len(res.Body) == 0 {
		t.Errorf("Flashhttp blank body")
	}

}

