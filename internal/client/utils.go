package client

func truncateBytes(b []byte, max int) []byte {
	if b == nil {
		return nil
	}

	if max > len(b) {
		return b
	}

	return b[:max]
}
