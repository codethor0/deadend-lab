package dee

// Mode identifies DEE operational mode.
type Mode byte

const (
	Safe  Mode = ModeSafe
	Naive Mode = ModeNaive
)

func (m Mode) String() string {
	switch m {
	case Safe:
		return "SAFE"
	case Naive:
		return "NAIVE"
	default:
		return "UNKNOWN"
	}
}

// IsSafe returns true for DEE_SAFE mode.
func (m Mode) IsSafe() bool {
	return m == Safe
}
