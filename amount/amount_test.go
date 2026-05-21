package amount

import (
	"math/big"
	"strings"
	"testing"
)

func TestStringToBigInt(t *testing.T) {
	cases := []struct {
		in   string
		dec  int64
		want string
	}{
		{"0.001", 18, "1000000000000000"},
		{"1", 18, "1000000000000000000"},
		{"1.5", 6, "1500000"},
		{"0", 18, "0"},
		{"0.5", 18, "500000000000000000"},
		{"1234", 18, "1234000000000000000000"},
		{"1.01", 2, "101"},
	}
	for _, c := range cases {
		n, err := StringToBigInt(c.in, c.dec)
		if err != nil {
			t.Fatalf("StringToBigInt(%q): %v", c.in, err)
		}
		if n.String() != c.want {
			t.Fatalf("StringToBigInt(%q)=%s want %s", c.in, n, c.want)
		}
	}
}

func TestStringToBigInt_StrictFormatRejected(t *testing.T) {
	cases := []struct {
		in  string
		dec int64
	}{
		{"0.0", 18},
		{"0.00", 18},
		{"00", 18},
		{"005", 18},
		{"1.10", 2},
		{"1.0", 2},
		{"0.0010", 18},
		{".5", 18},
		{"5.", 18},
		{" 1", 18},
		{"1 ", 18},
	}
	for _, c := range cases {
		if _, err := StringToBigInt(c.in, c.dec); err == nil {
			t.Fatalf("StringToBigInt(%q) should be rejected in strict mode", c.in)
		}
	}
}

func TestStringToBigInt_RejectInvalid(t *testing.T) {
	_, err := StringToBigInt("1e18", 18)
	if err == nil {
		t.Fatal("scientific notation (1e18) should be rejected")
	}

	_, err = StringToBigInt("1E18", 18)
	if err == nil {
		t.Fatal("scientific notation (1E18) should be rejected")
	}
	_, err = StringToBigInt("1.5E+6", 18)
	if err == nil {
		t.Fatal("scientific notation (1.5E+6) should be rejected")
	}

	_, err = StringToBigInt("", 18)
	if err == nil {
		t.Fatal("empty string should be rejected")
	}

	_, err = StringToBigInt("-1", 18)
	if err == nil {
		t.Fatal("negative amount should be rejected")
	}

	_, err = StringToBigInt("-", 18)
	if err == nil {
		t.Fatal("minus sign only should be rejected")
	}

	_, err = StringToBigInt("+1.5", 18)
	if err == nil {
		t.Fatal("plus sign (+1.5) should be rejected")
	}
	_, err = StringToBigInt("+100", 6)
	if err == nil {
		t.Fatal("plus sign (+100) should be rejected")
	}

	_, err = StringToBigInt("0.0000000000000000001", 18)
	if err == nil {
		t.Fatal("excess fraction digits should be rejected (no silent truncate)")
	}

	_, err = StringToBigInt("1", MaxDecimal+1)
	if err == nil {
		t.Fatal("decimal above MaxDecimal should be rejected")
	}
}

func TestBigIntToDecimal(t *testing.T) {
	wei, _ := new(big.Int).SetString("3000000000000000", 10)
	got, err := BigIntToDecimal(wei, 18)
	if err != nil {
		t.Fatalf("BigIntToDecimal error: %v", err)
	}
	if got != "0.003" {
		t.Fatalf("got %q want 0.003", got)
	}

	zero := big.NewInt(0)
	got, err = BigIntToDecimal(zero, 18)
	if err != nil {
		t.Fatalf("BigIntToDecimal(0) error: %v", err)
	}
	if got != "0" {
		t.Fatalf("0 should return 0, got %s", got)
	}

	_, err = BigIntToDecimal(nil, 18)
	if err == nil {
		t.Fatal("nil big.Int should return error")
	}

	neg := big.NewInt(-1)
	_, err = BigIntToDecimal(neg, 18)
	if err == nil {
		t.Fatal("negative big.Int should return error")
	}

	whole, _ := new(big.Int).SetString("1000000000000000000", 10)
	got, err = BigIntToDecimal(whole, 18)
	if err != nil {
		t.Fatalf("BigIntToDecimal(1e18) error: %v", err)
	}
	if got != "1" {
		t.Fatalf("1e18 wei should be 1 ETH, got %s", got)
	}
}

func TestRoundtrip(t *testing.T) {
	cases := []string{"0", "1", "0.003", "1.5", "1234.567"}
	for _, s := range cases {
		n, err := StringToBigInt(s, 18)
		if err != nil {
			t.Fatalf("StringToBigInt(%q): %v", s, err)
		}
		out, err := BigIntToDecimal(n, 18)
		if err != nil {
			t.Fatalf("BigIntToDecimal after %q: %v", s, err)
		}
		if out != s {
			t.Fatalf("roundtrip %q -> %q", s, out)
		}
	}
}

func TestHumanToChainUnit(t *testing.T) {
	_, err := HumanToChainUnit("0.0000000000000000001", 18)
	if err == nil {
		t.Fatal("want error for excess decimals")
	}

	val, err := HumanToChainUnit("100", 0)
	if err != nil {
		t.Fatalf("decimal=0 should allow integer: %v", err)
	}
	if val.String() != "100" {
		t.Fatalf("decimal=0: 100 should be 100, got %s", val)
	}

	_, err = HumanToChainUnit("1.1", 0)
	if err == nil {
		t.Fatal("decimal=0 should reject fraction")
	}

	_, err = HumanToChainUnit("1", -1)
	if err == nil {
		t.Fatal("negative decimal should be rejected")
	}

	_, err = HumanToChainUnit("1", MaxDecimal+1)
	if err == nil {
		t.Fatal("decimal above MaxDecimal should be rejected")
	}

	_, err = HumanToChainUnit("-0.1", 18)
	if err == nil {
		t.Fatal("negative amount should be rejected")
	}

	_, err = HumanToChainUnit("", 18)
	if err == nil {
		t.Fatal("empty amount should be rejected")
	}
}

func TestSumHumanTotal(t *testing.T) {
	human, chain, err := SumHumanTotal([]string{"0.001", "0.002"}, 18)
	if err != nil {
		t.Fatal(err)
	}
	if human != "0.003" || chain.String() != "3000000000000000" {
		t.Fatalf("human=%s chain=%s", human, chain)
	}

	human, chain, err = SumHumanTotal([]string{}, 18)
	if err != nil {
		t.Fatalf("empty amounts should not error: %v", err)
	}
	if human != "0" || chain.String() != "0" {
		t.Fatalf("empty amounts should be 0, got human=%s chain=%s", human, chain)
	}

	_, _, err = SumHumanTotal([]string{"0.001", "-0.002"}, 18)
	if err == nil {
		t.Fatal("negative amount in list should be rejected")
	}

	_, _, err = SumHumanTotal([]string{"1", "1.2.3", "2"}, 18)
	if err == nil {
		t.Fatal("invalid amount (1.2.3) in middle of list should propagate error")
	}

	_, _, err = SumHumanTotal([]string{"1", "1.10"}, 2)
	if err == nil {
		t.Fatal("strict format in list should be rejected")
	}
}

func TestFormatErrors(t *testing.T) {
	_, err := StringToBigInt("1.2.3", 18)
	if err == nil {
		t.Fatal("multiple dots should error")
	}

	_, err = StringToBigInt("1,000", 18)
	if err == nil {
		t.Fatal("comma should be rejected")
	}
}

func TestChainUnitString(t *testing.T) {
	if ChainUnitString(nil) != "0" {
		t.Fatal("nil should return 0")
	}
	if ChainUnitString(big.NewInt(42)) != "42" {
		t.Fatal("42 should return 42")
	}
}

func TestHumanToChainUnit_DecimalZeroRejectsDot(t *testing.T) {
	_, err := HumanToChainUnit("1.0", 0)
	if err == nil {
		t.Fatal("decimal=0 with dot should error")
	}
}

func TestStringToBigInt_FractionExactLength(t *testing.T) {
	// 小数位恰好等于 decimal（边界内）
	in := "9." + strings.Repeat("9", 6)
	n, err := StringToBigInt(in, 6)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := new(big.Int).SetString(strings.ReplaceAll(in, ".", ""), 10)
	if n.Cmp(want) != 0 {
		t.Fatalf("got %s want %s", n, want)
	}
}
