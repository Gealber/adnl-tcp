package tl

import (
	"encoding/binary"
)

// EncodeByteArray encode array according to
// https://docs.ton.org/develop/data-formats/tl#encoding-bytes-array
func EncodeByteArray(data []byte) []byte {
	dataSize := len(data)

	if dataSize < 254 {
		return encodeSmallArray(data)
	}

	return encodeLargeArray(data)
}

// Serialize a struct into array of bytes according to TL.
// https://core.telegram.org/mtproto/TL
func Serialize(obj any) []byte {
	result := make([]byte, 0)

	return result
}

func encodeLargeArray(data []byte) []byte {
	resultSize := len(data) + 4
	resultSize += (resultSize % 4)
	result := make([]byte, resultSize)

	// byte indicating is a large array
	result[0] = 0xFE
	// appending size as little endians
	binary.LittleEndian.PutUint64(result[1:], uint64(len(data)))
	copy(result[4:], data)

	return result
}

func encodeSmallArray(data []byte) []byte {
	resultSize := len(data) + 1
	resultSize += (4 - resultSize)
	result := make([]byte, resultSize)

	binary.LittleEndian.PutUint16(result, uint16(len(data)))
	copy(result[1:], data)

	return result
}
