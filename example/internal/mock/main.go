package main

import (
	xerr "github.com/goclub/error"
	xhttp "github.com/goclub/http"
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

func main() {
	ms := xhttp.NewMockServer(xhttp.MockServerOption{
		RequestCheck: func(c *xhttp.Context, pattern xhttp.Pattern, reqPtr interface{}) (pass bool, err error) {
			if pattern.Equal(xhttp.Pattern{xhttp.GET, "/news"}) {
				err = c.BindRequest(reqPtr) ; if err != nil {
				    return
				}
				if newsRequest, ok := reqPtr.(*NewsRequest); ok {
					if newsRequest.NewsID == 0 {
						return false, c.WriteJSON(xerr.Resp{
							Error: xerr.RespError{
								Code:  1,
								Message: "missing query newsID",
							},
						})
					}
				}
				return true ,nil
			}
			return true, nil
		},
		DefaultReply: map[string]interface{}{
			"pass": xerr.Resp{},
			"fail": xerr.Resp{
				Error: xerr.RespError{
					Code: 1,
					Message: "错误消息",
				},
			},
		},
	})
	defer ms.Listen(3422)
	ms.URL(xhttp.Mock{
		Pattern:  xhttp.Pattern{xhttp.GET, "/news"},
		Request: func() interface{} {
			return &NewsRequest{}
		},
		Reply: map[string]interface{}{
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
		Pattern:  xhttp.Pattern{xhttp.GET, "/audit/status"},
		Request: func() interface{} {
			return &AuditRequest{}
		},
		Match: func(c *xhttp.Context) (key string) {
			return xhttp.MockMatchSceneCount(c, map[string]map[string]string{
				"finalDone": {
					"1": "queue",
					"2": "pass",
				},
				"finalReject": {
					"1": "queue",
					"2": "reject",
				},
			})
		},
		Reply: map[string]interface{}{
			"pass": AuditReply{
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
