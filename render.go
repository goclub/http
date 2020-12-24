package xhttp

import (
	xjson "github.com/goclub/json"
)

type Helper struct {

}
func (Helper) JSON(v interface{}) []byte {
	data, err := xjson.Marshal(v)
	if err != nil {
		return []byte("Error: render helper JSON Fail")
	}
	return data
}
