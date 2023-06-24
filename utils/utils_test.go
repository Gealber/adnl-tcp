package utils_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/Gealber/adnl-tcp/utils"
)

func Test_AssemblyBytesSlices(t *testing.T) {
	t.Run("assembly_bytes_slices", func(t *testing.T) {
		a := make([]byte, 32)
		b := make([]byte, 32)
		_, err := rand.Read(a)
		if err != nil {
			t.Fatal(err)
		}

		_, err = rand.Read(b)
		if err != nil {
			t.Fatal(err)
		}

		c := make([]byte, 64)
		err = utils.AssemblyBytesSlices(c, a, b)

		for i := 0; i < 32; i++ {
			if c[i] != a[i] {
				err := errors.New(fmt.Sprintf("mismatch position for slice a in pos: %d", i))
				t.Fatal(err)
			}
		}

		for i := 32; i < 64; i++ {
			if c[i] != b[i-32] {
				err := errors.New(fmt.Sprintf("mismatch position for slice b in pos: %d", i))
				t.Fatal(err)
			}
		}
	})
}
