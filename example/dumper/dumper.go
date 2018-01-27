package main

import (
	"github.com/marcinbor85/gohex"
	"os"
)

func main() {
	file, err := os.Create("output.hex")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	
	mem := gohex.NewMemory()
	mem.SetStartAddress(0x80008000)
	mem.AddBinary(0x10008000, []byte{0x01,0x02,0x03,0x04})
	mem.AddBinary(0x20000000, make([]byte, 256))
	
	mem.DumpIntelHex(file, 16)
}