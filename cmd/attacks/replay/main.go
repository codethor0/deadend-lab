package main

import (
	"fmt"
	"math/rand"
	"os"

	"deadend-lab/pkg/dee"
)

func main() {
	fmt.Println("=== NAIVE replay attack demo ===")

	rng := rand.New(rand.NewSource(54321))
	initMsg, initSession, err := dee.HandshakeInit(dee.Naive, rng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeInit: %v\n", err)
		os.Exit(1)
	}
	respMsg, respSession, err := dee.HandshakeResp(dee.Naive, initMsg, rng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeResp: %v\n", err)
		os.Exit(1)
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeComplete: %v\n", err)
		os.Exit(1)
	}

	plaintext := []byte("replay me")
	ct, _ := initSession.Encrypt(plaintext, nil)
	header := respSession.WireHeader(0)

	pt1, err := respSession.Decrypt(ct, header)
	if err != nil {
		fmt.Fprintf(os.Stderr, "first decrypt: %v\n", err)
		os.Exit(1)
	}
	pt2, err := respSession.Decrypt(ct, header)
	if err != nil {
		fmt.Printf("NAIVE replay: second decrypt failed (unexpected for NAIVE)\n")
		os.Exit(1)
	}

	fmt.Println("Steps: send same ciphertext twice, both decrypt succeed (no replay protection)")
	fmt.Printf("Replay accepted: %v\n", string(pt1) == string(pt2) && string(pt1) == string(plaintext))
}
