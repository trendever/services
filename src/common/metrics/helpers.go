package metrics

import "time"

//ToMs converts time.Duration to milliseconds
func ToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()/int64(time.Millisecond)) + float64(d.Nanoseconds()%int64(time.Millisecond))*1e-9
}
