package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestConfigure(t *testing.T) {
	cfg := zap.NewDevelopmentConfig()
	if err := Configure(&cfg); err != nil {
		panic(err)
	}
	Default().Debug("debug")
	Default().Info("info")
	Default().Warn("warn")
	Default().Error("error")

	Default().Debugf("debug : %+v", `"hello"`)
	Default().Infof("info : %+v", `"hello"`)
	Default().Warnf("warn : %+v", `"hello"`)
	Default().Errorf("error : %v", `"hello"`)

	Named("name").Debug("debug")
	Named("name").Info("info")
	Named("name").Warn("warn")
	Named("name").Error("error")

	FromContext(context.Background()).Debug("debug")
	FromContext(context.Background()).Info("info")
	FromContext(context.Background()).Warn("warn")
	FromContext(context.Background()).Error("error")
}

func TestConfigureWithEmpty(t *testing.T) {
	if err := Configure(&zap.Config{Level: zap.NewAtomicLevel()}); err != nil {
		panic(err)
	}
	Default().Debug("debug")
	Default().Info("info")
	Default().Warn("warn")
	Default().Error("error")

	Default().Debugf("debug : %+v", `"hello"`)
	Default().Infof("info : %+v", `"hello"`)
	Default().Warnf("warn : %+v", `"hello"`)
	Default().Errorf("error : %v", `"hello"`)

	Named("name").Debug("debug")
	Named("name").Info("info")
	Named("name").Warn("warn")
	Named("name").Error("error")

	FromContext(context.Background()).Debug("debug")
	FromContext(context.Background()).Info("info")
	FromContext(context.Background()).Warn("warn")
	FromContext(context.Background()).Error("error")
}
