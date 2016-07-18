package website

import (
	"github.com/fatih/color"
	"log"

	. "api/conf"
)

func IsDebugging() bool {
	return GetSettings().Debug
}

func DebugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		log.Printf(""+format, values...)
	}
}

func DebugPrintError(err error) {
	if err != nil {
		color.Set(color.FgRed)
		DebugPrint("[ERROR] %v\n", err)
		color.Unset()
	}
}
