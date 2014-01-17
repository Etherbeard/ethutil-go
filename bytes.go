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

func BytesToNumber(b []byte) uint64 {
	var number uint64

	// Make sure the buffer is 64bits
	data := make([]byte, 8)
	data = append(data[:len(b)], b...)

	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &number)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Println("number", number)

	return number
}
