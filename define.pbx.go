package pbx

import (
	. "github.com/obase/api"
	"github.com/obase/log"
)

// 临时适配
var (
	Errorf LoggerFunc = log.Error
	Inforf LoggerFunc = log.Info
	Debugf LoggerFunc = log.Debug
)

func ParsingRequestError(err error, tag string) error {
	return &Response{
		Code: PARSING_REQUEST_ERROR,
		Msg:  err.Error(),
		Tag:  tag,
	}
}
