package main

import (
	"bytes"
	"github.com/CloudyKit/jet/v6"
	xerr "github.com/goclub/error"
	xhttp "github.com/goclub/http"
	xjson "github.com/goclub/json"
	"net/http"
	"os"
	"path"
	"reflect"
)

type NewsRequest struct {
	NewsID int64 `query:"newsID"`
}
type NewsReply struct {
	Title string `json:"title"`
	Context string `json:"context"`
	xerr.Resp
}
type AuditRequest struct {
	NewsID int64 `query:"newsID"`
}
type AuditStatus int8
const (
	EnumAuditStatusQueue AuditStatus = 1
	EnumAuditStatusDone		  AuditStatus = 2
	EnumAuditStatusReject     AuditStatus = 3
)
type AuditReply struct {
	Status AuditStatus `json:"status"`
	RejectReason string `json:"rejectReason"`
	xerr.Resp
}

var view *jet.Set

func init (){
	loader := jet.NewOSFileSystemLoader(path.Join(os.Getenv("GOPATH"), "src/github.com/goclub/http/example/internal/mock"))
	opts := []jet.Option{}
	opts = append(opts, jet.InDevelopmentMode())
	opts = append(opts, jet.WithDelims("[[", "]]"))
	view = jet.NewSet(
		loader,
		opts...
	)
	view.AddGlobalFunc("xjson", func(a jet.Arguments) reflect.Value {
		v := a.Get(0).Interface()
		buffer := bytes.NewBuffer(nil)
		err := xjson.NewEncoder(buffer).Encode(v) ; if err != nil {
			return reflect.ValueOf("encode json fail")
		}
		return reflect.ValueOf(buffer.Bytes())
	})
}

type TemplateRender struct {

}

func (tr TemplateRender) Render(templatePath string, data interface{}, w http.ResponseWriter) (err error) {
	t, err := view.GetTemplate(templatePath) ; if err != nil {
	    return
	}
	return t.Execute(w, nil, data)
}

func main() {
	ms := xhttp.NewMockServer(xhttp.MockServerOption{
		DefaultReply: map[string]interface{}{
			"pass": xerr.Resp{},
			"fail": xerr.Resp{
				Error: xerr.RespError{
					Code: 1,
					Message: "错误消息",
				},
			},
		},
		Render: TemplateRender{},
	})
	defer ms.Listen(3422)
	ms.URL(xhttp.Mock{
		Route:  xhttp.Route{xhttp.GET, "/news"},
		Request: xhttp.MockRequest{
			"main": NewsRequest{
				NewsID: 1,
			},
		},
		Reply: xhttp.MockReply{
			"pass": NewsReply{
				Title: "goclub/http 发布 mock 功能",
				Context:  "全新版本,全新体验,解放前端",
			},
			"fail": xerr.Resp{
				Error: xerr.RespError{
					Code:  1,
					Message:  "新闻ID错误",
				},
			},
		},
	})
	ms.URL(xhttp.Mock{
		Route:  xhttp.Route{xhttp.GET, "/audit/status"},
		Request: xhttp.MockRequest{
			"main": AuditReply{},
		},
		Match: func(c *xhttp.Context) (key string) {
			return xhttp.MockMatchSceneCount(c, map[string]map[string]string{
				"": {
					"1": "queue",
					"2": "pass",
					"": "pass",
				},
				"finalReject": {
					"1": "queue",
					"2": "reject",
					"": "reject",
				},
			})
		},
		DisableDefaultReply: "pass",
		Reply: xhttp.MockReply{
			"done": AuditReply{
				Status: EnumAuditStatusDone,
			},
			"reject": AuditReply{
				Status: EnumAuditStatusReject,
				RejectReason:  "内容非法",
			},
			"queue": AuditReply{
				Status: EnumAuditStatusQueue,
			},
		},
	})
	ms.URL(xhttp.Mock{
		Route:               xhttp.Route{xhttp.GET, "/handleFunc"},
		HandleFunc: func(c *xhttp.Context, data interface{}) error {
			return c.WriteJSON("handleFunc")
		},
	})
	ms.URL(xhttp.Mock{
		Route:               xhttp.Route{xhttp.GET, "/render"},
		Reply: xhttp.MockReply{
			"pass": map[string]interface{}{
				"name": "nimo",
			},
		},
		Render: "./page.html",
	})

}
