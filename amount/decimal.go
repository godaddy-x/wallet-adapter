// Package amount 区块链资产金额换算：人类可读十进制字符串 ↔ 链上最小单位大整数
//
// 约束规则（严格模式）：
//   - 仅支持非负数（>= 0），负数一律返回 error
//   - 仅支持标准数字格式：0 / 1 / 0.1 / 1.01 / 1111
//   - 禁止：逗号、科学计数法（e/E）、负号（-）、加号（+）
//   - 禁止：无前导零（.5）、尾随小数点（5.）
//   - 禁止：整数前导零（005 / 0123 / 00）
//   - 禁止：小数末尾零（1.0 / 1.10 / 123.450）
//   - 允许：0.1 / 0.12 / 0.999（正确格式）
//   - 唯一零值：0
//   - 小数超长直接报错，不静默截断
//   - decimal 范围：[0, 78]
package amount

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
)

const (
	decimalBase     = 10
	MaxDecimal      = 78
	MaxAmountDigits = 1000
)

var pow10Cache sync.Map

// isAllZeros 检查字符串是否全部由字符 '0' 组成
func isAllZeros(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

// pow10 返回 10^decimal 的副本，使用 int64 确保 sync.Map 缓存键类型绝对统一
func pow10(decimal int64) (*big.Int, error) {
	if decimal < 0 || decimal > MaxDecimal {
		return nil, fmt.Errorf("decimal out of range: [0, %d], got %d", MaxDecimal, decimal)
	}
	key := decimal

	if val, ok := pow10Cache.Load(key); ok {
		return new(big.Int).Set(val.(*big.Int)), nil
	}

	res := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimal), nil)
	actual, _ := pow10Cache.LoadOrStore(key, new(big.Int).Set(res))
	return new(big.Int).Set(actual.(*big.Int)), nil
}

// StringToBigInt 唯一严格校验入口
func StringToBigInt(amountStr string, decimal int64) (*big.Int, error) {
	if amountStr == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	// ==============================
	// 严格白名单校验：只允许 '0'-'9' 和 '.'
	// ==============================
	hasDot := false
	for i, c := range amountStr {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' {
			// 检查是否出现多个小数点
			if hasDot {
				return nil, fmt.Errorf("invalid amount: multiple dots")
			}
			hasDot = true
			continue
		}
		// 只要不是数字或小数点，直接报错（包含空格、符号、字母等）
		return nil, fmt.Errorf("invalid character at position %d: '%c'", i, c)
	}

	// 小数精度范围
	if decimal < 0 || decimal > MaxDecimal {
		return nil, fmt.Errorf("decimal out of range: [0, %d], got %d", MaxDecimal, decimal)
	}

	// 拆分整数与小数
	parts := strings.Split(amountStr, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount: multiple dots")
	}

	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// 禁止无前导零（如 .5）
	if strings.HasPrefix(amountStr, ".") {
		return nil, fmt.Errorf("invalid amount: missing leading zero before dot")
	}

	// 禁止尾随小数点（如 123.）
	if len(parts) == 2 && fracPart == "" {
		return nil, fmt.Errorf("invalid amount: trailing dot not allowed")
	}

	// ==============================
	// 显式零值处理：唯一合法零值是 "0"
	// ==============================
	if amountStr == "0" {
		return big.NewInt(0), nil
	}
	// 拦截 "0.0", "0.00" 等零值变体
	if strings.HasPrefix(amountStr, "0.") && isAllZeros(fracPart) {
		return nil, fmt.Errorf("invalid zero format: use '0' instead of %s", amountStr)
	}

	// ==============================
	// 严格模式：禁止整数部分前导零（如 005）
	// ==============================
	if len(intPart) > 1 && intPart[0] == '0' {
		return nil, fmt.Errorf("invalid amount: leading zeros not allowed")
	}

	// ==============================
	// 严格模式：禁止小数末尾零（如 1.10）
	// ==============================
	if len(fracPart) > 0 && strings.HasSuffix(fracPart, "0") {
		return nil, fmt.Errorf("invalid amount: trailing zeros in fraction not allowed")
	}

	// 小数位数溢出 (注意 len 返回 int，需转换为 int64 比较)
	if int64(len(fracPart)) > decimal {
		return nil, fmt.Errorf("decimal overflow: max %d places, got %d", decimal, len(fracPart))
	}

	// 整数长度限制
	if len(intPart) > MaxAmountDigits {
		return nil, fmt.Errorf("integer part too long: max %d digits, got %d", MaxAmountDigits, len(intPart))
	}

	// 拼接计算 (strings.Repeat 需要 int 类型)
	fracPadded := fracPart + strings.Repeat("0", int(decimal)-len(fracPart))
	combined := intPart + fracPadded
	combined = strings.TrimLeft(combined, "0")
	if combined == "" {
		return big.NewInt(0), nil
	}

	num := new(big.Int)
	if _, ok := num.SetString(combined, decimalBase); !ok {
		return nil, fmt.Errorf("invalid amount number")
	}

	return num, nil
}

// BigIntToDecimal 链上 → 可读
func BigIntToDecimal(b *big.Int, decimal int64) (string, error) {
	if b == nil {
		return "", fmt.Errorf("big.Int is nil")
	}
	if b.Sign() < 0 {
		return "", fmt.Errorf("amount cannot be negative")
	}
	if b.Sign() == 0 {
		return "0", nil
	}

	if decimal < 0 || decimal > MaxDecimal {
		return "", fmt.Errorf("decimal out of range: [0, %d], got %d", MaxDecimal, decimal)
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

	// 小数左侧补 0（例如 0.001 的 rem 为 1，需补成 001）
	fracStr := rem.String()
	if pad := int(decimal) - len(fracStr); pad > 0 {
		fracStr = strings.Repeat("0", pad) + fracStr
	}

	// 严格模式：输出时去除末尾零，保证与输入格式严格对称
	fracStr = strings.TrimRight(fracStr, "0")
	if fracStr == "" {
		return intStr, nil
	}

	return intStr + "." + fracStr, nil
}

// ChainUnitString 链上最小单位字符串
func ChainUnitString(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

// HumanToChainUnit 业务入口
func HumanToChainUnit(amount string, decimal int64) (*big.Int, error) {
	if amount == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	if decimal == 0 && strings.Contains(amount, ".") {
		return nil, fmt.Errorf("decimal=0 does not allow fraction")
	}

	return StringToBigInt(amount, decimal)
}

// HumanToChainUnitString 合约 ABI 专用
func HumanToChainUnitString(amount string, decimal int64) (string, error) {
	num, err := HumanToChainUnit(amount, decimal)
	if err != nil {
		return "", err
	}
	return num.String(), nil
}

// SumHumanTotal 多笔金额求和
func SumHumanTotal(amounts []string, decimal int64) (human string, chainUnit *big.Int, err error) {
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
