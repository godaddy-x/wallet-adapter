package amount

import (
	"math/big"
	"strings"
	"testing"
)

func repeatChar(c byte, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(string(c), n)
}

func TestStringToBigInt_DecimalBoundaries(t *testing.T) {
	if _, err := StringToBigInt("1", -1); err == nil {
		t.Fatal("decimal -1 should error")
	}
	if _, err := StringToBigInt("1", MaxDecimal+1); err == nil {
		t.Fatal("decimal above MaxDecimal should error")
	}

	// decimal=0: integers only
	n, err := StringToBigInt("42", 0)
	if err != nil || n.String() != "42" {
		t.Fatalf("decimal=0 integer: got %v %v", n, err)
	}
	if _, err = StringToBigInt("1.1", 0); err == nil {
		t.Fatal("decimal=0 with fraction should error (overflow)")
	}

	// decimal=MaxDecimal: fractional digits exactly at limit
	frac := repeatChar('9', int(MaxDecimal))
	in := "1." + frac
	n, err = StringToBigInt(in, MaxDecimal)
	if err != nil {
		t.Fatalf("max fraction at MaxDecimal: %v", err)
	}
	want, _ := new(big.Int).SetString("1"+frac, 10)
	if n.Cmp(want) != 0 {
		t.Fatalf("got %s want %s", n, want)
	}

	// fractional digits = decimal+1 should be rejected
	over := "0." + repeatChar('1', int(MaxDecimal)+1)
	if _, err = StringToBigInt(over, MaxDecimal); err == nil {
		t.Fatal("fraction len > decimal should error")
	}

	// fractional digits exactly = decimal (non-MaxDecimal case)
	in18 := "0." + repeatChar('1', 18)
	n, err = StringToBigInt(in18, 18)
	if err != nil {
		t.Fatalf("18 frac digits: %v", err)
	}
	if _, err = StringToBigInt("0."+repeatChar('1', 19), 18); err == nil {
		t.Fatal("19 frac digits at decimal 18 should error")
	}
	_ = n
}

func TestStringToBigInt_IntegerLengthBoundaries(t *testing.T) {
	okInt := repeatChar('9', MaxAmountDigits)
	n, err := StringToBigInt(okInt, 0)
	if err != nil {
		t.Fatalf("MaxAmountDigits integer should pass: %v", err)
	}
	if len(n.String()) != MaxAmountDigits {
		t.Fatalf("got len %d want %d", len(n.String()), MaxAmountDigits)
	}

	tooLong := repeatChar('9', MaxAmountDigits+1)
	if _, err = StringToBigInt(tooLong, 0); err == nil {
		t.Fatal("integer part > MaxAmountDigits should error")
	}
}

func TestStringToBigInt_SmallestNonZero(t *testing.T) {
	cases := []struct {
		dec  int64
		in   string
		want string
	}{
		{18, "0.000000000000000001", "1"},
		{6, "0.000001", "1"},
		{1, "0.1", "1"},
		{78, "0." + repeatChar('0', 77) + "1", "1"},
	}
	for _, c := range cases {
		n, err := StringToBigInt(c.in, c.dec)
		if err != nil {
			t.Fatalf("dec=%d in=%q: %v", c.dec, c.in, err)
		}
		if n.String() != c.want {
			t.Fatalf("dec=%d in=%q got %s want %s", c.dec, c.in, n, c.want)
		}
	}
}

func TestBigIntToDecimal_DecimalBoundaries(t *testing.T) {
	if _, err := BigIntToDecimal(big.NewInt(1), -1); err == nil {
		t.Fatal("decimal -1 should error")
	}
	if _, err := BigIntToDecimal(big.NewInt(1), MaxDecimal+1); err == nil {
		t.Fatal("decimal above MaxDecimal should error")
	}

	got, err := BigIntToDecimal(big.NewInt(99), 0)
	if err != nil || got != "99" {
		t.Fatalf("decimal=0: got %q err=%v", got, err)
	}

	// rem needs left padding: 1 wei @ 18
	got, err = BigIntToDecimal(big.NewInt(1), 18)
	if err != nil || got != "0.000000000000000001" {
		t.Fatalf("smallest display @18: got %q err=%v", got, err)
	}

	// exact division, no remainder
	oneEth, _ := new(big.Int).SetString("1000000000000000000", 10)
	got, err = BigIntToDecimal(oneEth, 18)
	if err != nil || got != "1" {
		t.Fatalf("exact whole: got %q err=%v", got, err)
	}

	// MaxDecimal smallest non-zero unit (on-chain 1)
	got, err = BigIntToDecimal(big.NewInt(1), MaxDecimal)
	want := "0." + repeatChar('0', int(MaxDecimal)-1) + "1"
	if err != nil || got != want {
		t.Fatalf("MaxDecimal smallest: got %q want %q err=%v", got, want, err)
	}
}

func TestBigIntToDecimal_LargeIntegerPart(t *testing.T) {
	// on-chain value = 10*10^78 + 10^77 (human-readable 10.1 @ decimal 78)
	divisor, err := pow10(MaxDecimal)
	if err != nil {
		t.Fatal(err)
	}
	b := new(big.Int).Add(
		new(big.Int).Mul(big.NewInt(10), divisor),
		new(big.Int).Div(divisor, big.NewInt(10)),
	)
	got, err := BigIntToDecimal(b, MaxDecimal)
	if err != nil {
		t.Fatal(err)
	}
	if got != "10.1" {
		t.Fatalf("large int part: got %q want 10.1", got)
	}
}

func TestHumanToChainUnitString(t *testing.T) {
	s, err := HumanToChainUnitString("1.5", 6)
	if err != nil || s != "1500000" {
		t.Fatalf("got %q err=%v", s, err)
	}
	if _, err = HumanToChainUnitString("1.10", 2); err == nil {
		t.Fatal("strict format should error")
	}
}

func TestSumHumanTotal_LargeBatch(t *testing.T) {
	const n = 500
	amounts := make([]string, n)
	for i := range amounts {
		amounts[i] = "0.001"
	}
	human, chain, err := SumHumanTotal(amounts, 18)
	if err != nil {
		t.Fatal(err)
	}
	wantChain, _ := new(big.Int).SetString("500000000000000000", 10) // 0.001 * 500
	if chain.Cmp(wantChain) != 0 {
		t.Fatalf("chain=%s want %s", chain, wantChain)
	}
	if human != "0.5" {
		t.Fatalf("human=%s want 0.5", human)
	}
}

func TestRoundtrip_MaxDecimal(t *testing.T) {
	s := "12.345678901234567890123456789012345678901234567890123456789012345678"
	n, err := StringToBigInt(s, MaxDecimal)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BigIntToDecimal(n, MaxDecimal)
	if err != nil {
		t.Fatal(err)
	}
	if out != s {
		t.Fatalf("roundtrip MaxDecimal: %q -> %q", s, out)
	}
}
