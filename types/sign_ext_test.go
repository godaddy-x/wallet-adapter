package types_test

import (
	"testing"

	"github.com/godaddy-x/wallet-adapter/types"
)

func TestBuildAndParseSignExt(t *testing.T) {
	raw, err := types.BuildSignExtJSON(map[string]string{
		types.SignExtKeyChainID:    "31337",
		types.SignExtKeySignScheme: types.SignExtSchemeEIP155,
	})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := types.ParseSignExt(raw)
	if err != nil {
		t.Fatal(err)
	}
	chainID, err := types.SignExtChainID(ext)
	if err != nil || chainID != 31337 {
		t.Fatalf("chainID got %d err %v", chainID, err)
	}
	if types.SignExtScheme(ext) != types.SignExtSchemeEIP155 {
		t.Fatal("scheme mismatch")
	}
}
