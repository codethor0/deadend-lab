package vectors

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"deadend-lab/internal/vectorgenerate"
	"deadend-lab/pkg/common"
)

type Vector struct {
	Description       string `json:"description"`
	PRKHex            string `json:"prk_hex"`
	ExpectedMasterHex string `json:"expected_master_hex"`
}

type MessageVectorEntry struct {
	Counter   uint64 `json:"counter"`
	MsgHex    string `json:"msg_hex"`
	ADHex     string `json:"ad_hex"`
	CipherHex string `json:"cipher_hex"`
}

type MessageVector struct {
	SessionIDTruncHex string               `json:"session_id_trunc_hex"`
	TranscriptHex     string               `json:"transcript_hex"`
	Messages          []MessageVectorEntry `json:"messages"`
	Label             string               `json:"label"`
}

func TestHKDFVector(t *testing.T) {
	path := filepath.Join("testdata", "handshake_vector.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("vector file not found: %v", err)
	}
	var v Vector
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	prk, err := hex.DecodeString(v.PRKHex)
	if err != nil {
		t.Fatalf("decode prk: %v", err)
	}
	expected, err := hex.DecodeString(v.ExpectedMasterHex)
	if err != nil {
		t.Fatalf("decode expected: %v", err)
	}

	got := common.Expand(prk, common.LabelMaster, 32)
	if len(got) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(got))
	}
	if !bytes.Equal(got, expected) {
		t.Errorf("key mismatch: got %x want %x", got, expected)
	}
}

func TestMessageVectorByteForByte(t *testing.T) {
	expected, err := vectorgenerate.GenerateMessageVector(vectorgenerate.VectorSeed)
	if err != nil {
		t.Fatalf("GenerateMessageVector: %v", err)
	}

	path := filepath.Join("testdata", "message_vector.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("vector file not found: %v (run: make vectors)", err)
	}
	var v MessageVector
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(v.Messages) < 2 {
		t.Fatal("vector must have at least 2 messages")
	}

	// Byte-for-byte validation. Test fails on mismatch.
	if v.SessionIDTruncHex != expected.SessionIDTruncHex {
		t.Errorf("session_id_trunc_hex: got %s want %s", v.SessionIDTruncHex, expected.SessionIDTruncHex)
	}
	if v.TranscriptHex != expected.TranscriptHex {
		t.Errorf("transcript_hex: mismatch (run: make vectors)")
	}
	for i := range v.Messages {
		if i >= len(expected.Messages) {
			t.Fatalf("file has more messages than expected")
		}
		if v.Messages[i].MsgHex != expected.Messages[i].MsgHex {
			t.Errorf("message %d msg_hex mismatch", i)
		}
		if v.Messages[i].ADHex != expected.Messages[i].ADHex {
			t.Errorf("message %d ad_hex mismatch", i)
		}
		if v.Messages[i].CipherHex != expected.Messages[i].CipherHex {
			t.Errorf("message %d cipher_hex: got %s want %s", i, v.Messages[i].CipherHex, expected.Messages[i].CipherHex)
		}
	}
}
