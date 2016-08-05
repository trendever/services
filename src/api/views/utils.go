package views

import (
	"api/soso"
	"errors"
	"strings"
)

// Convert []interface{} array with float64 values to []int64 array
func getIntArr(arr []interface{}) []int64 {
	out := make([]int64, len(arr), len(arr))
	for i, v := range arr {
		if floatVal, ok := v.(float64); ok {
			out[i] = int64(floatVal)
		}
	}

	return out
}

// getIP returns this session ip address
func getIP(session soso.Session) (string, error) {

	request := session.Request()
	if request == nil {
		return "", errors.New("Nil request in getIP()")
	}

	// addr format: "127.0.0.1:4242"
	remoteAddr := strings.Split(request.RemoteAddr, ":")
	if len(remoteAddr) != 2 {
		return "", errors.New("Invalid request.RemoteAddr in getIP()")
	}

	return remoteAddr[0], nil
}
