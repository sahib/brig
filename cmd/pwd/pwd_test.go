package pwd

import (
	"fmt"
	"testing"
	"time"

	zxcvbn "github.com/nbutton23/zxcvbn-go"
)

func TestLongPassword(t *testing.T) {
	start := time.Now()
	zxcvbn.PasswordStrength("68b329da9893e34099c7d8ad5cb9c940", nil)
	fmt.Println("Took", time.Since(start))
}

func BenchmarkLongPassword(b *testing.B) {
	fmt.Println("N", b.N)
	for i := 0; i < b.N; i++ {
		zxcvbn.PasswordStrength("1234567890123456", nil)
	}
}
