package ghcpotel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func RedactSchema(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(io.LimitReader(reader, maxExportBytes+1))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	totalBytes := 0
	rows := 0
	for line := 1; scanner.Scan(); line++ {
		rows++
		totalBytes += len(scanner.Bytes()) + 1
		if rows > maxExportRows || totalBytes > maxExportBytes {
			return fmt.Errorf("OTel export exceeds row or byte limit")
		}
		row, err := decodeObject(scanner.Bytes())
		if err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		if err := rejectContentKeys(row); err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		if err := encoder.Encode(schemaOf(row)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func schemaOf(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			result[key] = schemaOf(child)
		}
		return result
	case []any:
		return "array"
	case json.Number:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func rejectContentKeys(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if sensitiveKey(key) {
				return fmt.Errorf("content-bearing telemetry key %q is forbidden", key)
			}
			if err := rejectContentKeys(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := rejectContentKeys(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func sensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "gen_ai.prompt" ||
		normalized == "gen_ai.completion" ||
		strings.HasSuffix(normalized, ".prompt") ||
		strings.HasSuffix(normalized, ".content") ||
		strings.HasSuffix(normalized, ".payload") ||
		strings.HasSuffix(normalized, ".body")
}
