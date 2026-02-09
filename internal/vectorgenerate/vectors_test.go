package vectorgenerate

import (
	"testing"
)

func TestGenerateDeterministic(t *testing.T) {
	v1, err := GenerateMessageVector(VectorSeed)
	if err != nil {
		t.Fatalf("GenerateMessageVector: %v", err)
	}
	v2, err := GenerateMessageVector(VectorSeed)
	if err != nil {
		t.Fatalf("GenerateMessageVector: %v", err)
	}
	if v1.SessionIDTruncHex != v2.SessionIDTruncHex {
		t.Errorf("session_id: %s != %s", v1.SessionIDTruncHex, v2.SessionIDTruncHex)
	}
	for i := range v1.Messages {
		if v1.Messages[i].CipherHex != v2.Messages[i].CipherHex {
			t.Errorf("message %d cipher mismatch", i)
		}
	}
}
