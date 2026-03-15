package config

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// KVFromINIContent 从 INI 内容解析指定 section 的 key=value 映射，供 LoadAssetsConfig(MapConfig(kv)) 使用。
func KVFromINIContent(content, section string) (map[string]string, error) {
	return parseINIReader(strings.NewReader(content), section)
}

// KVFromINIFile 从 INI 文件解析指定 section 的 key=value 映射。
func KVFromINIFile(path, section string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseINIReader(f, section)
}

func parseINIReader(r io.Reader, section string) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	kv := make(map[string]string)
	var current string
	target := strings.ToLower(section)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if len(line) >= 2 && line[0] == '[' && line[len(line)-1] == ']' {
			current = strings.TrimSpace(strings.ToLower(line[1 : len(line)-1]))
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(line[:idx]))
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}
		if current == target {
			kv[key] = val
		}
	}
	return kv, scanner.Err()
}
