package amount

import (
	"math/big"
	"testing"
)

func TestStringToBigInt(t *testing.T) {
	cases := []struct {
		in   string
		dec  int32
		want string
	}{
		{"0.001", 18, "1000000000000000"},
		{"1", 18, "1000000000000000000"},
		{"1.5", 6, "1500000"},
		{"0.0010", 18, "1000000000000000"},
		{"0", 18, "0"},           // 边界：纯零
		{"0.0", 18, "0"},         // 边界：小数零
		{"0.5", 18, "500000000000000000"}, // 边界：前导零小数（正确格式）
		{"1234", 18, "1234000000000000000000"}, // 边界：无小数点整数
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

func TestStringToBigInt_RejectInvalid(t *testing.T) {
	// 科学计数法（小写 e）
	_, err := StringToBigInt("1e18", 18)
	if err == nil {
		t.Fatal("scientific notation (1e18) should be rejected")
	}

	// 🔍 盲区 1：科学计数法（大写 E）
	_, err = StringToBigInt("1E18", 18)
	if err == nil {
		t.Fatal("scientific notation (1E18) should be rejected")
	}
	_, err = StringToBigInt("1.5E+6", 18)
	if err == nil {
		t.Fatal("scientific notation (1.5E+6) should be rejected")
	}

	// 空字符串
	_, err = StringToBigInt("", 18)
	if err == nil {
		t.Fatal("empty string should be rejected")
	}

	// 负数
	_, err = StringToBigInt("-1", 18)
	if err == nil {
		t.Fatal("negative amount should be rejected")
	}

	// 仅负号
	_, err = StringToBigInt("-", 18)
	if err == nil {
		t.Fatal("minus sign only should be rejected")
	}

	// 🔍 盲区 2：正号 +
	_, err = StringToBigInt("+1.5", 18)
	if err == nil {
		t.Fatal("plus sign (+1.5) should be rejected")
	}
	_, err = StringToBigInt("+100", 6)
	if err == nil {
		t.Fatal("plus sign (+100) should be rejected")
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

	// 边界：0
	zero := big.NewInt(0)
	got, err = BigIntToDecimal(zero, 18)
	if err != nil {
		t.Fatalf("BigIntToDecimal(0) error: %v", err)
	}
	if got != "0" {
		t.Fatalf("0 should return 0, got %s", got)
	}

	// 边界：nil 安全
	_, err = BigIntToDecimal(nil, 18)
	if err == nil {
		t.Fatal("nil big.Int should return error")
	}

	// 边界：负数 big.Int
	neg := big.NewInt(-1)
	_, err = BigIntToDecimal(neg, 18)
	if err == nil {
		t.Fatal("negative big.Int should return error")
	}

	// 边界：整数值无余数
	whole, _ := new(big.Int).SetString("1000000000000000000", 10)
	got, err = BigIntToDecimal(whole, 18)
	if err != nil {
		t.Fatalf("BigIntToDecimal(1e18) error: %v", err)
	}
	if got != "1" {
		t.Fatalf("1e18 wei should be 1 ETH, got %s", got)
	}
}

func TestHumanToChainUnit(t *testing.T) {
	// 原有：超长小数应拒绝
	_, err := HumanToChainUnit("0.0000000000000000001", 18)
	if err == nil {
		t.Fatal("want error for excess decimals")
	}

	// decimal = 0 允许整数
	val, err := HumanToChainUnit("100", 0)
	if err != nil {
		t.Fatalf("decimal=0 should allow integer: %v", err)
	}
	if val.String() != "100" {
		t.Fatalf("decimal=0: 100 should be 100, got %s", val)
	}

	// decimal = 0 拒绝小数
	_, err = HumanToChainUnit("1.1", 0)
	if err == nil {
		t.Fatal("decimal=0 should reject fraction")
	}

	// 负数 decimal 报错
	_, err = HumanToChainUnit("1", -1)
	if err == nil {
		t.Fatal("negative decimal should be rejected")
	}

	// 负数金额报错
	_, err = HumanToChainUnit("-0.1", 18)
	if err == nil {
		t.Fatal("negative amount should be rejected")
	}

	// 空字符串报错
	_, err = HumanToChainUnit("", 18)
	if err == nil {
		t.Fatal("empty amount should be rejected")
	}
}

func TestSumHumanTotal(t *testing.T) {
	human, wei, err := SumHumanTotal([]string{"0.001", "0.002"}, 18)
	if err != nil {
		t.Fatal(err)
	}
	if human != "0.003" || wei.String() != "3000000000000000" {
		t.Fatalf("human=%s wei=%s", human, wei)
	}

	// 边界：空数组
	human, wei, err = SumHumanTotal([]string{}, 18)
	if err != nil {
		t.Fatalf("empty amounts should not error: %v", err)
	}
	if human != "0" || wei.String() != "0" {
		t.Fatalf("empty amounts should be 0, got human=%s wei=%s", human, wei)
	}

	// 边界：含负数的数组应报错
	_, _, err = SumHumanTotal([]string{"0.001", "-0.002"}, 18)
	if err == nil {
		t.Fatal("negative amount in list should be rejected")
	}

	// 🔍 盲区 4：数组中间某项非法（错误传播）
	_, _, err = SumHumanTotal([]string{"1", "1.2.3", "2"}, 18)
	if err == nil {
		t.Fatal("invalid amount (1.2.3) in middle of list should propagate error")
	}
}

func TestFormatErrors(t *testing.T) {
	// 多小数点
	_, err := StringToBigInt("1.2.3", 18)
	if err == nil {
		t.Fatal("multiple dots should error")
	}

	// 含逗号
	_, err = StringToBigInt("1,000", 18)
	if err == nil {
		t.Fatal("comma should be rejected")
	}
}
