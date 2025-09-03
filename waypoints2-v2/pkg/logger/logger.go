package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Logger
)

func init() {

	// Create a new log instance
	Log = logrus.New()

	// Enable colors for console output
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 MST",
		PadLevelText:    true,
	})

	Log.SetLevel(logrus.TraceLevel)

	// Create a log file
	file, err := os.OpenFile("waypoints-v2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.Warn("Failed to log to file, using default stderr")
	}
}
