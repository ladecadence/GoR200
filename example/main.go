package main

import (
	"encoding/hex"
	"fmt"

	r200 "github.com/ladecadence/GoR200"
)

var (
	rfidGainConfig = []uint8{r200.MIX_Gain_3dB, r200.IF_AMP_Gain_36dB, 0x00, 0xB0}
)

func main() {
	rfid, err := r200.New("/dev/ttyUSB0", 115200, 10, false)
	if err != nil {
		panic(err)
	}
	// configure RFID demodulator
	err = rfid.SendCommand(r200.CMD_SetReceiverDemodulatorParameters, rfidGainConfig)
	if err != nil {
		panic(err)
	}
	rcv, err := rfid.Receive()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", rcv)
	data, err := rfid.ReadTags()
	if err != nil {
		fmt.Println(err.Error())
	}
	tuppers := 0
	for _, d := range data {
		fmt.Printf("\tPC: 0x%0x\n", d.PC)
		fmt.Printf("\tEPC: %s\n", hex.EncodeToString(d.EPC))
		fmt.Printf("\tCRC: 0x%0x\n", d.CRC)
		tuppers++
	}
	fmt.Printf("Tuppers detected: %d\n", tuppers)
}
