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
	//rfid.SendCommand(r200.CMD_SetWorkArea, []uint8{0x02})
	rfid.SendCommand(r200.CMD_MultiplePollInstruction, []uint8{0x22, 0x00, 0x0a})
	resp, err := rfid.Receive()
	tuppers := []int{}
	for i, r := range resp {
		fmt.Printf("%d:\n", i+1)
		switch r.Command {
		case r200.CMD_SinglePollInstruction:
			pool := r200.R200PoolResponse{}
			pool.Parse(r.Params)
			fmt.Printf("\tPC: 0x%0x\n", pool.PC)
			fmt.Printf("\tEPC: %s\n", hex.EncodeToString(pool.EPC))
			fmt.Printf("\tCRC: 0x%0x\n", pool.CRC)
			if val, ok := registered[hex.EncodeToString(pool.EPC)]; ok {
				fmt.Printf("\tTupper %d\n", val)
				if !slices.Contains(tuppers, val) {
					tuppers = append(tuppers, val)
				}
			}
		case r200.CMD_ExecutionFailure:
			err := r200.R200ErrorResponse{Error: r.Params}
			fmt.Printf("\tError: %s\n", err.Parse())
		default:
			fmt.Printf("\tParams: %s\n", hex.EncodeToString(r.Params))
			fmt.Printf("\tChecksum: 0x%0x\n", r.Checksum)
			fmt.Printf("\tChecksum OK: %t\n", r.ChecksumOK)
		}
	}
	fmt.Printf("Tuppers detected: %v\n", tuppers)
}
