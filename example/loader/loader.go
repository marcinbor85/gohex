package main

import (
	"fmt"
	"github.com/marcinbor85/gohex"
	"os"
)

func main() {
	file, err := os.Open("example.hex")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	
	mem := gohex.NewMemory()
	err = mem.ParseIntelHex(file)
	if err != nil {
		panic(err)
	}
	for _, segment := range mem.GetDataSegments() {
		fmt.Printf("%+v\n", segment)
	}
	bytes := mem.ToBinary(0xFFF0, 128, 0x00)
	fmt.Printf("%v\n", bytes)
}
