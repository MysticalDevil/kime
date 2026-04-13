// Package jsonx provides an adapter between encoding/json and encoding/json/v2.
// Set KIME_JSON_V2=1 at runtime and build with GOEXPERIMENT=jsonv2 to use v2.
package jsonx

import "os"

var useV2 = os.Getenv("KIME_JSON_V2") == "1"

// Marshal delegates to the configured JSON implementation.
func Marshal(v any) ([]byte, error) {
	if useV2 {
		return marshalV2(v)
	}

	return marshalV1(v)
}

// MarshalIndent delegates to the configured JSON implementation.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	if useV2 {
		return marshalIndentV2(v, prefix, indent)
	}

	return marshalIndentV1(v, prefix, indent)
}

// Unmarshal delegates to the configured JSON implementation.
func Unmarshal(data []byte, v any) error {
	if useV2 {
		return unmarshalV2(data, v)
	}

	return unmarshalV1(data, v)
}
