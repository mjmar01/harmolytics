package log

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-errors/errors"
	"os"
)

const (
	TraceLevel = -1
	DebugLevel = 0
	InfoLevel  = 1
	WarnLevel  = 2
	ErrorLevel = 3
	FatalLevel = 4
	PanicLevel = 5
)

var (
	logLevel     = 1
	depth        = 0
	taskLevels   = []int{logLevel}
	taskLevel    int
	loggedOnTask = false
)

func CheckErr(err error, level int) bool {
	if err != nil {
		if level == PanicLevel {
			Panic(err)
		} else {
			logAtLevel(err.Error(), level)
		}
		return true
	}
	return false
}

func SetLogLevel(level int) {
	logLevel = level
	taskLevels = []int{logLevel}
}

func Task(msg string, level int) {
	taskLevel = level
	taskLevels = append(taskLevels, taskLevel)
	if taskLevel >= logLevel {
		logAtLevel(msg, taskLevel)
		loggedOnTask = false
		depth++
	}
}

func Done() {
	if taskLevel >= logLevel {
		depth--
	}
	if !loggedOnTask && taskLevel >= logLevel {
		fmt.Print(" DONE!")
	} else {
		if taskLevel >= logLevel {
			logAtLevel("DONE!", taskLevel)
		}
	}
	taskLevels = taskLevels[:len(taskLevels)-1]
	taskLevel = taskLevels[len(taskLevels)-1]
}

func Trace(msg string) {
	if logLevel <= TraceLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgCyan)
		fmt.Print("TRACE: ")
		color.Unset()
		fmt.Print(msg)
		loggedOnTask = true
	}
}

func Debug(msg string) {
	if logLevel <= DebugLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgBlue)
		fmt.Print("DEBUG: ")
		color.Unset()
		fmt.Print(msg)
		loggedOnTask = true
	}
}

func Info(msg string) {
	if logLevel <= InfoLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgGreen)
		fmt.Print("INFO:  ")
		color.Unset()
		fmt.Print(msg)
		loggedOnTask = true
	}
}

func Warn(msg string) {
	if logLevel <= WarnLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgYellow)
		fmt.Print("WARN:  ")
		color.Unset()
		fmt.Print(msg)
		loggedOnTask = true
	}
}

func Error(msg string) {
	if logLevel <= ErrorLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgRed)
		fmt.Print("ERROR: ")
		color.Unset()
		fmt.Print(msg)
		loggedOnTask = true
	}
}

func Fatal(msg string) {
	if logLevel <= FatalLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgMagenta)
		fmt.Print("FATAL: ")
		color.Unset()
		fmt.Print(msg)
	}
	fmt.Print("\n")
	os.Exit(1)
}

func Panic(err error) {
	if logLevel <= PanicLevel {
		fmt.Print("\n")
		addDepth()
		color.Set(color.FgBlack)
		fmt.Print("PANIC: ")
		color.Unset()
		fmt.Println(err.Error())
	}
	fmt.Print("\n")
	switch err.(type) {
	case *errors.Error:
		fmt.Println(err.(*errors.Error).ErrorStack())
	}
	os.Exit(1)
}

func logAtLevel(msg string, level int) {
	switch level {
	case TraceLevel:
		Trace(msg)
		break
	case DebugLevel:
		Debug(msg)
		break
	case InfoLevel:
		Info(msg)
		break
	case WarnLevel:
		Warn(msg)
		break
	case ErrorLevel:
		Error(msg)
		break
	case FatalLevel:
		Fatal(msg)
		break
	}
}

func addDepth() {
	for i := 0; i < depth; i++ {
		fmt.Print("|  ")
	}
}
