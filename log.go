package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type LogLevel int

const (
	Fatal LogLevel = iota
	Error
	Info
	Debug
)

// Set the global log level
const Level LogLevel = Debug

func InfoLog(format string) {
	if Level < Info {
		return
	}
	color.New(color.FgHiBlack).Printf("%s\n", format)
}
func InfoLogf(format string, args ...any) {
	if Level < Info {
		return
	}
	color.New(color.FgHiBlack).Printf(fmt.Sprintf("%s\n", format), args)
}

func DebugLog(format string) {
	if Level < Debug {
		return
	}
	color.New(color.FgCyan).Printf("%s\n", format)
}
func DebugLogf(format string, args ...any) {
	if Level < Debug {
		return
	}
	color.New(color.FgCyan).Printf(fmt.Sprintf("%s\n", format), args)
}

func ErrorLog(format string) {
	if Level < Error {
		return
	}
	color.New(color.FgRed).Printf("%s\n", format)
}
func ErrorLogf(format string, args ...any) {
	if Level < Error {
		return
	}
	color.New(color.FgRed).Printf(fmt.Sprintf("%s\n", format), args)
}

func FatalLog(format string) {
	color.New(color.BgRed).Add(color.FgBlack).Add(color.Bold).Printf("%s\n", format)
	os.Exit(1)
}
func FatalLogf(format string, args ...any) {
	color.New(color.BgRed).Add(color.FgBlack).Add(color.Bold).Printf(fmt.Sprintf("%s\n", format), args)
	os.Exit(1)
}
