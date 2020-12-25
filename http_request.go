package hclient

import (
	"bytes"
	mp "github.com/m-murad/ordered-sync-map"
	"github.com/sujit-baniya/log"
	"github.com/sujit-baniya/utils"
	"github.com/sujit-baniya/utils/pool"
	"github.com/sujit-baniya/xid"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpRequest struct {
	ID           string
	Url          string
	Payload      interface{}
	Headers      *mp.Map
	RetryMax     int
	RetryWaitMax time.Duration
	Timeout      time.Duration
	ReqPerSec    int
	MaxPoolSize  int
	Response     *http.Response
	HttpError    error
	Status       int
	client       *Client
}

func (w *HttpRequest) Client() *Client {
	if w.client == nil {
		opts := Options{
			RetryWaitMax: w.RetryWaitMax,
			Timeout:      w.Timeout,
			RetryMax:     w.RetryMax,
			MaxPoolSize:  w.MaxPoolSize,
			KillIdleConn: true,
			ReqPerSec:    w.ReqPerSec,
		}
		w.client = NewClient(opts)
	}
	return w.client
}

func (w *HttpRequest) Get(payload interface{}) *HttpRequest {
	var y *HttpRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Get(w.Url, payload, w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HttpRequest) Log(payload interface{}, err error) {
	req, _ := utils.Json.Marshal(payload)
	mu.RLock()

	var e *log.Entry
	response := ""
	statusCode := 20
	if w.Response != nil {
		resp, _ := ioutil.ReadAll(w.Response.Body)
		response = string(resp)
		statusCode = w.Response.StatusCode
		e = log.Info()
	} else {
		e = log.Error()
	}

	if err != nil {
		e = log.Error()
	} else {
		e = log.Info()
	}
	headers := make(map[string]string)
	w.Headers.UnorderedRange(func(key interface{}, value interface{}) {
		headers[key.(string)] = value.(string)
	})
	head, _ := utils.Json.Marshal(headers)
	e.
		Str("request_id", xid.New().String()).
		Str("url", w.Url).
		RawJSON("request_payload", req).
		RawJSON("request_header", head).
		Int("status", statusCode).
		Str("response", response).
		Msg("Client Response")
	mu.RUnlock()
}

func (w *HttpRequest) PostJson(payload interface{}) *HttpRequest {
	req, _ := utils.Json.Marshal(payload)
	var y *HttpRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().PostJson(w.Url, bytes.NewBufferString(string(req)), w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}
	y.Log(payload, nil)
	return y
}

func (w *HttpRequest) GetJson(payload interface{}) *HttpRequest {
	var y *HttpRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().GetJson(w.Url, payload, w.Headers)
	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HttpRequest) Post(payload interface{}) *HttpRequest {
	req, _ := utils.Json.Marshal(payload)
	var y *HttpRequest
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Post(w.Url, bytes.NewBufferString(string(req)), w.Headers)

	y = w
	y.Response = resp
	y.HttpError = err
	if err != nil {
		y.Log(payload, err)
		return y
	}

	y.Log(payload, nil)
	return y
}

func (w *HttpRequest) AsyncGet(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.Get(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncPostJson(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.PostJson(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncGetJson(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}
		response := w.GetJson(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncPost(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		response := w.Post(payload)
		if response.HttpError != nil {
			return nil, response.HttpError
		}
		return w, nil
	}
}
