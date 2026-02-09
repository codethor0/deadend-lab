package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"deadend-lab/internal/vectorgenerate"
)

type jsonMessageEntry struct {
	Counter   uint64 `json:"counter"`
	MsgHex    string `json:"msg_hex"`
	ADHex     string `json:"ad_hex"`
	CipherHex string `json:"cipher_hex"`
}

type jsonMessageVector struct {
	SessionIDTruncHex string             `json:"session_id_trunc_hex"`
	TranscriptHex     string             `json:"transcript_hex"`
	Messages          []jsonMessageEntry `json:"messages"`
	Label             string             `json:"label"`
}

func main() {
	outDir := flag.String("out", "tests/vectors/testdata", "Output directory")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "MkdirAll: %v\n", err)
		os.Exit(1)
	}

	v, err := vectorgenerate.GenerateMessageVector(vectorgenerate.VectorSeed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GenerateMessageVector: %v\n", err)
		os.Exit(1)
	}

	entries := make([]jsonMessageEntry, len(v.Messages))
	for i, m := range v.Messages {
		entries[i] = jsonMessageEntry{Counter: m.Counter, MsgHex: m.MsgHex, ADHex: m.ADHex, CipherHex: m.CipherHex}
	}
	jv := jsonMessageVector{
		SessionIDTruncHex: v.SessionIDTruncHex,
		TranscriptHex:     v.TranscriptHex,
		Messages:          entries,
		Label:             v.Label,
	}

	b, _ := json.MarshalIndent(jv, "", "  ")
	path := filepath.Join(*outDir, "message_vector.json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "WriteFile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", path)
}
