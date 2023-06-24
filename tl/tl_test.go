package tl_test

import (
	// "encoding/binary"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/Gealber/adnl-tcp/tl"
	// to check our implementation we are using tonutils-go
	externalTL "github.com/xssnick/tonutils-go/tl"
)

func Test_EncodeByteArray(t *testing.T) {
	t.Run("encode byte array simple dataSize < 254", func(t *testing.T) {
		data := []byte{0xAA, 0xBB}
		expectedResult := []byte{0x2, 0xAA, 0xBB, 0x00}

		result := tl.EncodeByteArray(data)

		err := errorCheck(result, expectedResult)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("encode byte array dataSize > 254", func(t *testing.T) {
		data := make([]byte, 254)
		_, err := rand.Read(data)
		if err != nil {
			t.Fatal(err)
		}

		expectedResult := externalTL.ToBytes(data)
		result := tl.EncodeByteArray(data)

		err = errorCheck(result, expectedResult)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func errorCheck(result, expectedResult []byte) error {
	if len(result) != len(expectedResult) {
		err := errors.New(fmt.Sprintf("mistmatch length result: %d expected: %d", len(result), len(expectedResult)))

		return err
	}

	for i := 0; i < len(result); i++ {
		if result[i] != expectedResult[i] {
			err := errors.New(fmt.Sprintf("mistmatch in position result:  %d expected: %d", result[i], expectedResult[i]))
			return err
		}
	}

	return nil
}
