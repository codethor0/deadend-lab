package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"deadend-lab/pkg/dee"
)

type CorpusEntry struct {
	Label   string `json:"label"`
	Mode    string `json:"mode"`
	Payload string `json:"payload,omitempty"`
	Err     string `json:"err,omitempty"`
}

func main() {
	outDir := flag.String("out", "challenge/datasets", "Output directory")
	seed := flag.Int64("seed", 42, "Random seed (ignored, uses crypto/rand)")
	_ = seed
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "MkdirAll: %v\n", err)
		os.Exit(1)
	}

	corpus := []CorpusEntry{}

	// Nonce reuse (NAIVE allows caller-supplied nonce; same nonce twice)
	{
		initMsg, initSession, _ := dee.HandshakeInit(dee.Naive, nil)
		respMsg, _, _ := dee.HandshakeResp(dee.Naive, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		nonce := make([]byte, 12)
		if _, err := rand.Read(nonce); err == nil {
			ct1, _ := initSession.EncryptNaiveWithNonce([]byte("secret1"), nil, nonce)
			ct2, _ := initSession.EncryptNaiveWithNonce([]byte("secret2"), nil, nonce)
			corpus = append(corpus, CorpusEntry{Label: "nonce_reuse_ct1", Mode: "NAIVE", Payload: fmt.Sprintf("%x", ct1)})
			corpus = append(corpus, CorpusEntry{Label: "nonce_reuse_ct2", Mode: "NAIVE", Payload: fmt.Sprintf("%x", ct2)})
		}
		_ = respMsg
	}

	// Replay (SAFE rejects)
	{
		initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
		respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		ct, _ := initSession.Encrypt([]byte("once"), nil)
		header := respSession.WireHeader(0)
		pt, err := respSession.Decrypt(ct, header)
		_ = pt
		corpus = append(corpus, CorpusEntry{Label: "replay_first_ok", Mode: "SAFE", Payload: fmt.Sprintf("%x", ct)})
		_, err2 := respSession.Decrypt(ct, header)
		corpus = append(corpus, CorpusEntry{Label: "replay_second_reject", Mode: "SAFE", Err: fmt.Sprintf("%v", err2)})
		_ = err
	}

	// Tamper
	{
		initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
		respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		ct, _ := initSession.Encrypt([]byte("secret"), nil)
		if len(ct) > 20 {
			ct[20] ^= 0x01
		}
		header := respSession.WireHeader(0)
		_, err := respSession.Decrypt(ct, header)
		corpus = append(corpus, CorpusEntry{Label: "tamper_reject", Mode: "SAFE", Err: fmt.Sprintf("%v", err)})
	}

	// Normal roundtrip
	{
		initMsg, initSession, _ := dee.HandshakeInit(dee.Safe, nil)
		respMsg, respSession, _ := dee.HandshakeResp(dee.Safe, initMsg, nil)
		_ = initSession.HandshakeComplete(respMsg)
		plain := make([]byte, 32)
		if _, err := rand.Read(plain); err == nil {
			ct, _ := initSession.Encrypt(plain, nil)
			header := respSession.WireHeader(0)
			pt, _ := respSession.Decrypt(ct, header)
			corpus = append(corpus, CorpusEntry{Label: "roundtrip_ok", Mode: "SAFE", Payload: fmt.Sprintf("%x", pt)})
		}
		_ = respMsg
	}

	outPath := filepath.Join(*outDir, "corpus.json")
	b, _ := json.MarshalIndent(corpus, "", "  ")
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "WriteFile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", outPath)
}
