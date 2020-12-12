package xhttp_test

import (
	"context"
	xhttp "github.com/goclub/http"
	xjson "github.com/goclub/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestExample(t *testing.T) {
	ExampleClient_DO()
	ExampleClient_Send_Query()
}
func ExampleClient_DO() {
	client := xhttp.NewClient(&http.Client{})
	request, err := http.NewRequest("GET", "https://unpkg.com/goclub@0.0.1/package.json", nil) ; if err != nil {
		panic(err)
	}
	resp, bodyClose, statusCode, err := client.Do(request) ; if err != nil {
		panic(err)
	}
	defer bodyClose()
	if statusCode != 200 {
		panic(errors.New("response " + resp.Status))
	}
	data, err := ioutil.ReadAll(resp.Body) ; if err != nil {
		panic(err)
	}
	log.Print("ExampleClient_DO response string", string(data))
}

type ExampleYahooLocation struct {
	AppID string
	Output string
	Command string
}
func (q ExampleYahooLocation) Query() (url.Values, error) {
	v := url.Values{}
	v.Set("appid", q.AppID)
	v.Set("output", q.Output)
	v.Set("command", q.Command)
	return v, nil
}
func ExampleClient_Send_Query() {
	client := xhttp.NewClient(&http.Client{})
	// http://sugg.us.search.yahoo.net/gossip-gl-location/?appid=weather&output=json&command=%E5%B9%BF
	url := "http://sugg.us.search.yahoo.net/gossip-gl-location/"
	req := ExampleYahooLocation{
		AppID: "weather",
		Output: "json",
		Command: "上海",
	}
	resp, bodyClose, statusCode, err := client.Send(context.TODO(), xhttp.GET, url, xhttp.Request{
		Debug: true,
		Query: req,
	}) ; if err != nil {
		panic(err)
	}
	defer bodyClose()
	if statusCode != 200 {
		panic(errors.New("url:" + url + " status:" + resp.Status))
	}
	reply := struct {
		Gossip struct{
			Gprid string `json:"gprid"`
		} `json:"gossip"`
	}{}
	err = xjson.NewDecoder(resp.Body).Decode(&reply) ; if err != nil {
		panic(err)
	}
	log.Printf("ExampleClient_Send %+v", reply)
}