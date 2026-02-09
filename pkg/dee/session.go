package dee

import (
	"crypto/ecdh"
	"encoding/binary"
	"errors"

	"deadend-lab/pkg/common"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrDecrypt     = errors.New("decryption failed")
	ErrReplay      = errors.New("decryption failed")
	ErrCounter     = errors.New("decryption failed")
	ErrInvalidMode = errors.New("invalid mode")

	rekeyEveryForTest uint64
)

// Session holds DEE session state.
type Session struct {
	mode           Mode
	sessionID      []byte
	transcriptHash []byte
	kAead          []byte
	kNonce         []byte
	kAudit         []byte
	kRekey         []byte
	kMs            []byte
	counterTx      uint64
	counterRx      uint64
	initMsg        []byte
	respMsg        []byte
	established    bool
	isInitiator    bool

	// Pre-handshake (cleared after)
	xPriv     *ecdh.PrivateKey
	kyberPriv interface{}
}

func newSessionFromKeys(mode Mode, sessionID, transcriptHash, kMs []byte) (*Session, error) {
	s := &Session{
		mode:           mode,
		sessionID:      append([]byte(nil), sessionID...),
		transcriptHash: append([]byte(nil), transcriptHash...),
		kMs:            append([]byte(nil), kMs...),
		counterRx:      0,
		established:    false,
	}
	s.deriveKeys(kMs)
	return s, nil
}

func (s *Session) deriveKeys(kMs []byte) {
	s.kAead = common.Expand(kMs, common.LabelAEADKey, 32)
	s.kNonce = common.Expand(kMs, common.LabelNonceBase, 32)
	s.kAudit = common.Expand(kMs, common.LabelAuditTag, 32)
	s.kRekey = common.Expand(kMs, common.LabelRekey, 32)
}

// SessionID returns the session identifier.
func (s *Session) SessionID() []byte {
	return append([]byte(nil), s.sessionID...)
}

// WireHeader returns the wire header for the given counter. Used for building AD.
func (s *Session) WireHeader(counter uint64) []byte {
	return s.buildHeader(counter, 0)
}

// Encrypt encrypts plaintext with optional associated data.
func (s *Session) Encrypt(plaintext, ad []byte) (ciphertext []byte, err error) {
	if !s.established {
		return nil, ErrDecrypt
	}

	s.maybeRekey()

	var nonce []byte
	if s.mode.IsSafe() {
		nonce = s.deriveNonce(ad)
	} else {
		nonce = make([]byte, NonceSize)
		binary.BigEndian.PutUint64(nonce[4:], s.counterTx)
	}

	aead, err := chacha20poly1305.New(s.kAead)
	if err != nil {
		return nil, err
	}

	header := s.buildHeader(s.counterTx, 0)
	additionalData := append(header, ad...)
	ct := aead.Seal(nil, nonce, plaintext, additionalData)

	if s.mode.IsSafe() {
		auditInput := common.TranscriptHash(s.transcriptHash, header, uint64ToBytes(s.counterTx))
		auditTag := common.HMAC256Truncate(s.kAudit, auditInput, 16)
		ciphertext = make([]byte, 16+len(ct))
		copy(ciphertext, auditTag)
		copy(ciphertext[16:], ct)
	} else {
		ciphertext = ct
	}

	s.counterTx++
	return ciphertext, nil
}

// EncryptNaiveWithNonce allows caller-supplied nonce in NAIVE mode only.
func (s *Session) EncryptNaiveWithNonce(plaintext, ad, callerNonce []byte) (ciphertext []byte, err error) {
	if !s.established || s.mode.IsSafe() {
		return nil, ErrDecrypt
	}
	if len(callerNonce) != NonceSize {
		return nil, ErrDecrypt
	}

	aead, err := chacha20poly1305.New(s.kAead)
	if err != nil {
		return nil, err
	}
	header := s.buildHeader(s.counterTx, 0)
	additionalData := append(header, ad...)
	ciphertext = aead.Seal(nil, callerNonce, plaintext, additionalData)
	s.counterTx++
	return ciphertext, nil
}

// Decrypt decrypts ciphertext. ad must contain the full header (44 bytes) then optional user AD.
func (s *Session) Decrypt(ciphertext, ad []byte) (plaintext []byte, err error) {
	if !s.established {
		return nil, ErrDecrypt
	}

	aead, err := chacha20poly1305.New(s.kAead)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aead.Overhead() {
		return nil, ErrDecrypt
	}
	if len(ad) < HeaderSize {
		return nil, ErrDecrypt
	}
	header := ad[:HeaderSize]
	counter := binary.BigEndian.Uint64(header[34:42])
	actualAD := ad[HeaderSize:]

	if s.mode.IsSafe() {
		// Strict monotonic: only accept counter == expectedRx, then increment by 1.
		if counter != s.counterRx {
			return nil, ErrDecrypt
		}
		auditInput := common.TranscriptHash(s.transcriptHash, header, uint64ToBytes(counter))
		expectedAudit := common.HMAC256Truncate(s.kAudit, auditInput, 16)
		if len(ciphertext) < 16+aead.Overhead() {
			return nil, ErrDecrypt
		}
		gotAudit := ciphertext[:16]
		ct := ciphertext[16:]
		if !common.EqualConstantTime(gotAudit, expectedAudit) {
			return nil, ErrDecrypt
		}
		nonce := s.deriveNonceForCounter(counter, actualAD)
		additionalData := append(header, actualAD...)
		plaintext, err = aead.Open(nil, nonce, ct, additionalData)
	} else {
		nonce := make([]byte, NonceSize)
		binary.BigEndian.PutUint64(nonce[4:], counter)
		additionalData := append(header, actualAD...)
		plaintext, err = aead.Open(nil, nonce, ciphertext, additionalData)
	}
	if err != nil {
		return nil, ErrDecrypt
	}
	s.counterRx++
	s.maybeRekeyRx(s.counterRx)
	return plaintext, nil
}

// DecryptFromFrame parses a framed message and decrypts.
func (s *Session) DecryptFromFrame(frame []byte) (plaintext []byte, err error) {
	if len(frame) < FrameOverhead {
		return nil, ErrDecrypt
	}
	header := frame[:HeaderSize]
	payloadLen := binary.BigEndian.Uint32(frame[44:48])
	if uint64(len(frame)) < 48+uint64(payloadLen) {
		return nil, ErrDecrypt
	}
	payload := frame[48 : 48+payloadLen]
	return s.Decrypt(payload, header)
}

// EncryptToFrame encrypts and returns a full frame.
func (s *Session) EncryptToFrame(plaintext, ad []byte) (frame []byte, err error) {
	ct, err := s.Encrypt(plaintext, ad)
	if err != nil {
		return nil, err
	}
	header := s.buildHeader(s.counterTx-1, 0)
	frame = make([]byte, 48+len(ct))
	copy(frame, header)
	binary.BigEndian.PutUint32(frame[44:48], uint32(len(ct)))
	copy(frame[48:], ct)
	return frame, nil
}

func (s *Session) buildHeader(counter uint64, flags uint16) []byte {
	b := make([]byte, HeaderSize)
	b[0] = Version
	b[1] = byte(s.mode)
	copy(b[2:34], s.sessionID)
	binary.BigEndian.PutUint64(b[34:42], counter)
	binary.BigEndian.PutUint16(b[42:44], flags)
	return b
}

func (s *Session) deriveNonce(ad []byte) []byte {
	return s.deriveNonceForCounter(s.counterTx, ad)
}

func (s *Session) deriveNonceForCounter(counter uint64, ad []byte) []byte {
	adHash := common.HashSHA256(ad)
	counterBytes := uint64ToBytes(counter)
	input := common.TranscriptHash(s.sessionID, s.transcriptHash, counterBytes, adHash)
	return common.HMAC256Truncate(s.kNonce, input, NonceSize)
}

func (s *Session) rekeyInterval() uint64 {
	if rekeyEveryForTest > 0 {
		return rekeyEveryForTest
	}
	return RekeyEvery
}

func (s *Session) maybeRekey() {
	n := s.rekeyInterval()
	if n == 0 {
		return
	}
	if s.counterTx > 0 && s.counterTx%n == 0 {
		s.ratchetForward(s.counterTx)
	}
}

func (s *Session) maybeRekeyRx(counterRx uint64) {
	n := s.rekeyInterval()
	if n == 0 {
		return
	}
	// Rekey after receiving last message of a block; both sides use same boundary.
	if counterRx > 0 && counterRx%n == 0 {
		s.ratchetForward(counterRx)
	}
}

func (s *Session) ratchetForward(counter uint64) {
	ctrBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(ctrBytes, counter)
	info := common.LabelRekeyRatchet + string(ctrBytes)
	s.kMs = common.Expand(s.kRekey, info, 32)
	s.deriveKeys(s.kMs)
}

func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
