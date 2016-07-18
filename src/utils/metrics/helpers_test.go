package metrics

import (
	"testing"
	"time"
)

func TestToMs(t *testing.T){
	d:= time.Millisecond * 10

	if ToMs(d)!=10 {
		t.Fail()
	}
}
