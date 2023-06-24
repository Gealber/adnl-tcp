package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/oasisprotocol/curve25519-voi/curve"
	ed25519crv "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/oasisprotocol/curve25519-voi/primitives/x25519"
	"github.com/xssnick/tonutils-go/tl"
)

var (
	ErrInvalidBytesSize = errors.New("invalid bytes size")
)

// GenerateRandomBytes shamelessly copied from
// https://blog.argcv.com/articles/5992.c
func GenerateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)

	return b, err
}

// InToIPV4 shamelessly copied from
// https://socketloop.com/tutorials/golang-convert-decimal-number-integer-to-ipv4-address
func IntToIPV4(ipInt int64) string {
	// need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)

	return b0 + "." + b1 + "." + b2 + "." + b3
}

func AssemblyBytesSlices(dst []byte, srcs ...[]byte) error {
	srcLen := 0
	for _, s := range srcs {
		srcLen += len(s)
	}

	if len(dst) < srcLen {
		return ErrInvalidBytesSize
	}

	lastWrittenByte := 0
	for _, src := range srcs {
		n := copy(dst[lastWrittenByte:], src)
		lastWrittenByte += n
	}

	return nil
}

// sharedKey generate encryption key based on our and server key, ECDH algorithm
// copied without issues from https://docs.ton.org/develop/network/adnl-tcp#getting-a-shared-key-using-ecdh
func SharedKey(ourKey ed25519.PrivateKey, serverKey ed25519.PublicKey) ([]byte, error) {
	comp, err := curve.NewCompressedEdwardsYFromBytes(serverKey)
	if err != nil {
		return nil, err
	}

	ep, err := curve.NewEdwardsPoint().SetCompressedY(comp)
	if err != nil {
		return nil, err
	}

	mp := curve.NewMontgomeryPoint().SetEdwards(ep)
	bb := x25519.EdPrivateKeyToX25519(ed25519crv.PrivateKey(ourKey))

	key, err := x25519.X25519(bb, mp[:])
	if err != nil {
		return nil, err
	}

	return key, nil
}

// also shamelessly copied from
// https://github.com/xssnick/tonutils-go/blob/2b5e5a0e6ceaf3f28309b0833cb45de81c580acc/liteclient/crypto.go#L16
func KeyID(key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key not 32 bytes")
	}

	// https://github.com/ton-blockchain/ton/blob/24dc184a2ea67f9c47042b4104bbb4d82289fac1/crypto/block/check-proof.cpp#L488
	// what the hell is this?
	magic := []byte{0xc6, 0xb4, 0x13, 0x48}
	hash := sha256.New()
	hash.Write(magic)
	hash.Write(key)
	s := hash.Sum(nil)

	return s, nil
}

// NewCTRCipher ...
func NewCTRCipher(key, iv []byte) (cipher.Stream, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(blk, iv), nil
}

func ToKeyID(key any) ([]byte, error) {
	data, err := tl.Serialize(key, true)
	if err != nil {
		return nil, fmt.Errorf("key serialize err: %w", err)
	}

	hash := sha256.New()
	hash.Write(data)
	s := hash.Sum(nil)

	return s, nil
}
