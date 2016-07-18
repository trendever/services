package soso

import (
	"encoding/json"
)

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

   resp := Request{
      ActionStr: "retrieve",
      DataType: "person",
      LogList: log
      RequestMap: {}
      TransMap:
   }

   Other:

    trans_map  : {
      auth_key : "c76aa3577f8b5a60206f9d041c76034a...",
      trans_id : "eb99ec08-7e90-400d-9585-62a1385ec158"
    }

*/

// Request
type Request struct {
	ActionStr  string                 `json:"action_str"`
	DataType   string                 `json:"data_type"`
	LogList    []Log                  `json:"log_list"`
	RequestMap map[string]interface{} `json:"request_map"`
	TransMap   map[string]interface{} `json:"trans_map"`
}

func NewRequest(msg string) (*Request, error) {
	var req *Request
	err := json.Unmarshal([]byte(msg), &req)
	return req, err
}
