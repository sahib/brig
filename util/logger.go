package main

import (
	"bytes"
	"fmt"
	"time"
)

import (
	"os"

	logrus "github.com/Sirupsen/logrus"
)

type MooseFormatter struct{}

var symbolTable = map[logrus.Level]string{
	logrus.DebugLevel: "⚙",
	logrus.InfoLevel:  "⚐",
	logrus.WarnLevel:  "⚠",
	logrus.ErrorLevel: "⚡",
	logrus.FatalLevel: "☣",
	logrus.PanicLevel: "☠",
}

var colorTable = map[logrus.Level]int{
	logrus.DebugLevel: 36, // Cyan
	logrus.InfoLevel:  32, // Green
	logrus.WarnLevel:  33, // Yellow
	logrus.ErrorLevel: 31, // Red
	logrus.FatalLevel: 35, // magenta
	logrus.PanicLevel: 41, // BG Red
}

func colorEscape(level logrus.Level) []byte {
	return []byte(fmt.Sprintf("\033[0;%dm", colorTable[level]))
}

var resetEscape = []byte("\033[0m")

func formatColored(buffer *bytes.Buffer, msg string, level logrus.Level) {
	buffer.Write(colorEscape(level))
	buffer.WriteString(msg)
	buffer.Write(resetEscape)
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

func (*MooseFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buffer := bytes.Buffer{}

	// Add the timestamp:
	buffer.Write(colorEscape(entry.Level))
	formatTimestamp(&buffer, entry.Time)
	buffer.WriteByte(' ')

	// Add the symbol:
	buffer.WriteString(symbolTable[entry.Level])
	buffer.Write(resetEscape)

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

func init() {
	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)

	// Log pretty text
	logrus.SetFormatter(&MooseFormatter{})
	// logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"omg":    true,
				"err":    err,
				"number": 100,
			}).Fatal("The ice breaks!")
		}
	}()

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	logrus.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	logrus.WithFields(logrus.Fields{
		"temperature": -4,
	}).Debug("Temperature changes")

	logrus.Error("Stuff!")

	logrus.WithFields(logrus.Fields{
		"animal": "orca",
		"size":   9009,
	}).Panic("It's over 9000!")
}
