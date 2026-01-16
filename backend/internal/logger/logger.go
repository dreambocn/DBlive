// 日志封装
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zap.Field

func New(level string) (*zap.Logger, error) {
	// 使用生产配置并解析日志级别
	cfg := zap.NewProductionConfig()
	parsed, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(parsed)
	return cfg.Build()
}

func String(key, val string) Field               { return zap.String(key, val) }
func Int(key string, val int) Field              { return zap.Int(key, val) }
func Int64(key string, val int64) Field          { return zap.Int64(key, val) }
func Error(err error) Field                      { return zap.Error(err) }
func Duration(key string, val interface{}) Field { return zap.Any(key, val) }
