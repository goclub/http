package xhttp_test

import (
	xerr "github.com/goclub/error"
	xhttp "github.com/goclub/http"
	"testing"
)

type Mock struct {
	Requests map[string]interface{}
	Replys map[string]interface{}
	Match func(c *xhttp.Context) (key string)
}
func URL(pattern xhttp.Pattern, mock Mock) {

}
type NewsRequest struct {
	NewsID int64 `json:"newsID"`
}
type NewsReply struct {
	Title string `json:"title"`
	Context string `json:"context"`
	xerr.Resp
}
type AuditRequest struct {
	NewsID int64 `json:"newsID"`
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
func MockMatchCount(c *xhttp.Context, routers map[string]string) (key string) {
	return MockMatchSceneCount(c, map[string]map[string]string{
		"": routers,
	})
}
func MockMatchSceneCount(c *xhttp.Context, routers map[string]map[string]string) (key string) {
	scene := c.Request.Header.Get("_scene")
	sceneData ,hasSceneData := routers[scene]
	if hasSceneData == false {
		return
	}
	var hasKey bool
	key, hasKey = sceneData[c.Request.Header.Get("_count")]
	if hasKey == false {
		return
	}
	return
}
func TestMock(t *testing.T) {

	URL(xhttp.Pattern{xhttp.POST, "/"}, Mock{
		Requests: map[string]interface{}{
			"normal": NewsRequest{
				NewsID: 1,
			},
		},
		Replys: map[string]interface{}{
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
	URL(xhttp.Pattern{xhttp.POST, "/audit/status"}, Mock{
		Requests: map[string]interface{}{
			"normal": AuditRequest{
				NewsID: 1,
			},
		},
		Match: func(c *xhttp.Context) (key string) {
			return MockMatchSceneCount(c, map[string]map[string]string{
				"finalDone": {
					"1": "queue",
					"2": "queue",
					"3": "done",
				},
				"finalReject": {
					"1": "queue",
					"2": "queue",
					"3": "reject",
				},
			})
		},
		Replys: map[string]interface{}{
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
}
