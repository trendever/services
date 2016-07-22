package log

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLog(t *testing.T) {
	call := false
	Init(false, "TEST", "")
	errorHandler = func(msg error, tags map[string]string) {
		call = true
		assert.Equal(t, "test error", msg.Error(), "Must be equal")
		assert.Equal(t, LevelError, tags["level"])
	}

	Error(errors.New("test error"))

	assert.True(t, call, "Error handler callback didn't been called")
}

func TestPanicHandler(t *testing.T) {
	Init(false, "TEST", "")
	errorHandler = func(msg error, tags map[string]string) {
		assert.Equal(t, "test error", msg.Error(), "Must be equal")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Error("Panic not catched!")
		}
	}()

	PanicLogger(func() {
		panic("test error")
	})

}

func TestDebugLog(t *testing.T) {
	call := false
	Init(true, "TEST", "")

	messageHandler = func(msg string, tags map[string]string) {
		call = true
		assert.Equal(t, "Some value", msg)
	}

	Debug("Some value")

	assert.True(t, call, "Error handler callback didn't been called")
}
