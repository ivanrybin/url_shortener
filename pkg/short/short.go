package short

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
)

const urlAlpha = "0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Shortener interface {
	Short(s string) string
}

func New() Shortener {
	return &URLShortener{}
}

type URLShortener struct{}

// Short shorts URL to 10 characters string
func (s *URLShortener) Short(url string) string {
	digest := digest10(url)

	sb := strings.Builder{}
	sb.Grow(10)
	for _, b := range digest {
		sb.WriteByte(urlAlpha[b%uint32(len(urlAlpha))])
	}

	return sb.String()
}

// digest10 gets sha256(input) and reduces it to 10 integers hash
//
// Algorithm:
//  1. calculates sha256 from input
//  2. truncates 32 bytes to first 30 bytes
//  3. converts every 3 bytes (from 30 bytes) to 32 bits unsigned integer
//  4. returns 10 unsigned integers == string hash
func digest10(s string) []uint32 {
	sha := sha256.Sum256([]byte(s))
	digest := make([]uint32, 10)
	for i := 0; i <= 27; i += 3 {
		// dst = [0, sha[i], sha[i+1], sha[i+2]]
		dst := make([]byte, 4)
		copy(dst[1:], sha[i:i+3])
		digest[i/3] = binary.BigEndian.Uint32(dst)
	}
	return digest
}
