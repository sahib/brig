// Package log implements utility methods for logging in a colorful manner.
package log

import (
	"bytes"
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

// ColorfulLogFormatter is the default logger for brig.
type ColorfulLogFormatter struct{}

var symbolTable = map[logrus.Level]string{
	logrus.DebugLevel: "⚙",
	logrus.InfoLevel:  "⚐",
	logrus.WarnLevel:  "⚠",
	logrus.ErrorLevel: "⚡",
	logrus.FatalLevel: "☣",
	logrus.PanicLevel: "☠",
}

var colorTable = map[logrus.Level]color.Attribute{
	logrus.DebugLevel: color.FgCyan,
	logrus.InfoLevel:  color.FgGreen,
	logrus.WarnLevel:  color.FgYellow,
	logrus.ErrorLevel: color.FgRed,
	logrus.FatalLevel: color.FgMagenta,
	logrus.PanicLevel: color.BgRed,
}

func colorByLevel(level logrus.Level) *color.Color {
	attr, ok := colorTable[level]
	if !ok {
		attr = color.Reset
	}

	return color.New(attr)
}

func formatColored(buffer *bytes.Buffer, msg string, level logrus.Level) {
	color.Output = buffer
	colorByLevel(level).Set()
	buffer.WriteString(msg)
	color.Unset()
}

func formatTimestamp(buffer *bytes.Buffer, t time.Time) {
	fmt.Fprintf(buffer, "%02d.%02d.%04d", t.Day(), t.Month(), t.Year())
	buffer.WriteByte('/')
	fmt.Fprintf(buffer, "%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}

func formatFields(buffer *bytes.Buffer, entry *logrus.Entry) {
	idx := 0
	buffer.WriteString(" [")

	for key, value := range entry.Data {
		// Make the key colored:
		formatColored(buffer, key, entry.Level)
		buffer.WriteByte('=')

		// A few special cases depending on the type:
		switch v := value.(type) {
		case *logrus.Entry:
			formatColored(buffer, v.Message, logrus.ErrorLevel)
		default:
			buffer.WriteString(fmt.Sprintf("%v", v))
		}

		// Print no space after the last element:
		if idx != len(entry.Data)-1 {
			buffer.WriteByte(' ')
		}

		idx++
	}

	buffer.WriteByte(']')
}

// Format logs a single entry according to our formatting ideas.
func (*ColorfulLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buffer := &bytes.Buffer{}

	// Add the timestamp:
	color.Output = buffer
	defer func() {
		color.Output = os.Stdout
	}()

	colorByLevel(entry.Level).Set()
	formatTimestamp(buffer, entry.Time)
	buffer.WriteByte(' ')

	// Add the symbol:
	buffer.WriteString(symbolTable[entry.Level])
	color.Unset()

	// Add the actual message:
	buffer.WriteByte(' ')
	buffer.WriteString(entry.Message)

	// Add the fields, if any:
	if len(entry.Data) > 0 {
		formatFields(buffer, entry)
	}

	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

var logLevelToFunc = map[logrus.Level]func(args ...interface{}){
	logrus.DebugLevel: logrus.Debug,
	logrus.InfoLevel:  logrus.Info,
	logrus.WarnLevel:  logrus.Warn,
	logrus.ErrorLevel: logrus.Error,
	logrus.FatalLevel: logrus.Fatal,
}

// Writer is an io.Writer that writes everything to logrus.
type Writer struct {
	// Level determines the severity for all messages.
	Level logrus.Level
}

func (l *Writer) Write(buf []byte) (int, error) {
	fn, ok := logLevelToFunc[l.Level]
	if !ok {
		logrus.Fatal("LogWriter: Bad loglevel passed.")
	} else {
		msg := string(buf)
		fn(strings.Trim(msg, "\n\r "))
	}

	return len(buf), nil
}

// SyslogWrapper is a hacky way to make the syslog more readable.
// This only works with ColorfulLogFormatter from above.
// It takes it's output, checks what log level was used and
// puts it into syslog with the right notice level.
type SyslogWrapper struct {
	w *syslog.Writer
}

func NewSyslogWrapper(w *syslog.Writer) SyslogWrapper {
	return SyslogWrapper{w: w}
}

func (sw SyslogWrapper) Write(data []byte) (int, error) {
	prefix := data

	if len(data) < 23 {
		return len(data), sw.w.Info(string(data))
	}

	// The logging symbol is currently definitely in the this part
	// of the log. It might span up to 4 bytes.
	prefix = data[19:23]
	if bytes.Index(prefix, []byte(symbolTable[logrus.DebugLevel])) > 0 {
		return len(data), sw.w.Debug(string(data))
	}

	if bytes.Index(prefix, []byte(symbolTable[logrus.InfoLevel])) > 0 {
		return len(data), sw.w.Info(string(data))
	}

	if bytes.Index(prefix, []byte(symbolTable[logrus.WarnLevel])) > 0 {
		return len(data), sw.w.Warning(string(data))
	}

	return len(data), sw.w.Err(string(data))
}
