package short

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"
)
import "github.com/stretchr/testify/assert"

func TestURLShortener_Short(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	alpha := "abcdefghijklmnopqrstuvwxyz./:_-&+^%$#@?=123456789"
	outLen := 10
	rndStringMaxLen := 30

	shortener := New()

	re := regexp.MustCompile(`[_a-zA-Z0-9]+`)
	for i := 0; i < 100_000; i++ {
		s := shortener.Short(genRandomString(rand.Int()%rndStringMaxLen, alpha))

		assert.Equal(t, outLen, len(s))
		assert.True(t, re.MatchString(s))
	}
}

func TestHash10_len(t *testing.T) {
	for n := 0; n < 100; n++ {
		hash := digest10(strings.Repeat("a", n))

		assert.Equal(t, 10, len(hash))
	}
}

func genRandomString(n int, alpha string) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(alpha[rand.Int()%len(alpha)])
	}
	return sb.String()
}

func benchHashString(s string, b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = digest10(s)
	}
}

func BenchmarkGenRandomString10(b *testing.B) {
	benchHashString(strings.Repeat("a", 10), b)
}

func BenchmarkGenRandomString100(b *testing.B) {
	benchHashString(strings.Repeat("a", 100), b)
}

func BenchmarkGenRandomString1000(b *testing.B) {
	benchHashString(strings.Repeat("a", 1000), b)
}
