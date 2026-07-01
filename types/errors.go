package types

import (
	"encoding/json"
	"fmt"
)

// Error code ranges: 2xxx transaction, 3xxx account/address, 4xxx network, 9xxx system; AdapterError is the unified adapter error type.
// Transaction-related error codes (2xxx)
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

// Account/address/contract error codes (3xxx)
const (
	ErrAccountNotFound     = 3001
	ErrAddressNotFound     = 3002
	ErrContractNotFound    = 3003
	ErrAddressEncodeFailed = 3004
	ErrAddressDecodeFailed = 3006
	ErrNonceInvalid        = 3007
)

// Network/node error codes (4xxx)
const (
	ErrCallFullNodeAPIFailed = 4001
	ErrNetworkRequestFailed  = 4002
)

// System/unknown error codes (9xxx)
const (
	ErrUnknownException = 9001
	ErrSystemException  = 9002
)

// AdapterError unified adapter error type with Code and Msg; JSON-serializable.
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
