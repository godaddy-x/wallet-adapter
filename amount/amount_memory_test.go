package amount

import (
	"math/big"
	"sync"
	"testing"
)

// TestPow10_ReturnsIndependentCopy 缓存命中时返回的 *big.Int 必须与缓存隔离，避免污染全局
func TestPow10_ReturnsIndependentCopy(t *testing.T) {
	const d int64 = 18
	a, err := pow10(d)
	if err != nil {
		t.Fatal(err)
	}
	b, err := pow10(d)
	if err != nil {
		t.Fatal(err)
	}
	aOrig := new(big.Int).Set(a)
	bOrig := new(big.Int).Set(b)

	a.Add(a, big.NewInt(1))
	if b.Cmp(bOrig) != 0 {
		t.Fatal("mutating first pow10 result changed second copy")
	}

	// 再次加载应仍等于初始缓存值
	c, err := pow10(d)
	if err != nil {
		t.Fatal(err)
	}
	if c.Cmp(aOrig) != 0 {
		t.Fatal("mutating caller copy corrupted pow10 cache")
	}
}

// TestPow10_CacheEndpoints 边界 decimal 可缓存且数值正确
func TestPow10_CacheEndpoints(t *testing.T) {
	for _, d := range []int64{0, 1, MaxDecimal} {
		p, err := pow10(d)
		if err != nil {
			t.Fatalf("pow10(%d): %v", d, err)
		}
		want := new(big.Int).Exp(big.NewInt(10), big.NewInt(d), nil)
		if p.Cmp(want) != 0 {
			t.Fatalf("pow10(%d)=%s want %s", d, p, want)
		}
	}
}

// TestPow10_Concurrent 并发读缓存不产生 data race（配合 go test -race）
func TestPow10_Concurrent(t *testing.T) {
	const goroutines = 32
	const loops = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < loops; i++ {
				d := int64((seed + i) % (MaxDecimal + 1))
				p, err := pow10(d)
				if err != nil {
					t.Errorf("pow10(%d): %v", d, err)
					return
				}
				want := new(big.Int).Exp(big.NewInt(10), big.NewInt(d), nil)
				if p.Cmp(want) != 0 {
					t.Errorf("pow10(%d) mismatch", d)
					return
				}
				// 偶发变异，验证不影响其他 goroutine 读到的缓存
				if i%17 == 0 {
					p.Add(p, big.NewInt(1))
				}
			}
		}(g)
	}
	wg.Wait()
}

// TestStringToBigInt_ReturnsNewBigInt 每次解析返回独立 *big.Int
func TestStringToBigInt_ReturnsNewBigInt(t *testing.T) {
	a, err := StringToBigInt("1.5", 6)
	if err != nil {
		t.Fatal(err)
	}
	b, err := StringToBigInt("1.5", 6)
	if err != nil {
		t.Fatal(err)
	}
	a.Add(a, big.NewInt(1))
	if b.String() != "1500000" {
		t.Fatalf("mutating a changed b: b=%s", b)
	}
}

// TestSumHumanTotal_DoesNotMutateInputs SumHumanTotal 不应修改调用方传入的 slice 元素语义
func TestSumHumanTotal_ReusesTotalAccumulator(t *testing.T) {
	human, chain, err := SumHumanTotal([]string{"0.001", "0.002"}, 18)
	if err != nil {
		t.Fatal(err)
	}
	chain.Add(chain, big.NewInt(999))
	human2, chain2, err := SumHumanTotal([]string{"0.001", "0.002"}, 18)
	if err != nil {
		t.Fatal(err)
	}
	if human2 != "0.003" || chain2.String() != "3000000000000000" {
		t.Fatalf("prior mutation leaked: human=%s chain=%s", human2, chain2)
	}
	_ = human
}

func TestBigIntToDecimal_DoesNotMutateInput(t *testing.T) {
	b, _ := new(big.Int).SetString("3000000000000000", 10)
	orig := new(big.Int).Set(b)
	if _, err := BigIntToDecimal(b, 18); err != nil {
		t.Fatal(err)
	}
	if b.Cmp(orig) != 0 {
		t.Fatal("BigIntToDecimal mutated input *big.Int")
	}
}

// BenchmarkPow10_CacheHit 热路径：缓存命中
func BenchmarkPow10_CacheHit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := pow10(18); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStringToBigInt_typical 典型 ETH 金额解析
func BenchmarkStringToBigInt_typical(b *testing.B) {
	const s = "1234.567890123456789"
	for i := 0; i < b.N; i++ {
		if _, err := StringToBigInt(s, 18); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBigIntToDecimal_typical 典型链上→可读
func BenchmarkBigIntToDecimal_typical(b *testing.B) {
	v, _ := new(big.Int).SetString("1234567890123456789", 10)
	for i := 0; i < b.N; i++ {
		if _, err := BigIntToDecimal(v, 18); err != nil {
			b.Fatal(err)
		}
	}
}
