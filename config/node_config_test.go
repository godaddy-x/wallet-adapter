package config

import (
	"testing"
)

func TestParseNodeConfigJSON_verifyAPIs(t *testing.T) {
	raw := `{
  "serverAPI": "https://primary",
  "broadcastAPI": "https://broadcast",
  "chainID": 1,
  "verifyAPIs": ["https://peer-a", "https://peer-b"]
}`
	kv, verify, err := ParseNodeConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if kv.String("serverAPI") != "https://primary" {
		t.Fatalf("serverAPI=%q", kv.String("serverAPI"))
	}
	if kv.String("chainID") != "1" {
		t.Fatalf("chainID=%q", kv.String("chainID"))
	}
	if len(verify) != 2 || verify[0] != "https://peer-a" {
		t.Fatalf("verify=%v", verify)
	}
}

func TestParseNodeConfigJSON_verifyAPIsInvalid(t *testing.T) {
	_, _, err := ParseNodeConfigJSON(`{"serverAPI":"https://x","verifyAPIs":"not-an-array"}`)
	if err == nil {
		t.Fatal("expected error for non-array verifyAPIs")
	}
}
