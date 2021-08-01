package xhttp



type Method string
func (m Method) String() string {
	return string(m)
}
const GET Method = "GET"
const POST Method = "POST"
const HEAD Method = "HEAD"
const PUT Method = "PUT"
const DELETE Method = "DELETE"
const CONNECT Method = "CONNECT"
const OPTIONS Method = "OPTIONS"
const TRACE Method = "TRACE"
const PATCH Method = "PATCH"

type Route struct {
	Method Method
	Path string
}
func (route Route) ID () string {
	return route.Method.String() + " " + route.Path
}
func (p Route) Equal(target Route) bool {
	return p.Method == target.Method && p.Path == target.Path
}