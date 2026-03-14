package types

// BalanceModelType 余额模型（按地址或按账户）。
type BalanceModelType uint32

const (
	BalanceModelTypeAddress BalanceModelType = 0
	BalanceModelTypeAccount BalanceModelType = 1
)

// SymbolInfo 链/币种信息接口（Symbol、Decimal、CurveType、FullName、BalanceModelType）。
type SymbolInfo interface {
	Symbol() string
	Decimal() int32
	CurveType() uint32
	FullName() string
	BalanceModelType() BalanceModelType
}

// SymbolInfoBase SymbolInfo 的默认空实现，供 ChainAdapterBase 嵌入。
type SymbolInfoBase struct{}

func (SymbolInfoBase) Symbol() string                    { return "" }
func (SymbolInfoBase) Decimal() int32                    { return 0 }
func (SymbolInfoBase) CurveType() uint32                  { return 0 }
func (SymbolInfoBase) FullName() string                  { return "" }
func (SymbolInfoBase) BalanceModelType() BalanceModelType { return BalanceModelTypeAddress }
