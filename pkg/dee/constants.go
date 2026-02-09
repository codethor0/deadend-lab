package dee

const (
	Version   = 0x01
	ModeSafe  = 0x01
	ModeNaive = 0x02

	SessionIDSize = 32
	NonceSize     = 12
	RekeyEvery    = 1000

	HandshakeTypeInit = 0x01
	HandshakeTypeResp = 0x02

	// Header: version(1) + mode(1) + session_id(32) + counter(8) + flags(2) = 44
	HeaderSize = 44
	// Frame: header + payload_len(4) = 48
	FrameOverhead = 48
)
