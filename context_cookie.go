package xhttp

import "net/http"

func (c *Context) Cookie(name string) (value *http.Cookie, has bool, err error) {
	var cookie *http.Cookie
	cookie, err = c.Request.Cookie(name)
	switch err {
	case nil:
		return cookie, true, nil
	case http.ErrNoCookie:
		return nil, false, nil
	default:
		return nil, false, err
	}
}
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}
func (c *Context) ClearCookie(cookie *http.Cookie) {
	// -1 标识清除cookie
	cookie.MaxAge = -1
	http.SetCookie(c.Writer, cookie)
}
