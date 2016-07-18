package soso

import (
	"encoding/json"
	auth_protocol "proto/auth"
	"utils/log"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"net/http"
	"api/auth"
	. "api/debug"
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
)

type Context struct {
	DataType   string
	ActionStr  string
	LogList    []Log
	RequestMap map[string]interface{}
	TransMap   map[string]interface{}

	Response *Response

	// Client socket session, public for testing convinience
	Session Session

	Token *auth_protocol.Token
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
	if token, ok := req.TransMap["token"].(string); ok {
		tokenObj, err := auth.GetTokenData(token)
		if err != nil {
			log.Error(err)
			ctx.ErrorResponse(http.StatusUnauthorized, LevelError, err)
			return nil
		}

		ctx.Token = tokenObj
		Sessions.Push(session, tokenObj.UID)
	}
	return ctx
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
