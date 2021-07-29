package xhttp

import (
	"fmt"
	xerr "github.com/goclub/error"
	xjson "github.com/goclub/json"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
)

type MockServer struct {
	router *Router
	db mockDatabase
	option MockServerOption
}
type mockDatabase struct {
	replyKey map[string]string
	scene string
}
type MockServerOption struct {
	DefaultReply map[string]interface{}
	RequestCheck func (c *Context, pattern Pattern, reqPtr interface{}) (pass bool, err error)
}
func NewMockServer(option MockServerOption) MockServer {
	server := MockServer{
		option: option,
		db: mockDatabase{
			replyKey: map[string]string{},
		},
		router: NewRouter(RouterOption{}),
	}
	server.systemHandle()
	return server
}
func (m MockServer) systemHandle() {
	m.router.Use(func(c *Context, next Next) (err error) {
		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		return next()
	})
	m.router.HandleFunc(Pattern{GET, "/mock"}, func(c *Context) (err error) {
		req := struct {
			Scene string `query:"scene"`
		}{}
		reply := struct {
			Action string `json:"action"`
			xerr.Resp
		}{}
		err = c.BindRequest(&req) ; if err != nil {
		    return
		}
		if req.Scene != "" {
			m.db.scene = req.Scene
			reply.Action = `scene was successfully set to `+ req.Scene
		}
		return c.WriteJSON(reply)
	})
}
type Mock struct {
	Pattern Pattern
	Request func() interface{}
	Reply map[string]interface{}
	Match func(c *Context) (key string)
}
func (ms MockServer) URL(mock Mock) {
	reply:= map[string]interface{}{}
	for replyKey, replyValue := range ms.option.DefaultReply {
		reply[replyKey] = replyValue
	}
	for replyKey, replyValue := range mock.Reply {
		reply[replyKey] = replyValue
	}
	ms.router.HandleFunc(mock.Pattern, func(c *Context) (err error) {
		if ms.option.RequestCheck != nil && mock.Request != nil {
			var pass bool
			pass, err = ms.option.RequestCheck(c, mock.Pattern, mock.Request())  ; if err != nil {
			    return
			}
			if pass == false {
				return
			}
		}
		var replyKey string

		replyKeyValues :=  reflect.ValueOf(reply).MapKeys()
		var replyKeyStrings []string
		for _, rValue := range replyKeyValues {
			replyKeyStrings = append(replyKeyStrings, rValue.String())
		}
		sort.Strings(replyKeyStrings)
		if len(replyKeyStrings) == 0 {
			c.WriteStatusCode(500)
			return c.WriteBytes([]byte(fmt.Sprintf("When xhttp.NewMock(option) option.DefaultReply is empty map,  MockServer{}.URL(mock) mock.Reply  can not be empty map. mock is %#+v", mock)))
		}
		replyKey = replyKeyStrings[0]
		for _, key := range replyKeyStrings {
			// 优先响应 pass
			if key == "pass" {
				replyKey = key
			}
		}
		currentReplyKey := ms.currentReplyKey(c, mock.Pattern, mock.Match)
		if currentReplyKey != "" {
			replyKey = currentReplyKey
		}
		response, hasResponse := reply[replyKey]
		if hasResponse == false {
			c.WriteStatusCode(500)
			replyBytes, err := xjson.MarshalIndent(mock.Reply, "", "  ") ; if err != nil {
				replyBytes = []byte(fmt.Sprintf("%+v", mock.Reply))
			}
			return c.WriteBytes([]byte(fmt.Sprintf("reply:%s\ncan not found key: %s", replyBytes, replyKey)))
		}
		return c.WriteJSON(response)
	})
}

func (server MockServer) currentReplyKey(c *Context, pattern Pattern, match func(*Context) (string)) (replyKey string){
	// database
	dbReplyKey,hasDBReplyKey := server.db.replyKey[pattern.mockID()]
	if hasDBReplyKey {
		replyKey = 	dbReplyKey
	}
	// header
	headerReplyKey := c.Request.URL.Query().Get("_reply")
	if headerReplyKey != "" {
		replyKey = 	headerReplyKey
	}
	// match
	headerScene :=  c.Request.URL.Query().Get("_scene")
	if headerScene == "" {
		c.Request.Header.Set("_scene", server.db.scene)
	}
	if match != nil {
		matchReplyKey := match(c)
		if matchReplyKey != "" {
			replyKey = 	matchReplyKey
		}
	}
	return
}

func MockMatchCount(c *Context, routers map[string]string) (key string) {
	return MockMatchSceneCount(c, map[string]map[string]string{
		"": routers,
	})
}
func MockMatchSceneCount(c *Context, routers map[string]map[string]string) (key string) {
	scene := c.Request.URL.Query().Get("_scene")
	sceneData ,hasSceneData := routers[scene]
	if hasSceneData == false {
		return
	}
	var hasKey bool
	key, hasKey = sceneData[c.Request.URL.Query().Get("_count")]
	if hasKey == false {
		return
	}
	return
}
func (server MockServer) Listen(port int) {
	s := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: server.router,
	}
	server.router.LogPatterns(s)
	log.Print(s.ListenAndServe())
}