package soso

import (
	"common/log"
	"common/metrics"
	"encoding/json"
	"errors"
	"github.com/fatih/color"
	"net/http"
	"reflect"
	"time"
)

type (
	HandlerFunc func(*Context)

	Route struct {
		Domain string
		Method string
		// should be
		// func(*Context) for old-style handlers thad uses RequestMap in context
		// or func(*Context, whatever) for typed handlers (RequestMap will be nil in this case)
		Handler interface{}
	}

	Router struct {
		// domain -> method -> handler
		handlers map[string]map[string]HandlerFunc
	}
)

func NewRouter() *Router {
	return &Router{handlers: map[string]map[string]HandlerFunc{}}
}

// determine whether handler has suitable type and type of its second argument
func checkHandler(handler interface{}) (suitable bool, argType reflect.Type) {
	hType := reflect.TypeOf(handler)
	if hType.Kind() != reflect.Func || hType.NumOut() != 0 {
		return false, nil
	}
	switch {
	case hType.NumIn() == 1:
	case hType.NumIn() == 2:
		argType = hType.In(1)
	default:
		return false, nil
	}
	if hType.In(0) != reflect.TypeOf(&Context{}) {
		return false, nil
	}
	return true, argType
}

func (r *Router) Handle(domain string, method string, handler interface{}) {
	ok, argType := checkHandler(handler)
	if !ok {
		log.Fatalf("handler for %v/%v has unexpected type", domain, method)
	}

	hValue := reflect.ValueOf(handler)
	var prepared HandlerFunc
	if argType == nil {
		// old-style handler, simple decode request into RequestMap
		prepared = func(ctx *Context) {
			err := json.Unmarshal([]byte(ctx.RawRequest), &ctx.RequestMap)
			if err != nil {
				ctx.ErrorResponse(http.StatusBadRequest, LevelError, errors.New("invalid json"))
				return
			}
			hValue.Call([]reflect.Value{reflect.ValueOf(ctx)})
		}
	} else {
		prepared = func(ctx *Context) {
			var argPtr reflect.Value
			if argType.Kind() != reflect.Ptr {
				argPtr = reflect.New(argType)
			} else {
				argPtr = reflect.New(argType.Elem())
			}
			if err := json.Unmarshal([]byte(ctx.RawRequest), argPtr.Interface()); err != nil {
				// probably there is no need to tell everyone what exactly happened
				log.Warn("failed to unmarshal argument for handler %v/%v: %v", domain, method, err)
				ctx.ErrorResponse(http.StatusBadRequest, LevelError, errors.New("bad request"))
				return
			}
			hValue.Call([]reflect.Value{reflect.ValueOf(ctx), argPtr})
		}
	}

	dom, ok := r.handlers[domain]
	if !ok {
		dom = map[string]HandlerFunc{}
		r.handlers[domain] = dom
	}
	dom[method] = prepared
}

func (r *Router) HandleRoutes(routes []Route) {
	for _, route := range routes {
		r.Handle(route.Domain, route.Method, route.Handler)
	}
}

func (r *Router) execute(session Session, msg string) {
	startTime := time.Now()

	req, err := NewRequest(msg)

	if err != nil {
		log.Debug("%s Error: incorrect request - ", logPrefix, msg)
		log.Error(err)
		return
	}

	ctx := NewContext(req, session)

	if ctx == nil {
		return
	}

	var handler HandlerFunc
	domain, found := r.handlers[req.Domain]
	if found {
		handler, found = domain[req.Method]
	}

	if !found {
		log.Debug("%s %s | %s -> %s | %s",
			logPrefix,
			time.Now().Format("2006/01/02 - 15:04:05"),
			color.RedString(req.Domain),
			color.GreenString(req.Method),
			"Route not found",
		)
		ctx.ErrorResponse(http.StatusNotFound, LevelError, errors.New("No model handler found"))
		return
	}

	handler(ctx)
	code := ""
	if ctx.Response != nil && len(ctx.Response.LogList) > 0 {
		code = ctx.Response.LogList[0].CodeKey
	}

	elapsedTime := time.Since(startTime)

	metrics.Add(
		"request_time",
		map[string]string{
			"service": "api",
			"action":  req.Method,
			"type":    req.Domain,
			"status":  code,
		},
		map[string]interface{}{
			"value": metrics.ToMs(elapsedTime),
		},
	)
	log.Debug("%s %s | %s -> %s | %s",
		logPrefix,
		startTime.Format("2006/01/02 - 15:04:05"),
		color.GreenString(req.Method),
		color.YellowString(req.Domain),
		elapsedTime,
	)
}
