package main

import (
	"encoding/hex"
	"fmt"
	"slices"

	r200 "github.com/ladecadence/GoR200"
)

var registered = map[string]int{
	"e28069150000501d63e8f8e4": 1,
	"e28069150000501d63e900e4": 2,
	"e28069150000401d63e904e4": 3,
	"e28069150000401d63e8fce4": 4,
}

func main() {
	rfid, err := r200.New("/dev/ttyUSB0", 115200, false)
	if err != nil {
		panic(err)
	}
	data, err := rfid.ReadTags()
	if err != nil {
		fmt.Println(err.Error())
	}
	tuppers := []int{}
	for _, d := range data {
		fmt.Printf("\tPC: 0x%0x\n", d.PC)
		fmt.Printf("\tEPC: %s\n", hex.EncodeToString(d.EPC))
		fmt.Printf("\tCRC: 0x%0x\n", d.CRC)
		if val, ok := registered[hex.EncodeToString(d.EPC)]; ok {
			fmt.Printf("\tTupper %d\n", val)
			if !slices.Contains(tuppers, val) {
				tuppers = append(tuppers, val)
			}
		}
	}
	fmt.Printf("Tuppers detected: %v\n", tuppers)
}
