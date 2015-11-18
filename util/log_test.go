package util

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestLog(t *testing.T) {
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)

	// Log pretty text
	logrus.SetFormatter(&BrigLogFormatter{})
	// logrus.SetFormatter(&logrus.JSONFormatter{})

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
