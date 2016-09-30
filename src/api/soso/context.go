package soso

import (
	"api/auth"
	. "api/debug"
	"encoding/json"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	auth_protocol "proto/auth"
	"strings"
	"utils/log"
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
		TokenMiddleware,
		IPMiddleware,
	}
)

// Context of the request
type Context struct {
	DataType   string
	ActionStr  string
	LogList    []Log
	RequestMap map[string]interface{}
	TransMap   map[string]interface{}

	Response *Response

	// Client socket session, public for testing convinience
	Session Session

	Token    *auth_protocol.Token
	RemoteIP string
}

func NewContext(req *Request, session Session) *Context {
	ctx := &Context{
		DataType:   req.DataType,
		ActionStr:  req.ActionStr,
		Session:    session,
		RequestMap: req.RequestMap,
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

	return ctx
}

func TokenMiddleware(req *Request, ctx *Context, session Session) error {
	if token, ok := req.TransMap["token"].(string); ok {
		tokenObj, err := auth.GetTokenData(token)
		if err != nil {
			return err
		}

		ctx.Token = tokenObj
		Sessions.Push(session, tokenObj.UID)
	}

	return nil
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
		DataType:  dataType,
		ActionStr: action,
	}
	ctx.Response = NewResponse(ctx)
	ctx.Response.ResponseMap = response
	return ctx
}

func (c *Context) sendJSON(data interface{}) {
	json_data, err := json.Marshal(data)
	if err != nil {
		DebugPrintError(err)
	}
	if err := c.Session.Send(string(json_data)); err == sockjs.ErrSessionNotOpen {
		Sessions.Pull(c.Session)
	}
}

func (c *Context) SendResponse() {
	c.Response.Log(log_code_by_action_type(c.ActionStr), LevelDebug, "")
	c.sendJSON(c.Response)
}

func (c *Context) ErrorResponse(code int, level Level, err error) {
	c.Response.Log(code, level, err.Error())
	c.sendJSON(c.Response)
}

func (c *Context) SuccessResponse(ResponseMap interface{}) {
	c.Response.ResponseMap = ResponseMap
	c.Response.Log(log_code_by_action_type(c.ActionStr), LevelDebug, "")

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
