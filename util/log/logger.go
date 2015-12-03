// Package log implements utility methods for logging in a colorful manner.
package log

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util/colors"
)

type ColorfulLogFormatter struct{}

var symbolTable = map[logrus.Level]string{
	logrus.DebugLevel: "⚙",
	logrus.InfoLevel:  "⚐",
	logrus.WarnLevel:  "⚠",
	logrus.ErrorLevel: "⚡",
	logrus.FatalLevel: "☣",
	logrus.PanicLevel: "☠",
}

var colorTable = map[logrus.Level]int{
	logrus.DebugLevel: colors.Cyan,
	logrus.InfoLevel:  colors.Green,
	logrus.WarnLevel:  colors.Yellow,
	logrus.ErrorLevel: colors.Red,
	logrus.FatalLevel: colors.Magenta,
	logrus.PanicLevel: colors.BackgroundRed,
}

func colorEscape(level logrus.Level) string {
	return colors.ColorEscape(colorTable[level])
}

func formatColored(buffer *bytes.Buffer, msg string, level logrus.Level) {
	buffer.WriteString(colorEscape(level))
	buffer.WriteString(msg)
	buffer.WriteString(colors.ColorResetEscape)
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
		// MAke the key colored:
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

func (*ColorfulLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buffer := bytes.Buffer{}

	// Add the timestamp:
	buffer.WriteString(colorEscape(entry.Level))
	formatTimestamp(&buffer, entry.Time)
	buffer.WriteByte(' ')

	// Add the symbol:
	buffer.WriteString(symbolTable[entry.Level])
	buffer.WriteString(colors.ColorResetEscape)

	// Add the actual message:
	buffer.WriteByte(' ')
	buffer.WriteString(entry.Message)

	// Add the fields, if any:
	if len(entry.Data) > 0 {
		formatFields(&buffer, entry)
	}

	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

var LogLevelToFunc = map[logrus.Level]func(args ...interface{}){
	logrus.DebugLevel: logrus.Debug,
	logrus.InfoLevel:  logrus.Info,
	logrus.WarnLevel:  logrus.Warn,
	logrus.ErrorLevel: logrus.Error,
	logrus.FatalLevel: logrus.Fatal,
}

type LogWriter struct {
	Level logrus.Level
}

func (l *LogWriter) Write(buf []byte) (int, error) {
	fn, ok := LogLevelToFunc[l.Level]
	if !ok {
		logrus.Fatal("LogWriter: Bad loglevel passed.")
	} else {
		msg := string(buf)
		fn(strings.Trim(msg, "\n\r "))
	}

	return len(buf), nil
}
