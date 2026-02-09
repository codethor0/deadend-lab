package invariants

import (
	"bytes"
	"crypto/rand"
	"testing"

	"deadend-lab/pkg/dee"
)

func TestNonceUniquenessSAFE(t *testing.T) {
	initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
	respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	plaintext := []byte("same")
	ad := []byte("ad")
	seen := make(map[string]bool)
	for i := 0; i < 256; i++ {
		ct, err := initSession.Encrypt(plaintext, ad)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		ctHex := string(ct)
		if seen[ctHex] {
			t.Fatalf("duplicate ciphertext at iteration %d: nonce reuse", i)
		}
		seen[ctHex] = true
		header := respSession.WireHeader(uint64(i))
		_, _ = respSession.Decrypt(ct, append(header, ad...))
	}

	t.Run("varies_AD", func(t *testing.T) {
		// Different AD must yield different nonce (AD hash is in derivation).
		msg := []byte("constant")
		ct1, _ := initSession.Encrypt(msg, []byte("ad1"))
		ct2, _ := initSession.Encrypt(msg, []byte("ad2"))
		if bytes.Equal(ct1, ct2) {
			t.Error("different AD must produce different ciphertext")
		}
	})

	t.Run("varies_msg", func(t *testing.T) {
		// Different plaintext must yield different ciphertext.
		ad := []byte("constant")
		ct1, _ := initSession.Encrypt([]byte("msg1"), ad)
		ct2, _ := initSession.Encrypt([]byte("msg2"), ad)
		if bytes.Equal(ct1, ct2) {
			t.Error("different plaintext must produce different ciphertext")
		}
	})
}

func TestCounterMonotonicityRejectReplay(t *testing.T) {
	initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
	respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct, _ := initSession.Encrypt([]byte("msg"), nil)
	header := respSession.WireHeader(0)
	_, err := respSession.Decrypt(ct, header)
	if err != nil {
		t.Fatalf("first decrypt: %v", err)
	}
	_, err = respSession.Decrypt(ct, header)
	if err == nil {
		t.Error("replay must be rejected")
	}
	if err != nil && err != dee.ErrDecrypt {
		t.Errorf("must return uniform ErrDecrypt, got %v", err)
	}
}

func TestCounterMonotonicityRejectOutOfOrder(t *testing.T) {
	initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
	respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct0, _ := initSession.Encrypt([]byte("zero"), nil)
	ct1, _ := initSession.Encrypt([]byte("one"), nil)

	header0 := respSession.WireHeader(0)
	header1 := respSession.WireHeader(1)

	_, err := respSession.Decrypt(ct1, header1)
	if err == nil {
		t.Error("counter 1 before 0 must be rejected")
	}
	if err != nil && err != dee.ErrDecrypt {
		t.Errorf("out-of-order must return uniform ErrDecrypt, got %v", err)
	}
	_, err = respSession.Decrypt(ct0, header0)
	if err != nil {
		t.Fatalf("counter 0: %v", err)
	}
	_, err = respSession.Decrypt(ct1, header1)
	if err != nil {
		t.Fatalf("counter 1: %v", err)
	}
}

func TestTranscriptBinding(t *testing.T) {
	initMsg1, initSession1, _ := dee.HandshakeInit(dee.Safe, rand.Reader)
	respMsg1, respSession1, _ := dee.HandshakeResp(dee.Safe, initMsg1, rand.Reader)
	_ = initSession1.HandshakeComplete(respMsg1)

	initMsg2, initSession2, _ := dee.HandshakeInit(dee.Safe, rand.Reader)
	respMsg2, respSession2, _ := dee.HandshakeResp(dee.Safe, initMsg2, rand.Reader)
	_ = initSession2.HandshakeComplete(respMsg2)

	if bytes.Equal(initSession1.SessionID(), initSession2.SessionID()) {
		t.Error("different transcripts must yield different session IDs")
	}

	plaintext := []byte("bind")
	ct1, _ := initSession1.Encrypt(plaintext, nil)
	ct2, _ := initSession2.Encrypt(plaintext, nil)
	if bytes.Equal(ct1, ct2) {
		t.Error("different sessions must produce different ciphertexts for same plaintext")
	}

	t.Run("mode_change", func(t *testing.T) {
		// Transcript includes version+mode; changing mode must change session.
		initSafe, initSessSafe, _ := dee.HandshakeInit(dee.Safe, nil)
		respSafe, respSessSafe, _ := dee.HandshakeResp(dee.Safe, initSafe, nil)
		_ = initSessSafe.HandshakeComplete(respSafe)

		initNaive, initSessNaive, _ := dee.HandshakeInit(dee.Naive, nil)
		respNaive, respSessNaive, _ := dee.HandshakeResp(dee.Naive, initNaive, nil)
		_ = initSessNaive.HandshakeComplete(respNaive)

		if bytes.Equal(initSessSafe.SessionID(), initSessNaive.SessionID()) {
			t.Error("different mode must yield different session ID")
		}
		_ = respSafe
		_ = respNaive
		_ = respSessSafe
		_ = respSessNaive
	})
	_ = respMsg1
	_ = respMsg2
	_ = respSession1
	_ = respSession2
}

func TestUniformFailure(t *testing.T) {
	initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
	respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
	_ = initSession.HandshakeComplete(respMsg)

	ct, _ := initSession.Encrypt([]byte("x"), nil)
	header := respSession.WireHeader(0)

	t.Run("tamper", func(t *testing.T) {
		bad := append([]byte(nil), ct...)
		if len(bad) > 20 {
			bad[20] ^= 0x01
		}
		_, err := respSession.Decrypt(bad, header)
		if err != dee.ErrDecrypt {
			t.Errorf("tamper: want ErrDecrypt, got %v", err)
		}
	})

	t.Run("wrong_AD", func(t *testing.T) {
		wrongHeader := respSession.WireHeader(0)
		wrongHeader[2] ^= 0x01
		_, err := respSession.Decrypt(ct, wrongHeader)
		if err != dee.ErrDecrypt {
			t.Errorf("wrong AD: want ErrDecrypt, got %v", err)
		}
	})

	t.Run("wrong_session", func(t *testing.T) {
		otherResp, otherRespSession, _ := dee.HandshakeResp(dee.Safe, initMsg, rand.Reader)
		otherHeader := otherRespSession.WireHeader(0)
		_, err := otherRespSession.Decrypt(ct, otherHeader)
		if err != dee.ErrDecrypt {
			t.Errorf("wrong session: want ErrDecrypt, got %v", err)
		}
		_ = otherResp
	})

	t.Run("wrong_counter", func(t *testing.T) {
		// Counter 1 when expecting 0 must return ErrDecrypt.
		header1 := respSession.WireHeader(1)
		_, err := respSession.Decrypt(ct, header1)
		if err != dee.ErrDecrypt {
			t.Errorf("wrong counter: want ErrDecrypt, got %v", err)
		}
	})
}
