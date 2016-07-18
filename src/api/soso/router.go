package soso

import (
	"errors"
	"github.com/fatih/color"
	"utils/log"
	"utils/metrics"
	"net/http"
	"time"
)

type (
	HandlerFunc func(*Context)

	Route struct {
		ActionStr string
		DataType  string

		Handler HandlerFunc
	}

	Router struct {
		Routes []Route
	}
)

func (r *Router) Handle(action_str string, data_type string, handler HandlerFunc) {
	route := Route{
		ActionStr: action_str,
		DataType:  data_type,
		Handler:   handler,
	}

	r.Routes = append(r.Routes, route)
}

func (r *Router) HandleList(routes []Route) {
	r.Routes = append(r.Routes, routes...)
}

func (r *Router) HandleLists(group [][]Route) {
	for _, lst := range group {
		r.HandleList(lst)
	}
}

func (r *Router) execute(session Session, msg string) {
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

	// обработка списка хендлеров
	for _, route := range r.Routes {
		if req.ActionStr == route.ActionStr && req.DataType == route.DataType {

			startTime := time.Now()

			route.Handler(ctx)

			elapsedTime := time.Since(startTime)
			code := ""
			if ctx.Response != nil && len(ctx.Response.LogList) > 0 {
				code = ctx.Response.LogList[0].CodeKey
			}

			metrics.Add(
				"request_time",
				map[string]string{
					"service": "api",
					"action":  req.ActionStr,
					"type":    req.DataType,
					"status":  code,
				},
				map[string]interface{}{
					"value": metrics.ToMs(elapsedTime),
				},
			)
			log.Debug("%s %s | %s -> %s | %s",
				logPrefix,
				startTime.Format("2006/01/02 - 15:04:05"),
				color.GreenString(req.ActionStr),
				color.YellowString(req.DataType),
				elapsedTime,
			)

			return
		}
	}

	log.Debug("%s %s | %s -> %s | %s",
		logPrefix,
		time.Now().Format("2006/01/02 - 15:04:05"),
		color.GreenString(req.ActionStr),
		color.RedString(req.DataType),
		"Route not found",
	)
	ctx.ErrorResponse(http.StatusNotFound, LevelError, errors.New("No model handler found"))

}
