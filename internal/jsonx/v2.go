//go:build jsonv2

package jsonx

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"

	jsonv2 "encoding/json/v2"
)

func marshalV1(v any) ([]byte, error) {
	return json.Marshal(v)
}

func marshalIndentV1(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func unmarshalV1(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func marshalV2(v any) ([]byte, error) {
	return jsonv2.Marshal(v)
}

func marshalIndentV2(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf, jsontext.WithIndent(prefix, indent))
	if err := jsonv2.MarshalEncode(enc, v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func unmarshalV2(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v)
}
