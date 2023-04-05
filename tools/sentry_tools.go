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
	// ERROR: 라는 prefix 와 날짜/시간 (표준플래그)를 왼쪽에 고정해서 출력
	logger := log.New(os.Stdout, "ERROR: ", log.LstdFlags)
	logger.Println(err)
	sentry.CaptureException(err)
}
