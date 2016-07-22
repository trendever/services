package log

import (
	"fmt"
	"github.com/getsentry/raven-go"
	golog "log"
	"os"
	"runtime"
	"strings"
)

//ErrorHandler is a function for handling errors
//this function will be called after error logging to std error output
type ErrorHandler func(err error, tags map[string]string)

//MessageHandler is a function for handling messages not errors
type MessageHandler func(msg string, tags map[string]string)

//Log levels
const (
	//LevelDebug for debug messages. Visible only in debug mode
	LevelDebug = "debug"
	//LevelError for error messages
	LevelError = "error"
	//LevelFatal for fatal error messages
	LevelFatal = "fatal"
	//LevelPanic for panic messages
	LevelPanic = "panic"
	//LevelInfo for info messages
	LevelInfo = "info"
	//LevelWarn for warn message, not error but pay attention
	LevelWarn = "warning"
)

var (
	debug          = false
	tag            = "Trendever"
	errorHandler   ErrorHandler
	messageHandler MessageHandler
	errLogger      *golog.Logger
	infoLogger     *golog.Logger
	ravenClient    *raven.Client
)

func init() {
	errLogger = golog.New(os.Stderr, "", golog.LstdFlags)
	infoLogger = golog.New(os.Stdout, "", golog.LstdFlags)
}

// Init initializes logging parameters
//  * logMode: set verbose mode
//  * tagMark: messages tag
//  * errHandler: function for handling errors
func Init(debugMode bool, tagMark string, sentryDSN string) {
	debug = debugMode
	tag = tagMark
	if sentryDSN != "" {
		var err error
		ravenClient, err = raven.NewClient(sentryDSN, map[string]string{
			"service": tagMark,
		})
		if err != nil {
			Error(err)
		} else {
			errorHandler = ravenErrorLogger
			messageHandler = ravenMessageLogger
		}
	}
}

func msgToStr(msg interface{}) string {
	strMsg, ok := msg.(string)
	if !ok {
		strMsg = fmt.Sprintf("%v", msg)
	}
	return strMsg
}

func msgToError(msg interface{}) error {
	errMsg, ok := msg.(error)
	if !ok {
		errMsg = fmt.Errorf("%v", msg)
	}

	return errMsg
}

func log(level string, msg interface{}) {

	switch {
	case level == LevelInfo || level == LevelWarn:
		logMessage(level, msgToStr(msg))
	case level == LevelDebug && debug:
		logMessage(level, msgToStr(msg))
	case level == LevelError || level == LevelFatal || level == LevelPanic:
		logError(level, msgToError(msg))
	}
}

func logMessage(level, msg string) {
	infoLogger.Printf("[%s] [%s] %s\n", tag, level, msg)
	if messageHandler != nil {
		messageHandler(msg, map[string]string{
			"level": level,
		})
	}
}

func logError(level string, err error) {
	f := filenameWithLineNum()
	errLogger.Printf("[%s] [%s:%s] %v\n", tag, level, f, err)
	if errorHandler != nil {
		errorHandler(err, map[string]string{
			"level": level,
		})
	}
}

//PanicLogger is a safely logging panic for a callback
func PanicLogger(f func()) {
	defer func() {
		if r := recover(); r != nil {
			log(LevelPanic, fmt.Errorf("%v", r))
		}
	}()
	f()
}

//Info puts info log
func Info(format string, values ...interface{}) {
	log(LevelInfo, fmt.Sprintf(format, values...))
}

//Debug puts debug log
func Debug(format string, values ...interface{}) {
	log(LevelDebug, fmt.Sprintf(format, values...))
}

//Warn puts warn log
func Warn(format string, values ...interface{}) {
	log(LevelWarn, fmt.Sprintf(format, values...))
}

// Error puts error log
func Error(err error) {
	if err == nil {
		return
	}
	log(LevelError, err)
}

//Fatal puts fatal error log
func Fatal(err error) {
	if err == nil {
		return
	}
	log(LevelFatal, err)
	os.Exit(1)
}

func filenameWithLineNum() string {
	var total = 10
	var results []string
	for i := 2; i < 15; i++ {
		if _, file, line, ok := runtime.Caller(i); ok {
			total--
			results = append(results[:0],
				append(
					[]string{fmt.Sprintf("%v:%v", strings.TrimPrefix(file, os.Getenv("GOPATH")+"src/"), line)},
					results[0:]...)...)

			if total == 0 {
				return strings.Join(results, "\n")
			}
		}
	}
	return ""
}
