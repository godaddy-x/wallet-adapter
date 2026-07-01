// Package amount blockchain asset amount conversion: human-readable decimal string ↔ on-chain smallest-unit big integer
//
// Constraints (strict mode):
//   - non-negative only (>= 0); negatives always return error
//   - standard numeric formats only: 0 / 1 / 0.1 / 1.01 / 1111
//   - rejected: commas, scientific notation (e/E), minus (-), plus (+)
//   - rejected: missing leading zero (.5), trailing decimal point (5.)
//   - rejected: integer leading zeros (005 / 0123 / 00)
//   - rejected: trailing fractional zeros (1.0 / 1.10 / 123.450)
//   - allowed: 0.1 / 0.12 / 0.999 (valid formats)
//   - sole zero value: 0
//   - excess fractional digits return error; no silent truncation
//   - decimal range: [0, 78]
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

// isAllZeros checks whether the string consists entirely of '0' characters
func isAllZeros(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

// pow10 returns a copy of 10^decimal; uses int64 keys so sync.Map cache key types stay uniform
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

// StringToBigInt sole strict validation entry point
func StringToBigInt(amountStr string, decimal int64) (*big.Int, error) {
	if amountStr == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	// ==============================
	// Strict whitelist: only '0'-'9' and '.'
	// ==============================
	hasDot := false
	for i, c := range amountStr {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' {
			// reject multiple decimal points
			if hasDot {
				return nil, fmt.Errorf("invalid amount: multiple dots")
			}
			hasDot = true
			continue
		}
		// reject anything that is not a digit or dot (spaces, signs, letters, etc.)
		return nil, fmt.Errorf("invalid character at position %d: '%c'", i, c)
	}

	// decimal precision range
	if decimal < 0 || decimal > MaxDecimal {
		return nil, fmt.Errorf("decimal out of range: [0, %d], got %d", MaxDecimal, decimal)
	}

	// split integer and fractional parts
	parts := strings.Split(amountStr, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount: multiple dots")
	}

	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// reject missing leading zero (e.g. .5)
	if strings.HasPrefix(amountStr, ".") {
		return nil, fmt.Errorf("invalid amount: missing leading zero before dot")
	}

	// reject trailing decimal point (e.g. 123.)
	if len(parts) == 2 && fracPart == "" {
		return nil, fmt.Errorf("invalid amount: trailing dot not allowed")
	}

	// ==============================
	// Explicit zero handling: sole valid zero is "0"
	// ==============================
	if amountStr == "0" {
		return big.NewInt(0), nil
	}
	// reject "0.0", "0.00", etc.
	if strings.HasPrefix(amountStr, "0.") && isAllZeros(fracPart) {
		return nil, fmt.Errorf("invalid zero format: use '0' instead of %s", amountStr)
	}

	// ==============================
	// Strict mode: reject integer leading zeros (e.g. 005)
	// ==============================
	if len(intPart) > 1 && intPart[0] == '0' {
		return nil, fmt.Errorf("invalid amount: leading zeros not allowed")
	}

	// ==============================
	// Strict mode: reject trailing fractional zeros (e.g. 1.10)
	// ==============================
	if len(fracPart) > 0 && strings.HasSuffix(fracPart, "0") {
		return nil, fmt.Errorf("invalid amount: trailing zeros in fraction not allowed")
	}

	// fractional digit overflow (len returns int; cast to int64 for comparison)
	if int64(len(fracPart)) > decimal {
		return nil, fmt.Errorf("decimal overflow: max %d places, got %d", decimal, len(fracPart))
	}

	// integer length limit
	if len(intPart) > MaxAmountDigits {
		return nil, fmt.Errorf("integer part too long: max %d digits, got %d", MaxAmountDigits, len(intPart))
	}

	// concatenate (strings.Repeat requires int)
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

// BigIntToDecimal on-chain -> human-readable
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

	// pad fractional part on the left (e.g. rem=1 for 0.001 becomes 001)
	fracStr := rem.String()
	if pad := int(decimal) - len(fracStr); pad > 0 {
		fracStr = strings.Repeat("0", pad) + fracStr
	}

	// strict mode: strip trailing zeros on output for symmetry with input format
	fracStr = strings.TrimRight(fracStr, "0")
	if fracStr == "" {
		return intStr, nil
	}

	return intStr + "." + fracStr, nil
}

// ChainUnitString decimal string of smallest on-chain unit
func ChainUnitString(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

// HumanToChainUnit business entry point
func HumanToChainUnit(amount string, decimal int64) (*big.Int, error) {
	if amount == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	if decimal == 0 && strings.Contains(amount, ".") {
		return nil, fmt.Errorf("decimal=0 does not allow fraction")
	}

	return StringToBigInt(amount, decimal)
}

// HumanToChainUnitString for contract ABI use
func HumanToChainUnitString(amount string, decimal int64) (string, error) {
	num, err := HumanToChainUnit(amount, decimal)
	if err != nil {
		return "", err
	}
	return num.String(), nil
}

// SumHumanTotal sums multiple human-readable amounts
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
