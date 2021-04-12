package util

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

type contextKey string

var (
	// Loglevel is the currently defined loglevel of the application
	Loglevel uint
	// Log is the central logger which is used as default logger
	Log *log.Entry

	hostname, node string
	formatter      = &log.TextFormatter{
		ForceColors: true,
		ForceQuote:  true,
	}

	// context  keys
	ContextKeyLogger    = contextKey("logger")
	ContextKeyRequestID = contextKey("reqID")
)

func init() {
	Loglevel = uint(4)
	log.SetLevel(log.Level(Loglevel))
	log.SetFormatter(formatter)
	Log = NewLogger(nil)
}

func ifEmptySetDash(val string) string {
	if val == "" {
		return "-"
	}
	return val
}

// SetupLog starts the central default logger
func SetupLog(loglevel uint) {
	var err error
	hostname, err = os.Hostname() //  why does os.Getenv("HOSTNAME") not work?
	if err != nil {
		hostname = "-"
	}
	node = ifEmptySetDash(os.Getenv("POD_NODE"))

	if loglevel > 0 {
		Loglevel = loglevel
	}

	log.SetLevel(log.Level(Loglevel))
	log.SetFormatter(formatter)

	Log.Debugf("Creating new central logger (level=%d)", Loglevel)

	Log = NewLogger(log.Fields{
		"hostname": hostname,
		"node":     node,
	})
}

// NewLogger creates a new basic logger with the provided key-values as fields
func NewLogger(fields map[string]interface{}) *log.Entry {
	return log.WithFields(fields)
}

// NewContextLogger creates a new logger with the request-id, hostname and node fields
// and saves it to the provided context
func NewContextLogger(ctx context.Context, requestID string) context.Context {

	Log.Debugf("Creating new context logger (level=%d) (type=json)", Loglevel)

	base := log.New()
	base.SetLevel(log.Level(Loglevel))
	base.SetFormatter(formatter)

	logger := base.WithFields(
		log.Fields{
			"hostname": hostname,
			"node":     node,
			"id":       ifEmptySetDash(requestID),
		},
	)

	loggerCtx := context.WithValue(ctx, ContextKeyLogger, logger)
	loggerCtx = context.WithValue(loggerCtx, ContextKeyRequestID, requestID)
	return loggerCtx
}

func GetValueFromContext(ctx context.Context, key contextKey) string {
	val := ctx.Value(key)
	if val == nil {
		log.Warnf("Failed to get key %s from context", key)
		return ""
	}
	return val.(string)
}

func GetLoggerFromContext(ctx context.Context) *log.Entry {
	entryLogger := ctx.Value(ContextKeyLogger)
	if entryLogger != nil {
		return entryLogger.(*log.Entry)
	}
	return log.NewEntry(log.New())
}
