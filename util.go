package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

func BinToHex(bin []byte) (string, error) {
	ret := make([]byte, len(bin)*2)
	_ = hex.Encode(ret, bin)
	return string(ret), nil
}

func BinToStr(bin []byte) string {
	return fmt.Sprintf("%v", bin)
}

func HexToBin(hex string, size int) ([]byte, error) {
	result := &bytes.Buffer{}

	i := 0
	for i < len(hex) && size > 0 {
		b := hex[i : i+2]
		val, err := strconv.ParseUint(b, 16, 8)
		if err != nil {
			return nil, err
		}
		if err := binary.Write(result, binary.LittleEndian, uint8(val)); err != nil {
			return nil, err
		}
		i += 2
		size--
	}
	return result.Bytes(), nil
}
