package main

import (
	"fmt"
	orderedsyncmap "github.com/m-murad/ordered-sync-map"
	"github.com/sujit-baniya/hclient"
	"io/ioutil"
	"time"
)

func main() {
	mp := orderedsyncmap.New()
	request := &hclient.HttpRequest{
		Url:       "http://116.203.188.34/send",
		Headers:   mp,
		Timeout:   10 * time.Second,
		ReqPerSec: 1000,
	}
	body := map[string]string{
		"user_id": "123",
	}
	request.GetJson(body)
	resp, _ := ioutil.ReadAll(request.Response.Body)
	fmt.Println(string(resp))
}
