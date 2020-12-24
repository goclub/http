package xhttp

import (
	"bytes"
	"encoding/base64"
	"github.com/goclub/time"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type URLBase64 struct {
	Value string
}
// 实现 juice.QueryValuer
func (b *URLBase64) MarshalRequest(value string) error {
	valueBytes, err := base64.URLEncoding.DecodeString(value)
	if err != nil {return err}
	b.Value = string(valueBytes)
	return nil
}
func TestBindRequestJSON(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `json:"school"`
	}
	type Job struct {
		Title string `json:"jobTitle"`
	}
	type Req struct {
		Name string `json:"name"`
		Age int `json:"age"`
		Elevation int `json:"elevation"`
		Happy bool `json:"happy"`
		UserID UserID `json:"userID"`
		School
		Job Job
		Time xtime.ChinaTime `json:"time"`
	}
	r := httptest.NewRequest(
		"GET",
		"http://github.com/og/juice",
		strings.NewReader(`{
		"name":"nimoc",
		"age": "27",
		"elevation": "-100",
		"happy": true,
		"userID": "a",
		"school": "xjtu",
		"job":{"jobTitle":"Programmer"},
		"time": "2020-10-01 22:46:53"
		}`),
	)
	r.Header.Set("Content-Type", "application/json")
	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Time: xtime.NewChinaTime(time.Date(2020,10,01,22,46,53,0, xtime.LocationChina)),
	}, req)
}

func TestBindRequestQuery(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `query:"school"`
	}
	type Job struct {
		Title string `query:"jobTitle"`
	}
	type Req struct {
		Name string `query:"name"`
		Age uint `query:"age"`
		Elevation int `query:"elevation"`
		Happy bool `query:"happy"`
		UserID UserID `query:"userID"`
		School
		Job Job
		Website URLBase64 `query:"website"`
	}
	r := httptest.NewRequest(
		"GET",
		"http://github.com/og/juice?"+
			"name=nimoc&"+
			"age=27&"+
			"elevation=-100&"+
			"happy=true&"+
			"userID=a&"+
			"school=xjtu&"+
			"jobTitle=Programmer&"+
			"website=aHR0cHM6Ly9naXRodWIuY29tL25pbW9j",
		nil,
	)
	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Website: URLBase64{Value: "https://github.com/nimoc"},
	}, req)
}
func TestBindRequestWWWForm(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `form:"school"`
	}
	type Job struct {
		Title string `form:"jobTitle"`
	}
	type Req struct {
		Name string `form:"name"`
		Age uint `form:"age"`
		Elevation int `form:"elevation"`
		Happy bool `form:"happy"`
		UserID UserID `form:"userID"`
		School
		Job Job
		Website URLBase64 `form:"website"`
	}

	r := httptest.NewRequest(
		"POST",
		"http://github.com/og/juice",
		strings.NewReader("name=nimoc&"+
			"age=27&"+
			"elevation=-100&"+
			"happy=true&"+
			"userID=a&"+
			"school=xjtu&"+
			"jobTitle=Programmer&"+
			"website=aHR0cHM6Ly9naXRodWIuY29tL25pbW9j"),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Website: URLBase64{Value: "https://github.com/nimoc"},
	}, req)
}
func TestBindRequestFormData(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `form:"school"`
	}
	type Job struct {
		Title string `form:"jobTitle"`
	}
	type Req struct {
		Name string `form:"name"`
		Age uint `form:"age"`
		Elevation int `form:"elevation"`
		Happy bool `form:"happy"`
		UserID UserID `form:"userID"`
		School
		Job Job
		Website URLBase64 `form:"website"`
	}

	var r *http.Request
	{
		values, err := url.ParseQuery("name=nimoc&"+
			"age=27&"+
			"elevation=-100&"+
			"happy=true&"+
			"userID=a&"+
			"school=xjtu&"+
			"jobTitle=Programmer&"+
			"website=aHR0cHM6Ly9naXRodWIuY29tL25pbW9j") ; if err != nil {panic(err)}
		bufferData := bytes.NewBuffer(nil)
		formWriter := multipart.NewWriter(bufferData)
		for key, values := range values {
			err := formWriter.WriteField(key, values[0]); if err != nil {panic(err)}
		}
		err = formWriter.Close() ; if err != nil {panic(err)}
		r = httptest.NewRequest(
			"POST",
			"http://github.com/og/juice",
			bufferData,
		)
		r.Header.Add("Content-Type", formWriter.FormDataContentType())
	}

	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Website: URLBase64{Value: "https://github.com/nimoc"},
	}, req)
}

func TestBindRequestQueryAndWWWForm(t *testing.T) {
	

	type Req struct {
		Name1 string `query:"name1"`
		Age1 uint `query:"age1"`
		Website1 URLBase64 `query:"website1"`
		Name2 string `form:"name2"`
		Age2 uint `form:"age2"`
		Website2 URLBase64 `form:"website2"`
	}

	var r *http.Request
	{

		r = httptest.NewRequest(
			"POST",
			"http://github.com/og/juice" +
				"?name1=nimoc1&"+
				"age1=1&"+
				"website1=aHR0cHM6Ly9naXRodWIuY29tL25pbW9j",
			strings.NewReader(
				"name2=nimoc2&"+
					"age2=2&"+
					"website2=Mg==",),
		)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name1: "nimoc1",
		Age1: 1,
		Website1: URLBase64{Value: "https://github.com/nimoc"},
		Name2: "nimoc2",
		Age2: 2,
		Website2: URLBase64{Value: "2"},
	}, req)
}


func TestBindRequestQueryAndJSON(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `json:"school"`
	}
	type Job struct {
		Title string `json:"jobTitle"`
	}
	type Req struct {
		Name string `json:"name"`
		Age int `json:"age"`
		Elevation int `json:"elevation"`
		Happy bool `json:"happy"`
		UserID UserID `json:"userID"`
		School
		Job Job
		Time xtime.ChinaTime `json:"time"`
		Demo string `query:"demo"`
	}
	r := httptest.NewRequest(
		"GET",
		"http://github.com/og/juice?demo=a",
		strings.NewReader(`{
		"name":"nimoc",
		"age": "27",
		"elevation": "-100",
		"happy": true,
		"userID": "a",
		"school": "xjtu",
		"job":{"jobTitle":"Programmer"},
		"time": "2020-10-01 22:46:53"
		}`),
	)
	r.Header.Set("Content-Type", "application/json")
	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Time: xtime.NewChinaTime(time.Date(2020,10,01,22,46,53,0, xtime.LocationChina)),
		Demo: "a",
	}, req)
}

func check (err error) {
	if err != nil {panic(err)}
}
func TestBindRequestQueryAndFormData(t *testing.T) {
	
	type UserID string
	type School struct {
		School string `form:"school"`
	}
	type Job struct {
		Title string `form:"jobTitle"`
	}
	type Req struct {
		Name string `form:"name"`
		Age uint `form:"age"`
		Elevation int `form:"elevation"`
		Happy bool `form:"happy"`
		UserID UserID `form:"userID"`
		School
		Job Job
		Website URLBase64 `form:"website"`
		Demo string `query:"demo"`
	}

	var r *http.Request
	{
		values, err := url.ParseQuery("name=nimoc&"+
			"age=27&"+
			"elevation=-100&"+
			"happy=true&"+
			"userID=a&"+
			"school=xjtu&"+
			"jobTitle=Programmer&"+
			"website=aHR0cHM6Ly9naXRodWIuY29tL25pbW9j") ; check(err)
		bufferData := bytes.NewBuffer(nil)
		formWriter := multipart.NewWriter(bufferData)
		for key, values := range values {
			check(formWriter.WriteField(key, values[0]))
		}
		check(formWriter.Close())
		r = httptest.NewRequest(
			"POST",
			"http://github.com/og/juice?demo=1",
			bufferData,
		)
		r.Header.Add("Content-Type", formWriter.FormDataContentType())
	}

	req := Req{}
	assert.NoError(t,BindRequest(&req, r))
	assert.Equal(t,Req{
		Name: "nimoc",
		Age: 27,
		Elevation: -100,
		Happy: true,
		UserID: UserID("a"),
		School: School{School: "xjtu"},
		Job: Job{Title: "Programmer"},
		Website: URLBase64{Value: "https://github.com/nimoc"},
		Demo: "1",
	}, req)
}
