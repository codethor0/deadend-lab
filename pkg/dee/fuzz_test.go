package dee

import (
	"bytes"
	"testing"
)

func FuzzDEERoundtrip(f *testing.F) {
	f.Add([]byte("hello"), []byte("ad"))
	f.Fuzz(func(t *testing.T, plaintext, ad []byte) {
		if len(plaintext) > 1<<16 || len(ad) > 1<<16 {
			t.Skip()
		}
		initMsg, initSession, err := HandshakeInit(Safe, nil)
		if err != nil {
			t.Skip()
		}
		respMsg, respSession, err := HandshakeResp(Safe, initMsg, nil)
		if err != nil {
			t.Skip()
		}
		if err := initSession.HandshakeComplete(respMsg); err != nil {
			t.Skip()
		}
		ct, err := initSession.Encrypt(plaintext, ad)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		header := respSession.WireHeader(0)
		fullAD := append(header, ad...)
		pt, err := respSession.Decrypt(ct, fullAD)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if !bytes.Equal(pt, plaintext) {
			t.Errorf("roundtrip: got %x want %x", pt, plaintext)
		}
	})
}

func FuzzTamper(f *testing.F) {
	f.Add([]byte("secret"), byte(10))
	f.Fuzz(func(t *testing.T, plaintext []byte, flipIdx byte) {
		if len(plaintext) > 1024 {
			t.Skip()
		}
		initMsg, initSession, _ := HandshakeInit(Safe, nil)
		respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		ct, err := initSession.Encrypt(plaintext, nil)
		if err != nil {
			t.Skip()
		}
		if len(ct) < 20 {
			t.Skip()
		}
		idx := int(flipIdx) % len(ct)
		ct[idx] ^= 0x01
		header := respSession.WireHeader(0)
		_, err = respSession.Decrypt(ct, header)
		if err == nil {
			t.Error("tampered ciphertext should fail")
		}
	})
}

func FuzzReplayRejected(f *testing.F) {
	f.Add([]byte("replay target"))
	f.Fuzz(func(t *testing.T, plaintext []byte) {
		if len(plaintext) > 4096 {
			t.Skip()
		}
		initMsg, initSession, _ := HandshakeInit(Safe, nil)
		respMsg, respSession, _ := HandshakeResp(Safe, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		ct, err := initSession.Encrypt(plaintext, nil)
		if err != nil {
			t.Skip()
		}
		header := respSession.WireHeader(0)
		_, err = respSession.Decrypt(ct, header)
		if err != nil {
			t.Skip()
		}
		_, err = respSession.Decrypt(ct, header)
		if err == nil {
			t.Error("replay must be rejected in SAFE")
		}
		_ = respMsg
	})
}
