package common

// Domain separation labels for HKDF.
const (
	LabelMaster       = "dee-v1-master"
	LabelAEADKey      = "dee-v1-aead-key"
	LabelNonceBase    = "dee-v1-nonce-base"
	LabelAuditTag     = "dee-v1-audit-tag-key"
	LabelRekey        = "dee-v1-rekey"
	LabelRekeyRatchet = "dee-v1-rekey-ratchet"
)
