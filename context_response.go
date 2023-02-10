package xhttp

import (
	"bytes"
	xjson "github.com/goclub/json"
	"net/http"
)

func (c *Context) Redirect(url string, code int) (err error) {
	http.Redirect(c.Writer, c.Request, url, code)
	return
}

func (c *Context) WriteStatusCode(statusCode int) {
	c.Writer.WriteHeader(statusCode)
}

// WriteBytes 等同于 writer.Write(data) ,但函数签名返回 error 不返回 int
func (c *Context) WriteBytes(b []byte) error {
	_, err := c.Writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// Render 设置 header Content-Type text/html; charset=UTF-8 并输出 buffer
func (c *Context) Render(render func(buffer *bytes.Buffer) error) error {
	buffer := bytes.NewBuffer(nil)
	err := render(buffer)
	if err != nil {
		return err
	}
	c.Writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	return c.WriteBytes(buffer.Bytes())
}

// WriteJSON 响应 json
func (c *Context) WriteJSON(v interface{}) error {
	data, err := xjson.Marshal(v)
	if err != nil {
		return err
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	return c.WriteBytes(data)
}
