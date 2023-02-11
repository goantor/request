package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goantor/x"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	GetMethod  MethodType = "GET"
	PostMethod MethodType = "POST"

	FormType ContentType = "application/x-www-form-urlencoded"
	JsonType ContentType = "application/json"
)

type ContentType string

type MethodType string

type Request struct {
	Method      MethodType
	ContentType ContentType
	Url         string
	Params      x.H
	Header      http.Header
	Timeout     time.Duration
}

func NewRequest(method MethodType, contentType ContentType, url string, params x.H, header http.Header, timeout time.Duration) *Request {
	if method == GetMethod {
		url = getRequestURL(url, params)
		params = nil
	}

	return &Request{Method: method, ContentType: contentType, Url: url, Params: params, Header: header, Timeout: timeout}
}

func DoRequest(req *Request) (*http.Response, error) {
	return do(req.Method, req.ContentType, req.Url, req.Params, req.Header, req.Timeout)
}

func Auto(method MethodType, contentType ContentType, url string, params x.H, header http.Header, duration time.Duration) (*http.Response, error) {
	if method == GetMethod {
		return Get(url, params)
	}

	if contentType == FormType {
		return Form(url, params, header, duration)
	}

	return Json(url, params, header, duration)
}

func Get(url string, params x.H) (*http.Response, error) {
	client := http.Client{}
	return client.Get(getRequestURL(url, params))
}

// getRequestURL 获取Get 请求
func getRequestURL(url string, params x.H) string {
	queryString := queryParams(params, "")
	return fmt.Sprintf("%s?%s", url, queryString)
}

func Form(url string, params x.H, header http.Header, duration time.Duration) (*http.Response, error) {
	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Type", string(FormType))
	return do(PostMethod, FormType, url, params, header, duration)
}

func Json(url string, params x.H, header http.Header, duration time.Duration) (*http.Response, error) {
	if header == nil {
		header = http.Header{}
	}

	header.Set("Content-Type", "application/json;charset=utf-8")
	return do(PostMethod, JsonType, url, params, header, duration)
}

func do(method MethodType, contentType ContentType, url string, params x.H, header http.Header, duration time.Duration) (resp *http.Response, err error) {
	req, err := makeRequest(method, contentType, url, params)
	if err != nil {
		return
	}

	req.Header = header
	client := http.Client{
		Timeout: duration,
	}

	return client.Do(req)
}

func makeRequest(method MethodType, typ ContentType, url string, params x.H) (*http.Request, error) {
	return http.NewRequest(string(method), url, getData(typ, params))
}

func getData(typ ContentType, params x.H) io.Reader {
	if typ == JsonType {
		js, _ := json.Marshal(params)
		return bytes.NewReader(js)
	}

	return strings.NewReader(queryParams(params, ""))
}

func queryParams(params x.H, format string) string {
	values := url.Values{}
	var nk, ret string
	for k, v := range params {
		if len(format) != 0 {
			nk = fmt.Sprintf(format, k)
		} else {
			nk = k
		}

		switch v.(type) {
		case string:
			values.Add(nk, v.(string))
			break
		case []byte:
			values.Add(nk, string(v.([]byte)))
			break
		case map[string]interface{}:
			ret += queryParams(v.(map[string]interface{}), nk+"[%s]")
			ret += "&"
		case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
			values.Add(nk, fmt.Sprintf("%d", v))
		}
	}

	ret += values.Encode()
	return ret
}
