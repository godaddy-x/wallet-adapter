// Package amount 区块链资产金额换算：人类可读十进制字符串 ↔ 链上最小单位大整数
//
// 约束规则（严格模式）：
//   - 仅支持非负数（>= 0），负数一律返回 error
//   - 仅支持标准数字格式：0 / 1 / 0.1 / 1.01 / 1111
//   - 禁止逗号、科学计数法（e/E）、负号（-）、加号（+）
//   - 非法参数直接返回 error，不静默容错、不自动修正
package amount

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
)

const decimalBase = 10

var pow10Cache sync.Map

// pow10 禁止负数精度，缓存键统一 int64，绝对类型安全
func pow10(decimal int32) (*big.Int, error) {
	if decimal < 0 {
		return nil, fmt.Errorf("decimal cannot be negative")
	}
	key := int64(decimal)

	if val, ok := pow10Cache.Load(key); ok {
		return new(big.Int).Set(val.(*big.Int)), nil
	}

	res := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	actual, _ := pow10Cache.LoadOrStore(key, new(big.Int).Set(res))
	return new(big.Int).Set(actual.(*big.Int)), nil
}

// StringToBigInt 将人类可读金额字符串转为链上最小单位大整数。
// 约束：仅支持非负数（>= 0），负数、科学计数法、逗号等均返回 error。
func StringToBigInt(amountStr string, decimal int32) (*big.Int, error) {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	// 严格拦截负号：仅支持 >= 0
	if strings.HasPrefix(amountStr, "-") {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	// 严格拦截正号：禁止 + 号前缀
	if strings.HasPrefix(amountStr, "+") {
		return nil, fmt.Errorf("amount invalid: plus sign not allowed")
	}

	// 拦截科学计数法
	if strings.ContainsAny(amountStr, "eE") {
		return nil, fmt.Errorf("amount invalid: scientific notation not allowed")
	}

	// 严格拦截逗号
	if strings.Contains(amountStr, ",") {
		return nil, fmt.Errorf("amount invalid: comma not allowed")
	}

	parts := strings.Split(amountStr, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount: multiple dots")
	}

	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// 严格模式：拒绝无前导零的小数（如 .5）
	if intPart == "" {
		return nil, fmt.Errorf("invalid amount: missing leading zero before decimal point")
	}

	exp := int(decimal)
	if len(fracPart) > exp {
		fracPart = fracPart[:exp]
	} else {
		fracPart += strings.Repeat("0", exp-len(fracPart))
	}

	combined := strings.TrimLeft(intPart+fracPart, "0")
	if combined == "" {
		return big.NewInt(0), nil
	}

	num := new(big.Int)
	// 消除魔法数字，使用常量 decimalBase
	if _, ok := num.SetString(combined, decimalBase); !ok {
		return nil, fmt.Errorf("invalid amount number")
	}

	return num, nil
}

// BigIntToDecimal 将链上最小单位大整数转为人类可读金额字符串。
// 约束：仅支持非负数（>= 0），负数和 nil 返回 error。
func BigIntToDecimal(b *big.Int, decimal int32) (string, error) {
	if b == nil {
		return "", fmt.Errorf("big.Int is nil")
	}
	if b.Sign() < 0 {
		return "", fmt.Errorf("amount cannot be negative")
	}
	if b.Sign() == 0 {
		return "0", nil
	}

	if decimal < 0 {
		return "", fmt.Errorf("decimal cannot be negative")
	}
	if decimal == 0 {
		return b.String(), nil
	}

	divisor, err := pow10(decimal)
	if err != nil {
		return "", err
	}

	intPart, rem := new(big.Int).QuoRem(b, divisor, new(big.Int))
	intStr := intPart.String()

	if rem.Sign() == 0 {
		return intStr, nil
	}

	fracStr := rem.String()
	if pad := int(decimal) - len(fracStr); pad > 0 {
		fracStr = strings.Repeat("0", pad) + fracStr
	}

	// 修复 1：TrimRight 后只在非空时拼接小数点 + 小数
	fracStr = strings.TrimRight(fracStr, "0")
	if fracStr == "" {
		return intStr, nil
	}
	return intStr + "." + fracStr, nil
}

// ChainUnitString 返回链上最小单位的十进制字符串表示。
// 若 b 为 nil 返回 "0"，不处理负数校验（调用方需确保非负）。
func ChainUnitString(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

// HumanToChainUnit 业务入口：人类可读金额 → 链上最小单位。
// 严格校验：仅支持非负数（>= 0），小数位超长返回 error（不静默截断）。
func HumanToChainUnit(amount string, decimal int32) (*big.Int, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	// 严格拦截负号：仅支持 >= 0
	if strings.HasPrefix(amount, "-") {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	if decimal < 0 {
		return nil, fmt.Errorf("decimal cannot be negative")
	}

	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount format")
	}

	if decimal == 0 && len(parts) == 2 {
		return nil, fmt.Errorf("decimal is 0, fraction not allowed")
	}

	if len(parts) == 2 && len(parts[1]) > int(decimal) {
		return nil, fmt.Errorf("decimal overflow: max %d places", decimal)
	}

	return StringToBigInt(amount, decimal)
}

// HumanToChainUnitString 合约 ABI 专用：人类可读金额 → 链上最小单位字符串。
// 严格校验：仅支持非负数（>= 0）。
func HumanToChainUnitString(amount string, decimal int32) (string, error) {
	num, err := HumanToChainUnit(amount, decimal)
	if err != nil {
		return "", err
	}
	return num.String(), nil
}

// SumHumanTotal 多笔人类可读金额求和，返回人类可读总额和链上总额。
// 约束：每笔金额必须 >= 0，否则返回 error。
func SumHumanTotal(amounts []string, decimal int32) (human string, wei *big.Int, err error) {
	total := new(big.Int)
	for i, s := range amounts {
		val, err := HumanToChainUnit(s, decimal)
		if err != nil {
			return "", nil, fmt.Errorf("amount[%d] invalid: %w", i, err)
		}
		total.Add(total, val)
	}
	human, err = BigIntToDecimal(total, decimal)
	if err != nil {
		return "", nil, err
	}
	return human, total, nil
}
