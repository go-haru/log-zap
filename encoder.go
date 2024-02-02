package logger

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-haru/log"
)

const (
	FormatJSON = "json"
	FormatText = "text"
)

var mainPath string

func init() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		mainPath = buildInfo.Main.Path + "/"
	}
}

func jsonEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func readableEncoder(enableColor bool, longTime bool) zapcore.Encoder {
	var LevelEncoder zapcore.LevelEncoder
	var NameEncoder zapcore.NameEncoder
	var RFC3339TimeEncoder zapcore.TimeEncoder
	var ShortCallerEncoder zapcore.CallerEncoder
	if enableColor {
		color.NoColor = false
		var BoldCyan = color.New(color.Bold, color.FgCyan)
		var BoldHiBlue = color.New(color.Bold, color.FgHiBlue)
		var HiBlack = color.New(color.FgHiBlack)
		var White = color.New(color.FgWhite)
		var yellow = color.New(color.FgYellow)
		var textHiBlackT = HiBlack.Sprint("T")
		var textYellowLSB = yellow.Sprint("[")
		var textYellowRSB = yellow.Sprint("]")
		LevelEncoder = zapcore.CapitalColorLevelEncoder
		NameEncoder = func(s string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("%s%s%s", textYellowLSB, s, textYellowRSB))
		}
		if longTime {
			RFC3339TimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				str := BoldCyan.Sprint(t.Format("2006-01-02")) + textHiBlackT +
					BoldHiBlue.Sprint(t.Format("15:04:05")) +
					HiBlack.Sprint(t.Format(".000000Z0700"))
				enc.AppendString(str)
			}
		} else {
			RFC3339TimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				str := BoldCyan.Sprint(t.Format("2006-01-02")) + textHiBlackT +
					BoldHiBlue.Sprint(t.Format("15:04:05Z07"))
				enc.AppendString(str)
			}
		}
		ShortCallerEncoder = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			var callerPath = strings.TrimPrefix(caller.File, mainPath)
			callerPath = strings.Replace(callerPath, "@v0.0.0-00010101000000-000000000000", "", 1)
			enc.AppendString(White.Sprintf("%s:%d", callerPath, caller.Line))
		}
	} else {
		NameEncoder = func(s string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("[%s]", s))
		}
		LevelEncoder = zapcore.CapitalLevelEncoder
		if longTime {
			RFC3339TimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02T15:04:05.000000Z0700"))
			}
		} else {
			RFC3339TimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02T15:04:05Z07"))
			}
		}
		ShortCallerEncoder = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			var callerPath = strings.TrimPrefix(caller.File, mainPath)
			callerPath = strings.Replace(callerPath, "@v0.0.0-00010101000000-000000000000", "", 1)
			enc.AppendString(fmt.Sprintf("%s:%d", callerPath, caller.Line))
		}
	}

	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    LevelEncoder,
		EncodeName:     NameEncoder,
		EncodeTime:     RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   ShortCallerEncoder,
	})
}

func newEncoder(opts *Options) zapcore.Encoder {
	switch opts.Format {
	default:
		fallthrough
	case FormatText:
		return readableEncoder(opts.WithColor, opts.LongTime)
	case FormatJSON:
		return jsonEncoder()
	}
}

func zapLevel(level log.Level) zapcore.Level {
	switch level {
	default:
		fallthrough
	case log.DebugLevel:
		return zap.DebugLevel
	case log.InfoLevel:
		return zap.InfoLevel
	case log.WarningLevel:
		return zap.WarnLevel
	case log.ErrorLevel:
		return zap.ErrorLevel
	case log.FatalLevel:
		return zap.FatalLevel
	}
}

func build(opts *Options) (cores []zapcore.Core, closers []context.CancelFunc, err error) {
	var ConsoleInfoCloser, ConsoleErrorCloser func()
	var stdOutSyncer, stdErrSyncer zapcore.WriteSyncer
	if stdOutSyncer, ConsoleInfoCloser, err = zap.Open("stdout"); err != nil {
		return nil, nil, fmt.Errorf("cant init logger console writeSyncer: stdout: %w", err)
	}
	if stdErrSyncer, ConsoleErrorCloser, err = zap.Open("stderr"); err != nil {
		return nil, nil, fmt.Errorf("cant init logger console writeSyncer: stderr: %w", err)
	}
	var minLevel = zap.InfoLevel
	if opts.Level != "" {
		if level, ok := log.ParseLevel(opts.Level); !ok {
			return nil, nil, fmt.Errorf("invalid log level: %q", opts.Level)
		} else {
			minLevel = zapLevel(level)
		}
	}
	var encoder = newEncoder(opts)
	cores = append(cores,
		zapcore.NewCore(encoder, stdOutSyncer,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= minLevel && lvl <= zapcore.WarnLevel }),
		),
		zapcore.NewCore(encoder, stdErrSyncer,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= minLevel && lvl > zapcore.WarnLevel }),
		),
	)
	closers = []context.CancelFunc{
		ConsoleInfoCloser,
		ConsoleErrorCloser,
	}
	return cores, closers, nil
}
