package pbx

import (
	"context"
	. "github.com/obase/api"
	"github.com/obase/log"
)

/*日志适配方法*/
type LoggerFunc func(ctx context.Context, format string, args ...interface{})

// 临时适配
var Errorf LoggerFunc = func(ctx context.Context, format string, args ...interface{}) {
	log.Error(ctx, format, args...)
}

var Inforf LoggerFunc = func(ctx context.Context, format string, args ...interface{}) {
	log.Info(ctx, format, args...)
}

var Debugf LoggerFunc = func(ctx context.Context, format string, args ...interface{}) {
	log.Debug(ctx, format, args...)
}

func ParsingRequestError(err error, tag string) error {
	return &Response{
		Code: PARSING_REQUEST_ERROR,
		Msg:  err.Error(),
		Tag:  tag,
	}
}
