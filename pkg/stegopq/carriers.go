package stegopq

// Carrier identifies a stego carrier type.
type Carrier byte

const (
	CarrierA Carrier = iota + 1 // JSON telemetry
	CarrierB                    // HTTP header
	CarrierC                    // URL query
)
