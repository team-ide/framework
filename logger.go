package framework

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"strings"
	"time"
)

type Logger interface {
	Debug(str string, args ...any)
	Info(str string, args ...any)
	Warn(str string, args ...any)
	Error(str string, args ...any)
	Fatal(str string, args ...any)
}

type defaultLogger struct {
	zapLogger *zap.Logger
}

func (this_ *defaultLogger) Debug(str string, args ...any) {
	strArgs, fields := FormatZapArgs(args...)
	this_.zapLogger.Debug(fmt.Sprintf(str, strArgs...), fields...)
}

func (this_ *defaultLogger) Info(str string, args ...any) {
	strArgs, fields := FormatZapArgs(args...)
	this_.zapLogger.Info(fmt.Sprintf(str, strArgs...), fields...)
}

func (this_ *defaultLogger) Warn(str string, args ...any) {
	strArgs, fields := FormatZapArgs(args...)
	this_.zapLogger.Warn(fmt.Sprintf(str, strArgs...), fields...)
}

func (this_ *defaultLogger) Error(str string, args ...any) {
	strArgs, fields := FormatZapArgs(args...)
	this_.zapLogger.Error(fmt.Sprintf(str, strArgs...), fields...)
}

func (this_ *defaultLogger) Fatal(str string, args ...any) {
	strArgs, fields := FormatZapArgs(args...)
	this_.zapLogger.Fatal(fmt.Sprintf(str, strArgs...), fields...)
}

var (
	DefaultLogger Logger
	Skip1Logger   Logger
	Skip2Logger   Logger
)

func Debug(str string, args ...any) {
	Skip1Logger.Debug(str, args...)
}

func Info(str string, args ...any) {
	Skip1Logger.Info(str, args...)
}

func Warn(str string, args ...any) {
	Skip1Logger.Warn(str, args...)
}

func Error(str string, args ...any) {
	Skip1Logger.Error(str, args...)
}

func Fatal(str string, args ...any) {
	Skip1Logger.Fatal(str, args...)
}

func Debug2(str string, args ...any) {
	Skip2Logger.Debug(str, args...)
}

func Info2(str string, args ...any) {
	Skip2Logger.Info(str, args...)
}

func Warn2(str string, args ...any) {
	Skip2Logger.Warn(str, args...)
}

func Error2(str string, args ...any) {
	Skip2Logger.Error(str, args...)
}

func Fatal2(str string, args ...any) {
	Skip2Logger.Fatal(str, args...)
}

func init() {
	initLogger(newConsoleLogger())
}

type LogConfig struct {
	Console    bool   `json:"console,omitempty" yaml:"console,omitempty"`
	Filename   string `json:"filename,omitempty" yaml:"filename,omitempty"`
	MaxSize    int    `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	MaxAge     int    `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	MaxBackups int    `json:"maxBackups,omitempty" yaml:"maxBackups,omitempty"`
	Level      string `json:"level,omitempty" yaml:"level,omitempty"`
}

func FormatZapArgs(args ...interface{}) (res []any, fields []zap.Field) {
	for _, arg := range args {
		switch tV := arg.(type) {
		case zap.Field:
			fields = append(fields, tV)
			break
		default:
			res = append(res, tV)
			break
		}
	}
	return
}

func LoggerInit(c *LogConfig) {
	oldLogger := DefaultLogger
	oldLogger.Info("logger config change start")
	l := NewZapLogger(c)
	initLogger(l)
	oldLogger.Info("logger config change success")
	l.Info("logger init success")
	return
}

func initLogger(l *zap.Logger) {
	//zapSkip1Logger := NewLoggerByCallerSkip(l, 1)
	zapSkip2Logger := NewLoggerByCallerSkip(l, 2)
	zapSkip3Logger := NewLoggerByCallerSkip(l, 3)

	DefaultLogger = &defaultLogger{zapLogger: l}
	Skip1Logger = &defaultLogger{zapLogger: zapSkip2Logger}
	Skip2Logger = &defaultLogger{zapLogger: zapSkip3Logger}
	return
}

// NewZapLogger creator a new zap logger
// hook {Filename, Maxsize(megabytes), MaxBackups, MaxAge(days)}
// level zap.Level { DebugLevel, InfoLevel, WarnLevel, ErrorLevel, }
func NewZapLogger(c *LogConfig) *zap.Logger {
	var writer io.Writer
	var cLevel string
	if c != nil {
		cLevel = c.Level
	}
	if c != nil && !c.Console {
		writer = &lumberjack.Logger{
			Filename:   c.Filename,
			MaxSize:    c.MaxSize,
			MaxAge:     c.MaxAge,
			MaxBackups: c.MaxBackups,
			Compress:   true,
		}
	} else {
		writer = os.Stdout
	}
	var level zapcore.Level
	switch cLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.DebugLevel
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		zapcore.AddSync(writer),
		zap.NewAtomicLevelAt(level),
	)
	res := zap.New(
		core,
		// 表示 输出 文件名 以及 行号
		zap.AddCaller(),
		// 表示 输出 堆栈跟踪 传入 level 表示 在哪个级别下输出
		zap.AddStacktrace(zapcore.ErrorLevel),
		//zap.AddCallerSkip(0),
	)

	return res
}

func newConsoleLogger() *zap.Logger {
	var level = zapcore.DebugLevel
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)
	caller := zap.AddCaller()

	return zap.New(core, caller,
		// 表示 输出 堆栈跟踪 传入 level 表示 在哪个级别下输出
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "S",
		//TraceKey:         "trackId",
		ConsoleSeparator: "] [",
		LineEnding:       "]\n",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("[2006-01-02 15:04:05.000"))
		},
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			if l == zapcore.InfoLevel || l == zapcore.WarnLevel {
				enc.AppendString(l.CapitalString() + " ")
			} else {
				enc.AppendString(l.CapitalString())
			}
		},
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendFloat64(float64(d) / float64(time.Second))
		},
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			str := caller.TrimmedPath()
			method := caller.Function
			dot := strings.LastIndex(method, ".")
			if dot > 0 {
				method = method[dot+1:]
				index := strings.LastIndex(str, ":")
				if index > 0 {
					str = str[0:index] + ":" + method + str[index:]
				} else {
					str += ":" + method
				}
			}
			enc.AppendString(str)
		},
		//EncodeName: func(name string, enc zapcore.PrimitiveArrayEncoder) {
		//	enc.AppendString("[" + name + "]")
		//},
	}
}

// NewLoggerByCallerSkip 跳过的调用方数量
// skip = 1 表示 输出的 文件名 行号等 为上层方法
func NewLoggerByCallerSkip(l *zap.Logger, skip int) *zap.Logger {
	logger := zap.New(
		l.Core(),
		GetDefaultOptions(skip)...,
	)
	return logger
}

func GetDefaultOptions(skip int) (options []zap.Option) {
	options = append(options,
		// 表示 输出 文件名 以及 行号
		zap.AddCaller(),
		// 表示 输出 堆栈跟踪 传入 level 表示 在哪个级别下输出
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(skip),
	)
	return
}
