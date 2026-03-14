package types

import (
	"encoding/json"
	"fmt"
)

// 错误码分区：2xxx 交易、3xxx 账户/地址、4xxx 网络、9xxx 系统；AdapterError 为适配器统一错误类型。
// 交易相关错误码（2xxx）
const (
	ErrInsufficientBalanceOfAccount      = 2001
	ErrInsufficientBalanceOfAddress      = 2002
	ErrInsufficientFees                  = 2003
	ErrDustLimit                         = 2004
	ErrCreateRawTransactionFailed        = 2005
	ErrSignRawTransactionFailed          = 2006
	ErrVerifyRawTransactionFailed        = 2007
	ErrSubmitRawTransactionFailed        = 2008
	ErrInsufficientTokenBalanceOfAddress = 2009
)

// 账户/地址/合约等错误码（3xxx）
const (
	ErrAccountNotFound     = 3001
	ErrAddressNotFound     = 3002
	ErrContractNotFound    = 3003
	ErrAddressEncodeFailed = 3004
	ErrAddressDecodeFailed = 3006
	ErrNonceInvalid        = 3007
)

// 网络/节点错误码（4xxx）
const (
	ErrCallFullNodeAPIFailed = 4001
	ErrNetworkRequestFailed  = 4002
)

// 系统/未知错误码（9xxx）
const (
	ErrUnknownException = 9001
	ErrSystemException  = 9002
)

// AdapterError 适配器统一错误类型，带 Code 与 Msg，可序列化为 JSON。
type AdapterError struct {
	Code uint64 `json:"code"`
	Msg  string `json:"msg"`
}

func (e *AdapterError) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Msg)
}

func NewError(code uint64, text string) *AdapterError {
	return &AdapterError{Code: code, Msg: text}
}

func Errorf(code uint64, format string, a ...interface{}) *AdapterError {
	return &AdapterError{Code: code, Msg: fmt.Sprintf(format, a...)}
}

func ConvertError(err error) *AdapterError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AdapterError); ok {
		return ae
	}
	return &AdapterError{Code: ErrUnknownException, Msg: err.Error()}
}

func (e *AdapterError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{"code": e.Code, "msg": e.Msg})
}

func (e *AdapterError) UnmarshalJSON(b []byte) error {
	var m struct {
		Code uint64 `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	e.Code = m.Code
	e.Msg = m.Msg
	return nil
}
