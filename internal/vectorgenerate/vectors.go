package vectorgenerate

import (
	"encoding/hex"

	"deadend-lab/internal/drbg"
	"deadend-lab/pkg/dee"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

// MessageVectorEntry is one message in the vector.
type MessageVectorEntry struct {
	Counter   uint64
	MsgHex    string
	ADHex     string
	CipherHex string
}

// MessageVector is the full deterministic message vector.
type MessageVector struct {
	SessionIDTruncHex string
	TranscriptHex     string
	Messages          []MessageVectorEntry
	Label             string
}

// Seed used for vector generation. Fixed for reproducibility.
const VectorSeed = 42

// GenerateMessageVector produces a deterministic 2-message vector. All randomness
// is driven from the DRBG via HandshakeInitDeterministic/HandshakeRespDeterministic;
// output is identical across runs and machines.
func GenerateMessageVector(seed int64) (MessageVector, error) {
	rng := drbg.NewSeed(seed)

	initMsg, initSession, err := dee.HandshakeInitDeterministic(dee.Safe, rng)
	if err != nil {
		return MessageVector{}, err
	}

	encSeed := make([]byte, kyber768.EncapsulationSeedSize)
	for i := range encSeed {
		encSeed[i] = byte(seed)
	}

	respMsg, respSession, err := dee.HandshakeRespDeterministic(dee.Safe, initMsg, rng, encSeed)
	if err != nil {
		return MessageVector{}, err
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		return MessageVector{}, err
	}

	msg0 := []byte("vector message 0")
	ad0 := []byte("associated data 0")
	msg1 := []byte("vector message 1")
	ad1 := []byte("associated data 1")

	ct0, err := initSession.Encrypt(msg0, ad0)
	if err != nil {
		return MessageVector{}, err
	}
	ct1, err := initSession.Encrypt(msg1, ad1)
	if err != nil {
		return MessageVector{}, err
	}

	sessionID := initSession.SessionID()
	truncLen := 8
	if len(sessionID) < truncLen {
		truncLen = len(sessionID)
	}
	sessionIDTrunc := sessionID[:truncLen]
	transcript := sessionID

	_ = respMsg
	_ = respSession

	return MessageVector{
		SessionIDTruncHex: hex.EncodeToString(sessionIDTrunc),
		TranscriptHex:     hex.EncodeToString(transcript),
		Messages: []MessageVectorEntry{
			{Counter: 0, MsgHex: hex.EncodeToString(msg0), ADHex: hex.EncodeToString(ad0), CipherHex: hex.EncodeToString(ct0)},
			{Counter: 1, MsgHex: hex.EncodeToString(msg1), ADHex: hex.EncodeToString(ad1), CipherHex: hex.EncodeToString(ct1)},
		},
		Label: "two_message_safe_deterministic",
	}, nil
}
