package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Setup() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	c.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	l, err := c.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l)
	return l, nil
}

func MustSetup() *zap.Logger {
	l, err := Setup()
	if err != nil {
		panic(err)
	}
	return l
}
