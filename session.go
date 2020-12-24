package xhttp

import (
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"net/http"
)

type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
}

type Session struct {
	name string
	store SessionStore
	c *Context
}
func NewSession(name string, sessionStore SessionStore, c *Context) Session {
	return Session{
		name: name,
		store: sessionStore,
		c: c,
	}
}
func (s Session) GetString(key string) (value string, has bool, err error) {
	sess, err := s.store.Get(s.c.Request, s.name)
	if err != nil {return }
	valueInterface := sess.Values[key]
	switch valueInterface.(type) {
	case string:
		return valueInterface.(string), true, nil
	case nil:
		return "",false, nil
	default:
		return "", false, errors.New("juice.Session{}.GetString(key string)(value string err error) value type is not string")
	}
}
func (s Session) SetString(key string, value string) (err error) {
	sess, err := s.store.Get(s.c.Request, s.name)
	if sess == nil {
		panic(errors.New("session is nil"))
	}
	sess.Values[key] = value
	err = sess.Save(s.c.Request, s.c.Writer) ; if err != nil {return}
	return nil
}
func (s Session) DelString(key string) (err error) {
	sess, err := s.store.Get(s.c.Request, s.name)
	if sess == nil {
		panic(errors.New("session is nil"))
	}
	delete(sess.Values, key)
	err = sess.Save(s.c.Request, s.c.Writer) ; if err != nil {return}
	return nil
}

func (c *Context) Session(sessionName string, sessionStore SessionStore) Session {
	return NewSession(sessionName, sessionStore, c)
}
