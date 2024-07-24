package logger

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/terui-ryota/admin/util/random"
	"go.opencensus.io/trace"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Nantes official log type
	LogTypeMine          = "mine"
	LogTypeOrion         = "orion"
	LogTypeElasticSearch = "elastic_search"

	// Application defined log type
	LogTypeApp    = "app"
	LogTypeAccess = "access_log"
)

const (
	LogKeyContents   = "contents"
	LogKeyContext    = "context"
	LogKeyTime       = "time"
	LogKeyLevel      = "level"
	LogKeyName       = "logger"
	LogKeyCaller     = "caller"
	LogKeyMsg        = "msg"
	LogKeyStacktrace = "stacktrace"
)

var (
	closers                = make(map[string]func() error)
	builder  LoggerBuilder = &ZapLoggerBuilder{}
	hostname string
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		Default().Warn("failed to lookup hostname", zap.Error(err))
		hostname = random.String(random.AlphanumericRuneSet, 40)
	}
}

var (
	stackedError = atomic.NewInt64(0)
	incrError    = func(v int) {
		stackedError.Add(int64(v))
		fmt.Println("[LOGGER] incrError is not configured / You should import 'metrics' package in entry point of application.")
	}
	stackedWarn = atomic.NewInt64(0)
	incrWarn    = func(v int) {
		stackedWarn.Add(int64(v))
		fmt.Println("[LOGGER] incrWarn is not configured / You should import 'metrics' package in entry point of application.")
	}
	stackedMine = atomic.NewInt64(0)
	incrMine    = func(v int) {
		stackedMine.Add(int64(v))
		fmt.Println("[LOGGER] incrWarn is not configured / You should import 'metrics' package in entry point of application.")
	}
	stackedOrion = atomic.NewInt64(0)
	incrOrion    = func(v int) {
		stackedOrion.Add(int64(v))
		fmt.Println("[LOGGER] incrWarn is not configured / You should import 'metrics' package in entry point of application.")
	}
)

func SetMetricsReporter(incrErrorFn, incrWarnFn, incrMineFn, incrOrionFn func(v int)) {
	incrError = incrErrorFn
	incrWarn = incrWarnFn
	incrMine = incrMineFn
	incrOrion = incrOrionFn

	incrError(int(stackedError.Load()))
	incrWarn(int(stackedWarn.Load()))
	incrMine(int(stackedMine.Load()))
	incrOrion(int(stackedOrion.Load()))
}

func addCloser(name string, closer func() error) {
	closers[name] = closer
}

func SetLoggerBuilder(lb LoggerBuilder) {
	builder = lb
}

func Close() error {
	em := strings.Builder{}

	for n, closer := range closers {
		if err := closer(); err != nil {
			if em.Len() > 0 {
				em.WriteByte(' ')
			}
			em.WriteString(fmt.Sprintf("Error[%s]=%v", n, err))
		}
	}

	if em.Len() == 0 {
		return nil
	}

	return errors.New(em.String())
}

type (
	LogFunc              func(msg string, fields ...zap.Field)
	LogContextFunc       func(ctx context.Context) LogFunc
	LogFormatFunc        func(format string, a ...interface{})
	LogFormatContextFunc func(ctx context.Context) LogFormatFunc
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Errorf(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
	Native() *zap.Logger
}

type LoggerBuilder interface {
	Default(logType string) Logger
	FromContext(ctx context.Context, logType string) Logger
	Named(logType, name string) Logger
	NamedFromContext(ctx context.Context, logType, name string) Logger
}

func Access(ctx context.Context) Logger {
	return builder.NamedFromContext(ctx, LogTypeAccess, LogTypeAccess)
}

func Default() Logger {
	return builder.Default(LogTypeApp)
}

func Named(name string) Logger {
	return builder.Named(LogTypeApp, name)
}

func FromContext(ctx context.Context) Logger {
	return builder.FromContext(ctx, LogTypeApp)
}

func NamedFromContext(ctx context.Context, name string) Logger {
	return builder.NamedFromContext(ctx, LogTypeApp, name)
}

func tracingField(ctx context.Context, sampled bool) zap.Field {
	span := trace.FromContext(ctx)

	var (
		traceID  string
		spanID   string
		traceURL string
	)

	if span != nil {
		sc := span.SpanContext()
		traceID = sc.TraceID.String()
		spanID = sc.SpanID.String()
		traceURL = fmt.Sprintf(
			"https://ca-media-nantes.datadoghq.com/apm/trace/%d?spanID=%d",
			binary.LittleEndian.Uint64(sc.TraceID[8:]),
			binary.LittleEndian.Uint64(sc.SpanID[:]),
		)
		if !sampled {
			sampled = span.SpanContext().IsSampled()
		}
	}

	return zap.Object("tracing", zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
		inner.AddString("trace_id", traceID)
		inner.AddString("span_id", spanID)
		inner.AddString("trace_url", traceURL)
		inner.AddBool("sampled", sampled)
		return nil
	}))
}
