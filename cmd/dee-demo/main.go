package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"deadend-lab/pkg/dee"
)

func main() {
	modeFlag := flag.String("mode", "SAFE", "DEE mode: SAFE or NAIVE")
	msgFlag := flag.String("msg", "hello", "Message to send")
	flag.Parse()

	var m dee.Mode
	switch *modeFlag {
	case "SAFE":
		m = dee.Safe
	case "NAIVE":
		m = dee.Naive
	default:
		fmt.Fprintf(os.Stderr, "invalid mode: %s\n", *modeFlag)
		os.Exit(1)
	}

	initMsg, initSession, err := dee.HandshakeInit(m, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeInit: %v\n", err)
		os.Exit(1)
	}

	respMsg, respSession, err := dee.HandshakeResp(m, initMsg, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeResp: %v\n", err)
		os.Exit(1)
	}

	if err := initSession.HandshakeComplete(respMsg); err != nil {
		fmt.Fprintf(os.Stderr, "HandshakeComplete: %v\n", err)
		os.Exit(1)
	}

	plaintext := []byte(*msgFlag)
	ct, err := initSession.Encrypt(plaintext, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Encrypt: %v\n", err)
		os.Exit(1)
	}

	header := respSession.WireHeader(0)
	pt, err := respSession.Decrypt(ct, header)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Decrypt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Mode: %s\n", m)
	fmt.Printf("SessionID: %s\n", hex.EncodeToString(initSession.SessionID()))
	fmt.Printf("Plaintext: %s\n", plaintext)
	fmt.Printf("Decrypted: %s\n", pt)
	fmt.Printf("Roundtrip OK: %v\n", string(pt) == string(plaintext))
}
