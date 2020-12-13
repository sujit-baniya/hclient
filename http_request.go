package hclient

import (
	mp "github.com/m-murad/ordered-sync-map"
	"github.com/sujit-baniya/log"
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

func (w *HttpRequest) Get(payload interface{}) error {
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Get(w.Url, payload, w.Headers)
	if err != nil {
		return err
	}
	w.Response = resp
	w.Log(payload, err)
	return err
}

func (w *HttpRequest) Log(payload interface{}, err error) {
	mu.RLock()

	var e *log.Entry
	resp, _ := ioutil.ReadAll(w.Response.Body)
	if err != nil {
		e = log.Error()
	} else {
		e = log.Info()
	}
	headers := make(map[string]string)
	w.Headers.UnorderedRange(func(key interface{}, value interface{}) {
		headers[key.(string)] = value.(string)
	})
	e.
		Str("request_id", xid.New().String()).
		Str("url", w.Url).
		Interface("request_payload", payload).
		Interface("request_header", headers).
		Int("status", w.Response.StatusCode).
		Str("response", string(resp)).
		Msg("Client Response")
	mu.RUnlock()
}

func (w *HttpRequest) PostJson(payload interface{}) error {
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().PostJson(w.Url, payload, w.Headers)
	if err != nil {
		return err
	}
	w.Response = resp
	return err
}

func (w *HttpRequest) GetJson(payload interface{}) error {
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().GetJson(w.Url, payload, w.Headers)
	if err != nil {
		return err
	}
	w.Response = resp
	w.Log(payload, err)
	return err
}

func (w *HttpRequest) Post(payload interface{}) error {
	if w.Headers == nil {
		w.Headers = mp.New()
	}
	resp, err := w.Client().Post(w.Url, payload, w.Headers)
	if err != nil {
		return err
	}
	w.Response = resp
	w.Log(payload, err)
	return err
}

func (w *HttpRequest) AsyncGet(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		err := w.Get(payload)
		if err != nil {
			panic(err)
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncPostJson(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		err := w.PostJson(payload)
		if err != nil {
			panic(err)
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncGetJson(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}
		err := w.GetJson(payload)
		if err != nil {
			panic(err)
		}
		return w, nil
	}
}

func (w *HttpRequest) AsyncPost(payload interface{}) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		err := w.Post(payload)
		if err != nil {
			panic(err)
		}
		return w, nil
	}
}
