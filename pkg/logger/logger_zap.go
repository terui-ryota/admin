package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/terui-ryota/admin/consts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger struct {
	zap   *zap.Logger
	mutex sync.Mutex
}

func init() {
	var err error
	cfg := fixAndFillZapConfig(defaultZapConfig())

	if levelStr, ok := os.LookupEnv("LOGGER_LEVEL"); ok {
		var lvl zapcore.Level
		if err := lvl.Set(levelStr); err != nil {
			panic(fmt.Errorf("failed to set logger level to %s : %w", levelStr, err))
		}
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}

	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.zap, err = cfg.Build(zap.AddCallerSkip(1)); err != nil {
		panic(err)
	}

	addCloser("zap", func() error {
		logger.mutex.Lock()
		defer logger.mutex.Unlock()
		//nolint:wrapcheck
		return logger.zap.Sync()
	})
}

func defaultZapConfig() *zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return &cfg
}

func fixAndFillZapConfig(cfg *zap.Config) zap.Config {
	if cfg == nil {
		cfg = defaultZapConfig()
	}

	if cfg.Encoding == "" {
		cfg.Encoding = "json"
	}
	if cfg.OutputPaths == nil || len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}
	if cfg.ErrorOutputPaths == nil || len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	cfg.EncoderConfig.TimeKey = LogKeyTime
	cfg.EncoderConfig.LevelKey = LogKeyLevel
	cfg.EncoderConfig.NameKey = LogKeyName
	cfg.EncoderConfig.CallerKey = LogKeyCaller
	cfg.EncoderConfig.MessageKey = LogKeyMsg
	cfg.EncoderConfig.StacktraceKey = LogKeyStacktrace
	if cfg.EncoderConfig.LineEnding == "" {
		cfg.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	}
	if cfg.EncoderConfig.EncodeLevel == nil {
		cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	if cfg.EncoderConfig.EncodeTime == nil {
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	}
	if cfg.EncoderConfig.EncodeDuration == nil {
		cfg.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	}
	if cfg.EncoderConfig.EncodeCaller == nil {
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return *cfg
}

type KDSOption struct {
	Stream   string        `yaml:"stream"   json:"stream"`
	Region   string        `yaml:"region"   json:"region"`
	Endpoint string        `yaml:"endpoint" json:"endpoint"`
	Timeout  time.Duration `yaml:"timeout"  json:"timeout"`
}

type LoggerOption struct {
	MineKDS  KDSOption
	OrionKDS KDSOption
}

type Option func(o *LoggerOption) error

func Configure(cfg *zap.Config, opts ...Option) error {
	o := LoggerOption{}
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return fmt.Errorf("failed to configure with option : %w", err)
		}
	}

	c := fixAndFillZapConfig(cfg)
	l, err := c.Build(zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("failed to configure logger : %w", err)
	}

	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.zap = l

	return nil
}

type ZapLogger struct {
	logType string
	logger  *zap.Logger
	ctx     context.Context
}

type contextParamsKey struct{}

func SetupContextFieldsHolder(ctx context.Context, m map[string]string) context.Context {
	ctx = context.WithValue(ctx, contextParamsKey{}, m)
	return ctx
}

// AddContextFields は、contextスコープの値をcontextへ格納します。（grpc method_nameなど）
// interceptorやhtto middlewareなどでリクエストのスコープでのstring値を格納することを想定しています。
// Fieldの値はログの.contextフィールド配下に出力されます。
func AddContextFields(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 != 0 {
		// NOTE: kvが意図しない個数であれば、落とすほどでもないので握りつぶす
		return ctx
	}

	m := map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}

	params := fromCtxScopeFields(ctx)
	if params == nil { // NOTE: ctxにセットされてなければそのままctxを返却する
		return ctx
	}

	for k, v := range m {
		params[k] = v
	}

	return context.WithValue(ctx, contextParamsKey{}, params)
}

func fromCtxScopeFields(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}

	anyParams := ctx.Value(contextParamsKey{})
	if anyParams == nil {
		return nil
	}

	params, ok := anyParams.(map[string]string)
	if !ok {
		return nil
	}

	return params
}

/*
以下のようなログを生成します

	{
		"time": "",
		"level": "",
		"logger": "",
		"log_type": "",
		"log_id": "",
		"service": "",
		"app_meta": {
			"vcs_rev": "",
			"app_ver": ""
		},
		"contents": {
			"hoge": "fuga",
			:
			:
		},
		"context": {
			"piyo": "puga",
			:
			:
		},
		"caller": "",
		"stacktrace": "",
	}
*/
func (l *ZapLogger) prepare(sampled bool, msg string, fields ...zap.Field) *zap.Logger {
	var logId string
	if uid, err := uuid.NewRandom(); err != nil {
		hostname, _ := os.Hostname()
		logId = fmt.Sprintf("uuiderr.%s.%d.%d", hostname, time.Now().Unix(), time.Now().Nanosecond())
	} else {
		logId = uid.String()
	}

	contentsFields :=
		append(make([]zap.Field, 0),
			zap.Namespace(LogKeyContents),
			zap.String(LogKeyMsg, msg),
		)
	contentsFields = append(contentsFields, fields...)

	ll := l.logger
	if l.ctx != nil {
		ll = l.logger.With(tracingField(l.ctx, sampled))
	}

	ll = ll.
		With(zap.String("log_type", l.logType)).
		With(zap.String("log_id", logId)).
		With(zap.String("service", consts.ServiceName)).
		With(zap.Object("app_meta", zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
			inner.AddString("vcs_rev", consts.VcsRevision)
			inner.AddString("app_ver", consts.Version)
			inner.AddString("libgo_ver", consts.LibGoVersion)
			return nil
		})))

	if kv := fromCtxScopeFields(l.ctx); len(kv) != 0 { // NOTE: contextに値があるときのみ出力する
		ll = ll.With(
			zap.Object(LogKeyContext, zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
				for k, v := range kv {
					inner.AddString(k, v)
				}
				return nil
			})))
	}

	return ll.With(contentsFields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	incrError(1)
	l.prepare(true, msg, fields...).Error("")
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	incrWarn(1)
	l.prepare(true, msg, fields...).Warn("")
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.prepare(false, msg, fields...).Info("")
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.prepare(false, msg, fields...).Debug("")
}

func (l *ZapLogger) Errorf(format string, a ...interface{}) {
	incrError(1)
	if ce := l.logger.Check(zapcore.ErrorLevel, format); ce != nil {
		l.prepare(true, fmt.Sprintf(format, a...)).Error("")
	}
}

func (l *ZapLogger) Warnf(format string, a ...interface{}) {
	incrWarn(1)
	if ce := l.logger.Check(zapcore.WarnLevel, format); ce != nil {
		l.prepare(true, fmt.Sprintf(format, a...)).Warn("")
	}
}

func (l *ZapLogger) Infof(format string, a ...interface{}) {
	if ce := l.logger.Check(zapcore.InfoLevel, format); ce != nil {
		l.prepare(false, fmt.Sprintf(format, a...)).Info("")
	}
}

func (l *ZapLogger) Debugf(format string, a ...interface{}) {
	if ce := l.logger.Check(zapcore.DebugLevel, format); ce != nil {
		l.prepare(false, fmt.Sprintf(format, a...)).Debug("")
	}
}

func (l *ZapLogger) Native() *zap.Logger {
	if l.ctx != nil {
		return l.logger.With(tracingField(l.ctx, false))
	}
	return l.logger
}

type ZapLoggerBuilder struct{}

func (b *ZapLoggerBuilder) Default(logType string) Logger {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return &ZapLogger{
		logType: logType,
		logger:  logger.zap,
	}
}

func (b *ZapLoggerBuilder) FromContext(ctx context.Context, logType string) Logger {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return &ZapLogger{
		logType: logType,
		logger:  logger.zap,
		ctx:     ctx,
	}
}

func (b *ZapLoggerBuilder) Named(logType, name string) Logger {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return &ZapLogger{
		logType: logType,
		logger:  logger.zap.Named(name),
	}
}

func (b *ZapLoggerBuilder) NamedFromContext(ctx context.Context, logType, name string) Logger {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return &ZapLogger{
		logType: logType,
		logger:  logger.zap.Named(name),
		ctx:     ctx,
	}
}
