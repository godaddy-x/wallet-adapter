// Package config provides common chain config interfaces and JSON parsing, reused by sub-adapters (e.g. github.com/godaddy-x/wallet-adapter-eth).
package config

import (
	"strconv"
)

// Configer is a read-only config interface for LoadAssetsConfig (e.g. JSON config sections, maps), compatible with common config libraries.
type Configer interface {
	String(key string) string
	Int64(key string) (int64, error)
}

// MapConfig wraps map[string]string as Configer (keys matched case-insensitively).
type MapConfig map[string]string

func (m MapConfig) String(key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func (m MapConfig) Int64(key string) (int64, error) {
	s := m.String(key)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
