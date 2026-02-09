package dee

import (
	"bytes"
	"testing"
)

func TestNAIVEAcceptsReplay(t *testing.T) {
	initMsg, initSession, _ := HandshakeInit(Naive, nil)
	respMsg, respSession, _ := HandshakeResp(Naive, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct, _ := initSession.Encrypt([]byte("once"), nil)
	header := respSession.WireHeader(0)

	pt1, err := respSession.Decrypt(ct, header)
	if err != nil || string(pt1) != "once" {
		t.Fatalf("first decrypt: %v", err)
	}
	pt2, err := respSession.Decrypt(ct, header)
	if err != nil {
		t.Error("NAIVE must accept replay")
	}
	if !bytes.Equal(pt1, pt2) {
		t.Error("replay must return same plaintext")
	}
}

func TestSAFERejectsCallerNonce(t *testing.T) {
	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	nonce := make([]byte, NonceSize)
	_, err := initSession.EncryptNaiveWithNonce([]byte("x"), nil, nonce)
	if err != ErrDecrypt {
		t.Errorf("SAFE must reject EncryptNaiveWithNonce, got %v", err)
	}
	_ = respSession
}
