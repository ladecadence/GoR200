package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ladecadence/GoR200/pkg/r200"
)

func main() {
	rfid, err := r200.New("/dev/ttyUSB1", 115200, true)
	if err != nil {
		panic(err)
	}
	rfid.SendCommand(r200.CMD_SinglePollInstruction, []uint8{})
	resp, err := rfid.Receive()
	for i, r := range resp {
		fmt.Printf("%d:\n", i+1)
		if r.Command == r200.CMD_SinglePollInstruction {
			pool := r200.R200PoolResponse{}
			pool.Parse(r.Params)
			fmt.Printf("PC: 0x%0x\n", pool.PC)
			fmt.Printf("PC: %s\n", hex.EncodeToString(pool.EPC))
			fmt.Printf("CRC: 0x%0x\n", pool.CRC)
		} else {
			fmt.Printf("\tParams: %s\n", hex.EncodeToString(r.Params))
			fmt.Printf("\tChecksum: 0x%0x\n", r.Checksum)
			fmt.Printf("\tChecksum OK: %t\n", r.ChecksumOK)
		}
	}
}
