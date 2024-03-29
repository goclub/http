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
	"strings"
	"sync"
)

type MockServer struct {
	router *Router
	db     mockDatabase
	option MockServerOption
}
type mockDatabase struct {
	sync.Mutex
	count    map[string]int64
	replyKey map[string]string
	scene    string
}
type MockServerOption struct {
	DefaultReply map[string]interface{}
	OnlineMock   string
	Render       MockRender
}
type MockRender interface {
	Render(templatePath string, data interface{}, w http.ResponseWriter) error
}

func NewMockServer(option MockServerOption) MockServer {
	server := MockServer{
		option: option,
		db: mockDatabase{
			count:    map[string]int64{},
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
	m.router.HandleFunc(Route{GET, "/mock"}, func(c *Context) (err error) {
		req := struct {
			Scene string `query:"scene"`
		}{}
		reply := struct {
			Action string `json:"action"`
			xerr.Resp
		}{}
		err = c.BindRequest(&req)
		if err != nil {
			return
		}
		if req.Scene != "" {
			m.db.scene = req.Scene
			reply.Action = `scene was successfully set to ` + req.Scene
		}
		return c.WriteJSON(reply)
	})
}

type Mock struct {
	Route               Route                              `note:"路由"`
	Request             MockRequest                        `note:"请求"`
	DisableDefaultReply string                             `note:"指定禁用默认响应的key"`
	Reply               MockReply                          `note:"响应"`
	Match               func(c *Context) (replyKey string) `note:"根据请求参数决定响应结果"`
	MaxAutoCount        int64                              `note:"最大计数,默认5"`
	HandleFunc          func(c *Context, replyKey string, data interface{}) (err error)
	Render              string
}
type MockRequest map[string]interface{}
type MockReply map[string]interface{}

func (ms MockServer) URL(mock Mock) {
	if mock.MaxAutoCount == 0 {
		mock.MaxAutoCount = 5
	}
	reply := map[string]interface{}{}
	for replyKey, replyValue := range ms.option.DefaultReply {
		if mock.DisableDefaultReply != "" {
			for _, disableDefaultReply := range strings.Split(mock.DisableDefaultReply, "|") {
				if replyKey == disableDefaultReply {
					continue
				}
			}
		}
		reply[replyKey] = replyValue
	}
	for replyKey, replyValue := range mock.Reply {
		reply[replyKey] = replyValue
	}
	ms.router.HandleFunc(mock.Route, func(c *Context) (err error) {
		// _count
		query := c.Request.URL.Query()
		queryCount := query.Get("_count")
		if queryCount == "" {
			ms.db.Lock()
			countKey := mock.Route.ID() + " " + query.Get("_scene")
			dbCount := ms.db.count[countKey]
			dbCount++
			if dbCount > mock.MaxAutoCount {
				dbCount = 1
			}
			ms.db.count[countKey] = dbCount
			query.Set("_count", strconv.FormatInt(dbCount, 10))
			c.Request.URL.RawQuery = query.Encode()
			ms.db.Unlock()
		}

		var replyKey string

		replyKeyValues := reflect.ValueOf(reply).MapKeys()
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
		currentReplyKey := ms.currentReplyKey(c, mock.Route, mock.Match)
		if currentReplyKey != "" {
			replyKey = currentReplyKey
		}
		response, hasResponse := reply[replyKey]
		if hasResponse == false {
			c.WriteStatusCode(500)
			replyBytes, err := xjson.MarshalIndent(mock.Reply, "", "  ")
			if err != nil {
				replyBytes = []byte(fmt.Sprintf("%+v", mock.Reply))
			}
			return c.WriteBytes([]byte(fmt.Sprintf("reply:%s\ncan not found key: %s", replyBytes, replyKey)))
		}
		if mock.HandleFunc != nil {
			return mock.HandleFunc(c, replyKey, response)

		}
		if mock.Render != "" {
			return ms.option.Render.Render(mock.Render, response, c.Writer)
		}
		return c.WriteJSON(response)
	})
}

func (server MockServer) currentReplyKey(c *Context, route Route, match func(*Context) string) (replyKey string) {
	// database
	dbReplyKey, hasDBReplyKey := server.db.replyKey[route.ID()]
	if hasDBReplyKey {
		replyKey = dbReplyKey
	}
	// query
	queryReplyKey := c.Request.URL.Query().Get("_")
	if queryReplyKey != "" {
		replyKey = queryReplyKey
	}
	// match
	queryScene := c.Request.URL.Query().Get("_scene")
	if queryScene == "" {
		c.Request.Header.Set("_scene", server.db.scene)
	}
	if match != nil {
		matchReplyKey := match(c)
		if matchReplyKey != "" {
			replyKey = matchReplyKey
		}
	}
	return
}

func MockMatchCount(c *Context, counts map[string]string) (key string) {
	return MockMatchSceneCount(c, map[string]map[string]string{
		"": counts,
	})
}
func MockMatchSceneCount(c *Context, routers map[string]map[string]string) (replyKey string) {
	count := c.Request.URL.Query().Get("_count")
	scene := c.Request.URL.Query().Get("_scene")
	defer func() {
		log.Printf("MockMatch: _scene(%s) _count(%s) replyKey(%s)", scene, count, replyKey)
	}()
	sceneData, hasSceneData := routers[scene]
	if hasSceneData == false {
		return ""
	}
	var hasKey bool
	replyKey, hasKey = sceneData[count]
	defaultReplyKey, hasDefaultReplyKey := sceneData[""]
	if hasKey == false {
		if hasDefaultReplyKey {
			return defaultReplyKey
		} else {
			return ""
		}
	}
	return replyKey
}
func (server MockServer) Handle(path string, handler http.Handler) {
	server.router.Handle(path, handler)
}
func (server MockServer) PrefixHandle(prefix string, handler http.Handler) {
	server.router.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, handler))
}
func (server MockServer) Listen(port int) {
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.router,
	}
	server.router.LogPatterns(s)
	log.Print(s.ListenAndServe())
}
