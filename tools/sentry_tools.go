package tools

import (
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"time"
)

// InitSentry .-

func InitSentry(environment string, dsn string, tracesSampleRate float64, version string) {
	err := sentry.Init(sentry.ClientOptions{
		Environment:      environment,
		Dsn:              dsn,
		TracesSampleRate: tracesSampleRate,
		Release:          version,
	})

	if err != nil {
		LogError(err)
	}
	defer sentry.Flush(time.Second * 5)
}

func LogError(err error) {
	logger := log.New(os.Stdout, "ERROR: ", log.LstdFlags)
	logger.Println(err)
	sentry.CaptureException(err)
}
