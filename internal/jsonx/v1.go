//go:build !jsonv2

package jsonx

import "encoding/json"

func marshalV1(v any) ([]byte, error) {
	return json.Marshal(v)
}

func marshalIndentV1(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func unmarshalV1(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func marshalV2(_ any) ([]byte, error) {
	panic("json v2 not available: build with GOEXPERIMENT=jsonv2")
}

func marshalIndentV2(_ any, _, _ string) ([]byte, error) {
	panic("json v2 not available: build with GOEXPERIMENT=jsonv2")
}

func unmarshalV2(_ []byte, _ any) error {
	panic("json v2 not available: build with GOEXPERIMENT=jsonv2")
}
