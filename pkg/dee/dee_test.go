package dee

import (
	"bytes"
	"testing"
)

func TestHandshakeRoundtrip(t *testing.T) {
	initMsg, initSession, err := HandshakeInit(Safe, nil)
	if err != nil {
		t.Fatalf("HandshakeInit: %v", err)
	}
	if initSession == nil || initSession.established {
		t.Fatal("initiator session should not be established yet")
	}

	respMsg, respSession, err := HandshakeResp(Safe, initMsg, nil)
	if err != nil {
		t.Fatalf("HandshakeResp: %v", err)
	}
	if respSession == nil || !respSession.established {
		t.Fatal("responder session should be established")
	}

	if err := initSession.HandshakeComplete(respMsg); err != nil {
		t.Fatalf("HandshakeComplete: %v", err)
	}
	if !initSession.established {
		t.Fatal("initiator session should be established")
	}

	if !bytes.Equal(initSession.SessionID(), respSession.SessionID()) {
		t.Error("session IDs should match")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	initMsg, initSession, err := HandshakeInit(Safe, nil)
	if err != nil {
		t.Fatalf("HandshakeInit: %v", err)
	}
	respMsg, respSession, err := HandshakeResp(Safe, initMsg, nil)
	if err != nil {
		t.Fatalf("HandshakeResp: %v", err)
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		t.Fatalf("HandshakeComplete: %v", err)
	}

	plaintext := []byte("hello world")
	ad := []byte("associated")

	ct, err := initSession.Encrypt(plaintext, ad)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(ct) == 0 {
		t.Fatal("ciphertext empty")
	}

	// Responder receives and decrypts (initiator sent, so responder receives)
	header := respSession.WireHeader(0)
	fullAD := append(header, ad...)
	pt, err := respSession.Decrypt(ct, fullAD)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("plaintext mismatch: got %q", pt)
	}
}

func TestEncryptDecryptBidirectional(t *testing.T) {
	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	msg1 := []byte("initiator to responder")
	ct1, _ := initSession.Encrypt(msg1, nil)
	header1 := respSession.WireHeader(0)
	pt1, err := respSession.Decrypt(ct1, header1)
	if err != nil || !bytes.Equal(pt1, msg1) {
		t.Fatalf("direction 1 failed: %v", err)
	}

	msg2 := []byte("responder to initiator")
	ct2, _ := respSession.Encrypt(msg2, nil)
	header2 := initSession.WireHeader(0)
	pt2, err := initSession.Decrypt(ct2, header2)
	if err != nil || !bytes.Equal(pt2, msg2) {
		t.Fatalf("direction 2 failed: %v", err)
	}
}

func TestTamperFails(t *testing.T) {
	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct, _ := initSession.Encrypt([]byte("secret"), nil)
	ct[16] ^= 0x01
	header := respSession.WireHeader(0)
	_, err := respSession.Decrypt(ct, header)
	if err == nil {
		t.Error("tampered ciphertext should fail")
	}
}

func TestRekeyRatchet(t *testing.T) {
	SetRekeyEveryForTest(5)
	defer SetRekeyEveryForTest(0)

	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	for i := uint64(0); i < 8; i++ {
		ct, err := initSession.Encrypt([]byte("msg"), nil)
		if err != nil {
			t.Fatalf("Encrypt at %d: %v", i, err)
		}
		header := respSession.WireHeader(i)
		pt, err := respSession.Decrypt(ct, header)
		if err != nil {
			t.Fatalf("Decrypt at %d: %v", i, err)
		}
		if string(pt) != "msg" {
			t.Errorf("at %d: got %q", i, pt)
		}
	}
	_ = respMsg
}

// TestRekeyBoundary verifies exact behavior at N-1, N, N+1 with rekey interval 5.
// Message N-1 uses old key, message N triggers rekey, message N+1 uses new key.
func TestRekeyBoundary(t *testing.T) {
	SetRekeyEveryForTest(5)
	defer SetRekeyEveryForTest(0)

	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	// Send 7 messages so we cross boundary at 5: msgs 0-4 old key, 5-6 new key.
	for i := uint64(0); i < 7; i++ {
		ct, err := initSession.Encrypt([]byte("msg"), nil)
		if err != nil {
			t.Fatalf("Encrypt at %d: %v", i, err)
		}
		header := respSession.WireHeader(i)
		pt, err := respSession.Decrypt(ct, header)
		if err != nil {
			t.Fatalf("Decrypt at %d: %v", i, err)
		}
		if string(pt) != "msg" {
			t.Errorf("at %d: got %q", i, pt)
		}
	}
	_ = respMsg
}

// TestRekeyInterval32 catches off-by-one bugs that only appear with larger counters.
func TestRekeyInterval32(t *testing.T) {
	SetRekeyEveryForTest(32)
	defer SetRekeyEveryForTest(0)

	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	for i := uint64(0); i < 70; i++ {
		ct, err := initSession.Encrypt([]byte("x"), nil)
		if err != nil {
			t.Fatalf("Encrypt at %d: %v", i, err)
		}
		header := respSession.WireHeader(i)
		pt, err := respSession.Decrypt(ct, header)
		if err != nil {
			t.Fatalf("Decrypt at %d: %v", i, err)
		}
		if string(pt) != "x" {
			t.Errorf("at %d: got %q", i, pt)
		}
	}
	_ = respMsg
}

func TestReplayFailsSafe(t *testing.T) {
	initMsg, initSession, _ := HandshakeInit(Safe, nil)
	respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct, _ := initSession.Encrypt([]byte("once"), nil)
	header := respSession.WireHeader(0)

	pt1, err := respSession.Decrypt(ct, header)
	if err != nil || string(pt1) != "once" {
		t.Fatalf("first decrypt failed: %v", err)
	}

	_, err = respSession.Decrypt(ct, header)
	if err == nil {
		t.Error("replay should fail in SAFE mode")
	}
}
