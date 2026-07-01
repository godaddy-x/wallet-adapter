package chain

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/scanner"
	"golang.org/x/crypto/ripemd160"
)

// Register/query by symbol: RegAdapter, GetAdapter, GetTransactionDecoder, GetBlockScanner, GetAddressDecoder, GetSmartContractDecoder, ListSymbols.

var (
	adapters   = make(map[string]ChainAdapter)
	adaptersMu sync.RWMutex
)

// RegAdapter registers a chain adapter; symbol is usually uppercase (e.g. BTC, ETH, BTM)
func RegAdapter(symbol string, a ChainAdapter) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if a == nil {
		panic("adapter is nil")
	}
	adapters[symbol] = a
}

// GetAdapter returns the chain adapter for symbol
func GetAdapter(symbol string) (ChainAdapter, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	a, ok := adapters[symbol]
	if !ok {
		return nil, fmt.Errorf("adapter not found for symbol: %s", symbol)
	}
	return a, nil
}

// GetTransactionDecoder returns the TransactionDecoder for symbol
func GetTransactionDecoder(symbol string) (decoder.TransactionDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetTransactionDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support transaction build/broadcast", symbol)
	}
	return d, nil
}

// GetBlockScanner returns the BlockScanner for symbol
func GetBlockScanner(symbol string) (scanner.BlockScanner, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	s := a.GetBlockScanner()
	if s == nil {
		return nil, fmt.Errorf("chain %s does not support block scanner", symbol)
	}
	return s, nil
}

// GetAddressDecoder returns the AddressDecoder for symbol
func GetAddressDecoder(symbol string) (decoder.AddressDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetAddressDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support address decoder", symbol)
	}
	return d, nil
}

// GetSmartContractDecoder returns the SmartContractDecoder for symbol (optional; error if unsupported)
func GetSmartContractDecoder(symbol string) (decoder.SmartContractDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetSmartContractDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support smart contract decoder", symbol)
	}
	return d, nil
}

// ListSymbols returns all registered symbols
func ListSymbols() []string {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	symbols := make([]string, 0, len(adapters))
	for s := range adapters {
		symbols = append(symbols, s)
	}
	return symbols
}

// GenContractID contract ID: SHA256 of symbol_address, then Base64-encoded
func GenContractID(symbol, address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%v_%v", symbol, address)))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// KeyIDVer Base58Check version byte, used by ComputeKeyID
const KeyIDVer = 0x00

// ComputeKeyID computes KeyID from seed: SHA256(seed) -> RIPEMD160 -> Base58Check(KeyIDVer)
func ComputeKeyID(seed []byte) string {
	// Step 1: SHA256(seed)
	h := sha256.Sum256(seed)
	// Step 2: RIPEMD160(SHA256(seed))
	rh := ripemd160.New()
	rh.Write(h[:])
	ripemd160Hash := rh.Sum(nil)
	// Step 3: Base58Check encode (with version byte)
	return base58CheckEncode(KeyIDVer, ripemd160Hash)
}

// base58CheckEncode Base58Check-encodes payload; version is a 1-byte version number
func base58CheckEncode(version byte, payload []byte) string {
	data := make([]byte, 0, 1+len(payload)+4)
	data = append(data, version)
	data = append(data, payload...)
	checksum := sha256.Sum256(data)
	checksum = sha256.Sum256(checksum[:])
	data = append(data, checksum[:4]...)
	return base58Encode(data)
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(data []byte) string {
	// Count leading zeros
	leadingZeros := 0
	for _, b := range data {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Convert big-endian bytes to big integer
	num := new(big.Int).SetBytes(data)

	// Encode in base58
	var result []byte
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Append leading '1's for each leading zero byte
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
