package views

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
