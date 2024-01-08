package main

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogger(config *Config) {
	level, err := zap.ParseAtomicLevel(config.Log.Level)
	if err != nil {
		log.Fatalln(err)
	}

	c := zap.NewProductionConfig()
	c.Level = level
	c.Encoding = "console"
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.OutputPaths = config.Log.Output

	logger, err := c.Build()
	if err != nil {
		log.Fatalln(err)
	}
	zap.ReplaceGlobals(logger)
}

// writerWarpper warps zap.Logger implements io.Writer
type writerWarpper struct {
	logger *zap.Logger
}

func (gw *writerWarpper) Write(p []byte) (n int, err error) {
	err = gw.logger.Core().Write(zapcore.Entry{
		Time:    time.Now(),
		Message: string(p),
	}, nil)
	return len(p), err
}
