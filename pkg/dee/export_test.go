package dee

// SetRekeyEveryForTest sets rekey interval for testing. Call with 0 to reset.
func SetRekeyEveryForTest(n uint64) {
	rekeyEveryForTest = n
}
