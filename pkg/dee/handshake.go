package dee

import (
	"crypto/ecdh"
	"crypto/rand"
	"errors"
	"io"

	"deadend-lab/pkg/common"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

const (
	x25519PubSize = 32
	kyberPubSize  = kyber768.PublicKeySize
	kyberCtSize   = kyber768.CiphertextSize
	kyberSSSize   = kyber768.SharedKeySize
)

var (
	ErrHandshake = errors.New("handshake failed")
)

// HandshakeInit starts a handshake as initiator. Sends X25519 pub + Kyber pub.
// Responder will encapsulate to Kyber pub and send ciphertext.
func HandshakeInit(mode Mode, randReader io.Reader) (initMsg []byte, session *Session, err error) {
	if randReader == nil {
		randReader = rand.Reader
	}

	curve := ecdh.X25519()
	xPriv, err := curve.GenerateKey(randReader)
	if err != nil {
		return nil, nil, err
	}
	xPub := xPriv.PublicKey().Bytes()

	kyberPk, kyberSk, err := kyber768.GenerateKeyPair(randReader)
	if err != nil {
		return nil, nil, err
	}
	kyberPubBytes := make([]byte, kyberPubSize)
	kyberPk.Pack(kyberPubBytes)

	initMsg = buildHandshakeInitMsg(Version, byte(mode), xPub, kyberPubBytes)

	session = &Session{
		mode:      mode,
		xPriv:     xPriv,
		initMsg:   initMsg,
		kyberPriv: kyberSk,
	}
	return initMsg, session, nil
}

// HandshakeResp completes handshake as responder. Encapsulates to init's Kyber pub.
func HandshakeResp(mode Mode, initMsg []byte, randReader io.Reader) (respMsg []byte, session *Session, err error) {
	if randReader == nil {
		randReader = rand.Reader
	}

	ver, m, xPubInit, kyberPubInit, err := parseHandshakeInitMsg(initMsg)
	if err != nil || ver != Version || m != byte(mode) {
		return nil, nil, ErrHandshake
	}

	curve := ecdh.X25519()
	xPriv, err := curve.GenerateKey(randReader)
	if err != nil {
		return nil, nil, err
	}
	xPub := xPriv.PublicKey().Bytes()

	peerXPub, err := curve.NewPublicKey(xPubInit)
	if err != nil {
		return nil, nil, ErrHandshake
	}
	xShared, err := xPriv.ECDH(peerXPub)
	if err != nil {
		return nil, nil, ErrHandshake
	}

	var kyberPk kyber768.PublicKey
	if len(kyberPubInit) != kyberPubSize {
		return nil, nil, ErrHandshake
	}
	kyberPk.Unpack(kyberPubInit)
	kyberCt := make([]byte, kyberCtSize)
	kyberSS := make([]byte, kyberSSSize)
	kyberPk.EncapsulateTo(kyberCt, kyberSS, nil)

	respMsg = buildHandshakeRespMsg(Version, byte(mode), xPub, kyberCt)
	return handshakeRespFinish(mode, initMsg, respMsg, xShared, kyberSS)
}

// HandshakeInitDeterministic is like HandshakeInit but uses drbg io.Reader for fully
// deterministic key generation. X25519 uses NewPrivateKey(drbg bytes); Kyber uses
// NewKeyFromSeed(drbg bytes). Used only for vector generation.
func HandshakeInitDeterministic(mode Mode, drbg io.Reader) ([]byte, *Session, error) {
	curve := ecdh.X25519()
	xBytes := make([]byte, 32)
	if _, err := io.ReadFull(drbg, xBytes); err != nil {
		return nil, nil, err
	}
	xPriv, err := curve.NewPrivateKey(xBytes)
	if err != nil {
		return nil, nil, err
	}
	xPub := xPriv.PublicKey().Bytes()

	kyberSeed := make([]byte, kyber768.KeySeedSize)
	if _, err := io.ReadFull(drbg, kyberSeed); err != nil {
		return nil, nil, err
	}
	kyberPk, kyberSk := kyber768.NewKeyFromSeed(kyberSeed)
	kyberPubBytes := make([]byte, kyberPubSize)
	kyberPk.Pack(kyberPubBytes)

	initMsg := buildHandshakeInitMsg(Version, byte(mode), xPub, kyberPubBytes)
	session := &Session{
		mode:      mode,
		xPriv:     xPriv,
		initMsg:   initMsg,
		kyberPriv: kyberSk,
	}
	return initMsg, session, nil
}

// HandshakeRespDeterministic is like HandshakeResp but uses deterministic key gen
// (NewPrivateKey, NewKeyFromSeed) and encSeed for Kyber encapsulation.
func HandshakeRespDeterministic(mode Mode, initMsg []byte, drbg io.Reader, encSeed []byte) ([]byte, *Session, error) {
	if drbg == nil {
		drbg = rand.Reader
	}
	ver, m, xPubInit, kyberPubInit, err := parseHandshakeInitMsg(initMsg)
	if err != nil || ver != Version || m != byte(mode) {
		return nil, nil, ErrHandshake
	}
	curve := ecdh.X25519()
	xBytes := make([]byte, 32)
	if _, err := io.ReadFull(drbg, xBytes); err != nil {
		return nil, nil, err
	}
	xPriv, err := curve.NewPrivateKey(xBytes)
	if err != nil {
		return nil, nil, err
	}
	xPub := xPriv.PublicKey().Bytes()
	peerXPub, err := curve.NewPublicKey(xPubInit)
	if err != nil {
		return nil, nil, ErrHandshake
	}
	xShared, err := xPriv.ECDH(peerXPub)
	if err != nil {
		return nil, nil, ErrHandshake
	}
	var kyberPk kyber768.PublicKey
	if len(kyberPubInit) != kyberPubSize {
		return nil, nil, ErrHandshake
	}
	kyberPk.Unpack(kyberPubInit)
	kyberCt := make([]byte, kyberCtSize)
	kyberSS := make([]byte, kyberSSSize)
	kyberPk.EncapsulateTo(kyberCt, kyberSS, encSeed)

	respMsg := buildHandshakeRespMsg(Version, byte(mode), xPub, kyberCt)
	return handshakeRespFinish(mode, initMsg, respMsg, xShared, kyberSS)
}

func handshakeRespFinish(mode Mode, initMsg, respMsg, xShared, kyberSS []byte) ([]byte, *Session, error) {

	transcript := common.TranscriptHash(initMsg, respMsg, []byte{byte(mode)}, []byte{Version})
	sessionID := transcript

	kRaw := common.Extract(append(xShared, kyberSS...), transcript)
	kMs := common.Expand(kRaw, common.LabelMaster, 32)

	sess, err := newSessionFromKeys(mode, sessionID, transcript, kMs)
	if err != nil {
		return nil, nil, err
	}
	sess.initMsg = initMsg
	sess.respMsg = respMsg
	sess.transcriptHash = transcript
	sess.isInitiator = false
	sess.established = true

	return respMsg, sess, nil
}

// HandshakeComplete finishes handshake for initiator after receiving respMsg.
func (s *Session) HandshakeComplete(respMsg []byte) error {
	if s.established {
		return ErrHandshake
	}
	ver, m, xPubResp, kyberCt, err := parseHandshakeRespMsg(respMsg)
	if err != nil || ver != Version || m != byte(s.mode) {
		return ErrHandshake
	}

	curve := ecdh.X25519()
	peerXPub, err := curve.NewPublicKey(xPubResp)
	if err != nil {
		return ErrHandshake
	}
	xShared, err := s.xPriv.ECDH(peerXPub)
	if err != nil {
		return ErrHandshake
	}

	sk, ok := s.kyberPriv.(*kyber768.PrivateKey)
	if !ok {
		return ErrHandshake
	}
	kyberSS := make([]byte, kyberSSSize)
	sk.DecapsulateTo(kyberSS, kyberCt)

	transcript := common.TranscriptHash(s.initMsg, respMsg, []byte{byte(s.mode)}, []byte{Version})
	s.sessionID = transcript
	s.transcriptHash = transcript

	kRaw := common.Extract(append(xShared, kyberSS...), transcript)
	kMs := common.Expand(kRaw, common.LabelMaster, 32)

	s.deriveKeys(kMs)
	s.respMsg = respMsg
	s.isInitiator = true
	s.established = true
	s.xPriv = nil
	s.kyberPriv = nil
	return nil
}

func buildHandshakeInitMsg(ver, mode byte, xPub, kyberPub []byte) []byte {
	b := make([]byte, 3+x25519PubSize+kyberPubSize)
	b[0] = ver
	b[1] = mode
	b[2] = HandshakeTypeInit
	copy(b[3:], xPub)
	copy(b[3+x25519PubSize:], kyberPub)
	return b
}

func buildHandshakeRespMsg(ver, mode byte, xPub, kyberCt []byte) []byte {
	b := make([]byte, 3+x25519PubSize+kyberCtSize)
	b[0] = ver
	b[1] = mode
	b[2] = HandshakeTypeResp
	copy(b[3:], xPub)
	copy(b[3+x25519PubSize:], kyberCt)
	return b
}

func parseHandshakeInitMsg(b []byte) (ver, mode byte, xPub, kyberPub []byte, err error) {
	if len(b) < 3+x25519PubSize+kyberPubSize {
		return 0, 0, nil, nil, ErrHandshake
	}
	if b[2] != HandshakeTypeInit {
		return 0, 0, nil, nil, ErrHandshake
	}
	ver = b[0]
	mode = b[1]
	xPub = append([]byte(nil), b[3:3+x25519PubSize]...)
	kyberPub = append([]byte(nil), b[3+x25519PubSize:3+x25519PubSize+kyberPubSize]...)
	return ver, mode, xPub, kyberPub, nil
}

func parseHandshakeRespMsg(b []byte) (ver, mode byte, xPub, kyberCt []byte, err error) {
	if len(b) < 3+x25519PubSize+kyberCtSize {
		return 0, 0, nil, nil, ErrHandshake
	}
	if b[2] != HandshakeTypeResp {
		return 0, 0, nil, nil, ErrHandshake
	}
	ver = b[0]
	mode = b[1]
	xPub = append([]byte(nil), b[3:3+x25519PubSize]...)
	kyberCt = append([]byte(nil), b[3+x25519PubSize:3+x25519PubSize+kyberCtSize]...)
	return ver, mode, xPub, kyberCt, nil
}
