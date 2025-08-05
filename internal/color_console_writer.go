package internal

import (
	"io"
	"log"
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type colorConsoleWriter struct {
	outLog io.WriteCloser
}

func newColorConsoleWriter() Writer {
	lw := newLogWriter(log.New(os.Stdout, "", flags))
	return &colorConsoleWriter{
		outLog: lw,
	}
}

func (w *colorConsoleWriter) Close() error {
	return w.outLog.Close()
}

func isColorSupported() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func (w *colorConsoleWriter) Trace(v any) {
	if isColorSupported() {
		output(w.outLog, color.HiCyanString(LevelTrace), v)
	} else {
		output(w.outLog, LevelTrace, v)
	}
}

func (w *colorConsoleWriter) Debug(v any) {
	if isColorSupported() {
		output(w.outLog, color.GreenString(LevelDebug), v)
	} else {
		output(w.outLog, LevelDebug, v)
	}
}

func (w *colorConsoleWriter) Error(v any) {
	if isColorSupported() {
		output(w.outLog, color.RedString(LevelError), v)
	} else {
		output(w.outLog, LevelError, v)
	}
}

func (w *colorConsoleWriter) Warn(v any) {
	if isColorSupported() {
		output(w.outLog, color.YellowString(LevelWarn), v)
	} else {
		output(w.outLog, LevelWarn, v)
	}
}

func (w *colorConsoleWriter) Info(v any) {
	if isColorSupported() {
		output(w.outLog, color.CyanString(LevelInfo), v)
	} else {
		output(w.outLog, LevelInfo, v)
	}
}

func (w *colorConsoleWriter) AccessRecord(v any) {
	if isColorSupported() {
		output(w.outLog, levelAccessRecord, v)
	} else {
		output(w.outLog, levelAccessRecord, v)
	}
}

func (w *colorConsoleWriter) WriteRawString(v string) {
	writer := w.outLog
	if writer == nil {
		log.Print(v)
		return
	}

	if _, err := writer.Write([]byte(v)); err != nil {
		log.Println(err.Error())
	}
}
