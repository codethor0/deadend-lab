package stegopq

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	MaxPayloadSize = 4096
)

var (
	ErrPayloadTooLarge = errors.New("payload too large")
	ErrDecode          = errors.New("decode failed")
)

// Encode embeds payload into a carrier, producing a string suitable for the transport.
func Encode(c Carrier, payload []byte) (string, error) {
	if len(payload) > MaxPayloadSize {
		return "", ErrPayloadTooLarge
	}
	b64 := base64.RawURLEncoding.EncodeToString(payload)
	switch c {
	case CarrierA:
		return encodeA(b64)
	case CarrierB:
		return encodeB(b64)
	case CarrierC:
		return encodeC(b64)
	default:
		return "", ErrDecode
	}
}

// Decode extracts payload from an encoded carrier string.
func Decode(c Carrier, encoded string) ([]byte, error) {
	var b64 string
	var err error
	switch c {
	case CarrierA:
		b64, err = decodeA(encoded)
	case CarrierB:
		b64, err = decodeB(encoded)
	case CarrierC:
		b64, err = decodeC(encoded)
	default:
		return nil, ErrDecode
	}
	if err != nil {
		return nil, err
	}
	return base64.RawURLEncoding.DecodeString(b64)
}

func encodeA(b64 string) (string, error) {
	obj := map[string]string{
		"trace_id": b64,
		"span_id":  "0",
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeA(encoded string) (string, error) {
	var obj map[string]string
	if err := json.Unmarshal([]byte(encoded), &obj); err != nil {
		return "", ErrDecode
	}
	traceID, ok := obj["trace_id"]
	if !ok {
		return "", ErrDecode
	}
	return traceID, nil
}

func encodeB(b64 string) (string, error) {
	return "X-Trace-ID: " + b64, nil
}

func decodeB(encoded string) (string, error) {
	prefix := "X-Trace-ID:"
	if !strings.HasPrefix(encoded, prefix) {
		return "", ErrDecode
	}
	return strings.TrimSpace(strings.TrimPrefix(encoded, prefix)), nil
}

func encodeC(b64 string) (string, error) {
	ff := b64
	if len(ff) > 8 {
		ff = ff[:8]
	}
	v := url.Values{}
	v.Set("ff", ff)
	v.Set("v", b64)
	return "?" + v.Encode(), nil
}

func decodeC(encoded string) (string, error) {
	encoded = strings.TrimPrefix(encoded, "?")
	if encoded == "" {
		return "", ErrDecode
	}
	vals, err := url.ParseQuery(encoded)
	if err != nil {
		return "", ErrDecode
	}
	if _, ok := vals["v"]; !ok {
		return "", ErrDecode
	}
	return vals.Get("v"), nil
}

// EncodeCarrierA returns JSON with trace_id and span_id.
func EncodeCarrierA(payload []byte) (string, error) {
	return Encode(CarrierA, payload)
}

// EncodeCarrierB returns HTTP header line.
func EncodeCarrierB(payload []byte) (string, error) {
	return Encode(CarrierB, payload)
}

// EncodeCarrierC returns URL query string.
func EncodeCarrierC(payload []byte) (string, error) {
	return Encode(CarrierC, payload)
}

// DecodeCarrierA parses JSON and extracts trace_id.
func DecodeCarrierA(encoded string) ([]byte, error) {
	return Decode(CarrierA, encoded)
}

// DecodeCarrierB parses X-Trace-ID header value.
func DecodeCarrierB(encoded string) ([]byte, error) {
	return Decode(CarrierB, encoded)
}

// DecodeCarrierC parses query and extracts v param.
func DecodeCarrierC(encoded string) ([]byte, error) {
	return Decode(CarrierC, encoded)
}

// CarrierName returns a short name for the carrier.
func CarrierName(c Carrier) string {
	switch c {
	case CarrierA:
		return "json"
	case CarrierB:
		return "header"
	case CarrierC:
		return "query"
	default:
		return fmt.Sprintf("unknown(%d)", c)
	}
}
