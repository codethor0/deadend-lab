package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"deadend-lab/pkg/dee"
)

var port string

func main() {
	flag.StringVar(&port, "port", "8080", "Listen port")
	flag.Parse()

	http.HandleFunc("/scenario/safe", handleScenarioSafe)
	http.HandleFunc("/scenario/naive", handleScenarioNaive)
	http.HandleFunc("/health", handleHealth)

	log.Printf("lab-server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type ScenarioResult struct {
	OK             bool   `json:"ok"`
	Mode           string `json:"mode"`
	Version        uint8  `json:"version"`
	Carrier        string `json:"carrier"`
	ReasonCode     string `json:"reason_code"`
	HandshakeMs    int64  `json:"handshake_ms,omitempty"`
	EncryptMs      int64  `json:"encrypt_ms,omitempty"`
	DecryptMs      int64  `json:"decrypt_ms,omitempty"`
	CiphertextLen  int    `json:"ciphertext_len,omitempty"`
	ReplayRejected *bool  `json:"replay_rejected,omitempty"`
	SessionIDTrunc string `json:"session_id_trunc,omitempty"`
}

func handleScenarioSafe(w http.ResponseWriter, r *http.Request) {
	handleScenario(w, r, dee.Safe)
}

func handleScenarioNaive(w http.ResponseWriter, r *http.Request) {
	handleScenario(w, r, dee.Naive)
}

func handleScenario(w http.ResponseWriter, r *http.Request, mode dee.Mode) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	res := ScenarioResult{
		Mode:    mode.String(),
		Version: dee.Version,
		Carrier: "none",
	}

	t0 := time.Now()
	initMsg, initSession, err := dee.HandshakeInit(mode, nil)
	if err != nil {
		res.ReasonCode = "error"
		writeResult(w, res)
		return
	}
	respMsg, respSession, err := dee.HandshakeResp(mode, initMsg, nil)
	if err != nil {
		res.ReasonCode = "error"
		writeResult(w, res)
		return
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		res.ReasonCode = "error"
		writeResult(w, res)
		return
	}
	res.HandshakeMs = time.Since(t0).Milliseconds()

	plaintext := []byte("test message")
	t1 := time.Now()
	ct, err := initSession.Encrypt(plaintext, nil)
	if err != nil {
		res.ReasonCode = "error"
		writeResult(w, res)
		return
	}
	res.EncryptMs = time.Since(t1).Milliseconds()
	res.CiphertextLen = len(ct)

	header := respSession.WireHeader(0)
	t2 := time.Now()
	pt, err := respSession.Decrypt(ct, header)
	if err != nil {
		res.ReasonCode = "error"
		writeResult(w, res)
		return
	}
	res.DecryptMs = time.Since(t2).Milliseconds()

	res.OK = string(pt) == string(plaintext)
	if res.OK {
		res.ReasonCode = "ok"
		sid := initSession.SessionID()
		if len(sid) >= 8 {
			res.SessionIDTrunc = hex.EncodeToString(sid[:8])
		}
		rejected := mode == dee.Safe
		res.ReplayRejected = &rejected
	}
	writeResult(w, res)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeResult(w http.ResponseWriter, r ScenarioResult) {
	_ = json.NewEncoder(w).Encode(r)
}
