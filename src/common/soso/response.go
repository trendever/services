package soso

/*
   Example:

   log := Log{
       CodeKey: "400",
       CodeStr: "Bad request",

       LevelInt: int(Level(3)), // or int(LevelError)
       LevelStr: Level(3), // or obj of LevelError

       LogID: "1096",
       UserMsg: "action_str required"
   }

   resp := Response{
      ActionStr: "RETRIEVED",
      DataType: "person",
      LogList: log
      ResponseMap: {}
      TransMap: {}
   }

    Other:

    trans_map  : map[string]interface{}{
      auth_key : "c76aa3577f8b5a60206f9d041c76034a",
      trans_id : "eb99ec08-7e90-400d-9585-62a1385ec158"
    }
*/

// direct and indirect responses
type Response struct {
	Domain      string      `json:"data_type"`
	Method      string      `json:"action_str"`
	LogList     []Log       `json:"log_list"`
	ResponseMap interface{} `json:"response_map"`
	TransMap    interface{} `json:"trans_map"`
}

func NewResponse(ctx *Context) *Response {
	return &Response{
		Domain:   ctx.Domain,
		Method:   reverse_action_type(ctx.Method),
		TransMap: ctx.TransMap,
	}
}

func (r *Response) Log(code_key int, lvl_str Level, user_msg string) *Response {
	r.LogList = append(r.LogList, NewLog(code_key, lvl_str, user_msg))
	return r
}

func (r Response) Result() Response { return r }
