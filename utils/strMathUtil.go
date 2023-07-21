package utils

import (
	"bytes"
	"encoding/binary"
)

func ToHexInt(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	HandleErr(err)
	return buff.Bytes()
}
