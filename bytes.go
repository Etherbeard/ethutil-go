package ethutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func NumberToBytes(num interface{}, bits int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	return buf.Bytes()[buf.Len()-(bits/8):]
}

func BytesToNumber(b []byte) (number uint64) {
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &number)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	return
}
