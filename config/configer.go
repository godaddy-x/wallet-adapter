// Package config 提供链配置的通用接口与 INI 解析，供各子类适配器（如 github.com/godaddy-x/wallet-adapter-eth）复用。
package config

import (
	"strconv"
)

// Configer 供 LoadAssetsConfig 使用的只读配置接口（如 INI 段、map），与常见 config 库兼容。
type Configer interface {
	String(key string) string
	Int64(key string) (int64, error)
}

// MapConfig 将 map[string]string 转为 Configer（key 转小写匹配）。
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
