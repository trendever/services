package test_tools

import (
	"encoding/json"
	"fmt"
)

const expectedButGot = "\nExpected %s must be %v, but got %v"

func ToJson(in interface{}) interface{} {
	result, err := json.MarshalIndent(in, "", " ")
	if err != nil {
		return ToStr(in)
	}
	return string(result)
}

func ToStr(in interface{}) string {
	return fmt.Sprintf("%v", in)
}

func ExpectedButGot(args ...interface{}) string {
	for i, arg := range args {
		args[i] = ToJson(arg)
	}
	return fmt.Sprintf(expectedButGot, args...)
}
