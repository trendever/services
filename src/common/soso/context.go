package soso

import (
	"common/log"
	"encoding/json"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	"strings"
)

var (
	actions = map[string]string{
		"create":   "CREATED",
		"retrieve": "RETRIEVED",
		"update":   "UPDATED",
		"delete":   "DELETED",
		"flush":    "FLUSHED",
	}

	middlewares = []func(*Request, *Context, Session) error{
		IPMiddleware,
	}
)

func AddMiddleware(fn func(*Request, *Context, Session) error) {
	middlewares = append(middlewares, fn)
}

type Token struct {
	UID uint64
	Exp int64
}

// Context of the request
type Context struct {
	Domain     string
	Method     string
	LogList    []Log
	RequestMap map[string]interface{}
	RawRequest json.RawMessage
	TransMap   map[string]interface{}

	Response *Response

	// Client socket session, public for testing convinience
	Session Session

	Token    *Token
	RemoteIP string
}

func NewContext(req *Request, session Session) *Context {
	ctx := &Context{
		Domain:     req.Domain,
		Method:     req.Method,
		Session:    session,
		RawRequest: req.RequestData,
		TransMap:   req.TransMap,
		LogList:    req.LogList,
	}
	ctx.Response = NewResponse(ctx)

	for _, mw := range middlewares {
		if err := mw(req, ctx, session); err != nil {
			log.Error(err)
			ctx.ErrorResponse(http.StatusUnauthorized, LevelError, err)
			return nil
		}
	}
	if ctx.Token != nil {
		Sessions.Push(session, ctx.Token.UID)
	}

	return ctx
}

func IPMiddleware(req *Request, ctx *Context, session Session) error {

	request := session.Request()
	if request == nil {
		return fmt.Errorf("No request (can not get client IP)")
	}

	forwarded := request.Header.Get("X-Forwarded-For")
	addr := strings.Split(forwarded, ", ")[0]

	if addr == "" {
		// addr format: "127.0.0.1:4242"
		tokens := strings.Split(request.RemoteAddr, ":")
		if len(tokens) != 2 {
			return fmt.Errorf("Can not parse remote addr (%v)", request.RemoteAddr)
		}
		addr = tokens[0]
	}

	ctx.RemoteIP = addr
	return nil
}

func NewRemoteContext(dataType, action string, response map[string]interface{}) *Context {
	ctx := &Context{
		Domain: dataType,
		Method: action,
	}
	ctx.Response = NewResponse(ctx)
	ctx.Response.ResponseMap = response
	return ctx
}

func (c *Context) sendJSON(data interface{}) {
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Errorf("failed to marshal json: %v", err)
	}
	if err := c.Session.Send(string(json_data)); err == sockjs.ErrSessionNotOpen {
		Sessions.Pull(c.Session)
	}
}

func (c *Context) SendResponse() {
	c.Response.Log(log_code_by_action_type(c.Method), LevelDebug, "")
	c.sendJSON(c.Response)
}

func (c *Context) ErrorResponse(code int, level Level, err error) {
	c.Response.Log(code, level, err.Error())
	c.sendJSON(c.Response)
}

func (c *Context) SuccessResponse(ResponseMap interface{}) {
	c.Response.ResponseMap = ResponseMap
	c.Response.Log(log_code_by_action_type(c.Method), LevelDebug, "")

	c.sendJSON(c.Response)
}

func reverse_action_type(action_str string) string {
	act, ok := actions[action_str]
	if !ok {
		act = strings.ToUpper(action_str)
	}
	return act
}

func log_code_by_action_type(action_str string) int {
	if action_str == "create" {
		return 201
	}
	return 200
}
