package internal

// ReverseBytes reverse slice content
func ReverseBytes(s []byte) []byte {
	for l, r := 0, len(s)-1; l < r; l, r = l+1, r-1 {
		s[l], s[r] = s[r], s[l]
	}
	return s
}
