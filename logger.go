package logger

import (
	"context"
	"fmt"
	sysLog "log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-haru/log"

	"github.com/go-haru/field"
)

type zapLogger struct {
	syncer context.CancelFunc
	level  log.Level
	logger *zap.Logger
}

func (l *zapLogger) Debug(v ...any) {
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *zapLogger) Debugf(f string, v ...any) {
	l.logger.Debug(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Info(v ...any) {
	l.logger.Info(fmt.Sprint(v...))
}

func (l *zapLogger) Infof(f string, v ...any) {
	l.logger.Info(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Warn(v ...any) {
	l.logger.Warn(fmt.Sprint(v...))
}

func (l *zapLogger) Warnf(f string, v ...any) {
	l.logger.Warn(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Error(v ...any) {
	l.logger.Error(fmt.Sprint(v...))
}

func (l *zapLogger) Errorf(f string, v ...any) {
	l.logger.Error(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Fatal(v ...any) {
	l.logger.Fatal(fmt.Sprint(v...))
}

func (l *zapLogger) Fatalf(f string, v ...any) {
	l.logger.Fatal(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Panic(v ...any) {
	l.logger.Panic(fmt.Sprint(v...))
}

func (l *zapLogger) Panicf(f string, v ...any) {
	l.logger.Panic(fmt.Sprintf(f, v...))
}

func (l *zapLogger) Print(v ...any) {
	l.logger.Info(fmt.Sprint(v...))
}

func (l *zapLogger) Printf(f string, v ...any) {
	l.logger.Info(fmt.Sprintf(f, v...))
}

func (l *zapLogger) WithLevel(level log.Level) log.Logger {
	return &zapLogger{logger: l.logger, level: level, syncer: l.syncer}
}

func (l *zapLogger) With(v ...field.Field) log.Logger {
	var newLogger = l.logger.With(l.zapFields(v)...)
	return &zapLogger{logger: newLogger, level: l.level, syncer: l.syncer}
}

func (l *zapLogger) WithName(name string) log.Logger {
	var newLogger = l.logger.Named(name)
	return &zapLogger{logger: newLogger, level: l.level, syncer: l.syncer}
}

func (l *zapLogger) AddDepth(depth int) log.Logger {
	var newLogger = l.logger.WithOptions(zap.AddCallerSkip(depth))
	return &zapLogger{logger: newLogger, level: l.level, syncer: l.syncer}
}

func (l *zapLogger) Flush() error {
	if l.syncer != nil {
		l.syncer()
	}
	return nil
}

func (l *zapLogger) Standard() *sysLog.Logger {
	var logger, _ = zap.NewStdLogAt(l.logger, l.zapLevel())
	return logger
}

func (l *zapLogger) zapLevel() zapcore.Level {
	switch l.level {
	case log.DebugLevel:
		return zapcore.DebugLevel
	case log.InfoLevel:
		return zapcore.InfoLevel
	case log.WarningLevel:
		return zapcore.WarnLevel
	case log.ErrorLevel:
		return zapcore.ErrorLevel
	case log.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *zapLogger) zapFields(fields field.Fields) []zapcore.Field {
	var converted = make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		switch c := f.Content.(type) {
		case nil:
			continue
		case field.ArrayContent:
			converted = append(converted, zap.Array(f.Key, zapArrayMarshaler{c.Raw()}))
		case field.BinaryContent:
			converted = append(converted, zap.Binary(f.Key, c.Raw()))
		case field.BoolContent:
			converted = append(converted, zap.Bool(f.Key, c.Raw()))
		case field.Complex128Content:
			converted = append(converted, zap.Complex128(f.Key, c.Raw()))
		case field.Complex64Content:
			converted = append(converted, zap.Complex64(f.Key, c.Raw()))
		case field.ErrorContent:
			converted = append(converted, zap.NamedError(f.Key, c.Raw()))
		case field.Float32Content:
			converted = append(converted, zap.Float32(f.Key, c.Raw()))
		case field.Float64Content:
			converted = append(converted, zap.Float64(f.Key, c.Raw()))
		case field.IntContent[int]:
			converted = append(converted, zap.Int(f.Key, c.Raw()))
		case field.IntContent[int8]:
			converted = append(converted, zap.Int8(f.Key, c.Raw()))
		case field.IntContent[int16]:
			converted = append(converted, zap.Int16(f.Key, c.Raw()))
		case field.IntContent[int32]:
			converted = append(converted, zap.Int32(f.Key, c.Raw()))
		case field.IntContent[int64]:
			converted = append(converted, zap.Int64(f.Key, c.Raw()))
		case field.UintContent[uint]:
			converted = append(converted, zap.Uint(f.Key, c.Raw()))
		case field.UintContent[uint8]:
			converted = append(converted, zap.Uint8(f.Key, c.Raw()))
		case field.UintContent[uint16]:
			converted = append(converted, zap.Uint16(f.Key, c.Raw()))
		case field.UintContent[uint32]:
			converted = append(converted, zap.Uint32(f.Key, c.Raw()))
		case field.UintContent[uint64]:
			converted = append(converted, zap.Uint64(f.Key, c.Raw()))
		case field.UintptrContent:
			converted = append(converted, zap.Uintptr(f.Key, c.Raw()))
		case field.StringContent:
			converted = append(converted, zap.String(f.Key, c.Raw()))
		case field.StringerContent:
			converted = append(converted, zap.Stringer(f.Key, c.Raw()))
		case field.TimeContent:
			converted = append(converted, zap.String(f.Key, c.Raw().Format(time.RFC3339Nano)))
		default:
			converted = append(converted, zap.Field{Key: f.Key, Type: zapcore.ReflectType, Interface: f.Data()})
		}
	}
	return converted
}

type zapArrayMarshaler struct {
	content []field.Content
}

func (z zapArrayMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) (err error) {
	for i, content := range z.content {
		switch c := content.(type) {
		case field.ArrayContent:
			if err = encoder.AppendArray(zapArrayMarshaler{c.Raw()}); err != nil {
				return fmt.Errorf("cant encode array item %d: %v", i, err)
			}
		case field.JSONContent:
			if err = encoder.AppendReflected(c.Raw()); err != nil {
				return fmt.Errorf("cant encode array item %d: %v", i, err)
			}
		case field.BinaryContent:
			encoder.AppendString(c.String())
		case field.BoolContent:
			encoder.AppendBool(c.Raw())
		case field.Complex128Content:
			encoder.AppendComplex128(c.Raw())
		case field.Complex64Content:
			encoder.AppendComplex64(c.Raw())
		case field.Float32Content:
			encoder.AppendFloat32(c.Raw())
		case field.Float64Content:
			encoder.AppendFloat64(c.Raw())
		case field.IntContent[int]:
			encoder.AppendInt(c.Raw())
		case field.IntContent[int8]:
			encoder.AppendInt8(c.Raw())
		case field.IntContent[int16]:
			encoder.AppendInt16(c.Raw())
		case field.IntContent[int32]:
			encoder.AppendInt32(c.Raw())
		case field.IntContent[int64]:
			encoder.AppendInt64(c.Raw())
		case field.UintContent[uint]:
			encoder.AppendUint(c.Raw())
		case field.UintContent[uint8]:
			encoder.AppendUint8(c.Raw())
		case field.UintContent[uint16]:
			encoder.AppendUint16(c.Raw())
		case field.UintContent[uint32]:
			encoder.AppendUint32(c.Raw())
		case field.UintContent[uint64]:
			encoder.AppendUint64(c.Raw())
		case field.UintptrContent:
			encoder.AppendUintptr(c.Raw())
		case field.StringContent:
			encoder.AppendString(c.Raw())
		case field.StringerContent:
			encoder.AppendString(c.Raw().String())
		case field.TimeContent:
			encoder.AppendString(c.Raw().Format(time.RFC3339Nano))
		default:
			_ = encoder.AppendReflected(content.Data())
		}
	}
	return nil
}

func New(opts Options) (_ log.Logger, err error) {
	var cores []zapcore.Core
	var syncers []context.CancelFunc
	if cores, syncers, err = build(&opts); err != nil {
		return nil, err
	}
	var underlying = zap.New(zapcore.NewTee(cores...), zap.AddCallerSkip(1), zap.AddCaller(), zap.AddStacktrace(
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= zapcore.DPanicLevel }),
	))
	var syncer = func() {
		_ = underlying.Sync()
		for _, closer := range syncers {
			if closer != nil {
				closer()
			}
		}
	}
	return &zapLogger{logger: underlying, syncer: syncer}, nil
}
