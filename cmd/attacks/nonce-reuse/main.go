package main

import (
	"fmt"
	"math/rand"
	"os"

	"deadend-lab/pkg/dee"
)

func main() {
	fmt.Println("=== NAIVE nonce-reuse attack demo ===")

	rng := rand.New(rand.NewSource(12345))
	initMsg, initSession, err := dee.HandshakeInit(dee.Naive, rng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeInit: %v\n", err)
		os.Exit(1)
	}
	respMsg, _, err := dee.HandshakeResp(dee.Naive, initMsg, rng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeResp: %v\n", err)
		os.Exit(1)
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeComplete: %v\n", err)
		os.Exit(1)
	}

	nonce := make([]byte, 12)
	for i := range nonce {
		nonce[i] = 0x41
	}

	p1 := []byte("AAAAAAAAAAAAAAAA")
	p2 := []byte("BBBBBBBBBBBBBBBB")
	ct1, err := initSession.EncryptNaiveWithNonce(p1, nil, nonce)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EncryptNaiveWithNonce 1: %v\n", err)
		os.Exit(1)
	}
	ct2, err := initSession.EncryptNaiveWithNonce(p2, nil, nonce)
	if err != nil {
		fmt.Fprintf(os.Stderr, "EncryptNaiveWithNonce 2: %v\n", err)
		os.Exit(1)
	}

	// ChaCha20: ct = keystream XOR plaintext for ciphertext portion
	// Same nonce = same keystream, so ct1 XOR ct2 = p1 XOR p2
	// Recover p2: p2 = ct1 XOR ct2 XOR p1
	recovered := make([]byte, len(p1))
	for i := range recovered {
		recovered[i] = ct1[i] ^ ct2[i] ^ p1[i]
	}

	fmt.Println("Steps: same nonce -> same keystream -> ct1 XOR ct2 = p1 XOR p2")
	fmt.Println("With known p1, recover p2 = ct1 XOR ct2 XOR p1")
	fmt.Printf("Recovered plaintext == expected: %v\n", string(recovered) == string(p2))
}
