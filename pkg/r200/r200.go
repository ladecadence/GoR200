package r200

import (
	"errors"

	"go.bug.st/serial"
)

const (
	R200_FrameHeader = 0xAA
	R200_FrameEnd    = 0xDD

	FrameType_Command      = 0x00
	FrameType_Response     = 0x01
	FrameType_Notification = 0x02

	R200_HeaderPos         = 0x00
	R200_TypePos           = 0x01
	R200_CommandPos        = 0x02
	R200_ParamLengthMSBPos = 0x03
	R200_ParamLengthLSBPos = 0x04
	R200_ParamPos          = 0x05

	CMD_GetModuleInfo                    = 0x03
	CMD_SinglePollInstruction            = 0x22
	CMD_MultiplePollInstruction          = 0x27
	CMD_StopMultiplePoll                 = 0x28
	CMD_SetSelectParameter               = 0x0C
	CMD_GetSelectParameter               = 0x0B
	CMD_SetSendSelectInstruction         = 0x12
	CMD_ReadLabel                        = 0x39
	CMD_WriteLabel                       = 0x49
	CMD_LockLabel                        = 0x82
	CMD_KillTag                          = 0x65
	CMD_GetQueryParameters               = 0x0D
	CMD_SetQueryParameters               = 0x0E
	CMD_SetWorkArea                      = 0x07
	CMD_SetWorkingChannel                = 0xAB
	CMD_GetWorkingChannel                = 0xAA
	CMD_SetAutoFrequencyHopping          = 0xAD
	CMD_AcquireTransmitPower             = 0xB7
	CMD_SetTransmitPower                 = 0xB6
	CMD_SetTransmitContinuousCarrier     = 0xB0
	CMD_GetReceiverDemodulatorParameters = 0xF1
	CMD_SetReceiverDemodulatorParameters = 0xF0
	CMD_TestRFInputBlockingSignal        = 0xF2
	CMD_TestChannelRSSI                  = 0xF3
	CMD_ControlIOPort                    = 0x1A
	CMD_ModuleSleep                      = 0x17
	CMD_SetModuleIdleSleepTime           = 0x1D
	CMD_ExecutionFailure                 = 0xFF

	ERR_CommandError  = 0x17
	ERR_FHSSFail      = 0x20
	ERR_InventoryFail = 0x15
	ERR_AccessFail    = 0x16
	ERR_ReadFail      = 0x09
	ERR_WriteFail     = 0x10
	ERR_LockFail      = 0x13
	ERR_KillFail      = 0x12
)

type R200Response struct {
	Type       uint8
	Command    uint8
	Checksum   uint8
	ChecksumOK bool
	params     []uint8
}

type R200PoolResponse struct {
	Rssi uint8
	PC   uint16
	EPC  []uint8
	CRC  uint16
}

func (r *R200PoolResponse) Parse(params []uint8) error {
	if len(params) < 17 {
		return errors.New("Not enough data")
	}
	r.Rssi = params[0]
	r.PC = (uint16(params[1]) << 8) + uint16(params[2])
	r.EPC = params[3:15]
	r.CRC = (uint16(params[15]) << 8) + uint16(params[16])
	return nil
}

type R200 interface {
	Close() error
}

type r200 struct {
	port  serial.Port
	debug bool
}

func New(port string, speed int, debug bool) (*r200, error) {
	r := r200{debug: debug}

	// prepare port
	mode := &serial.Mode{
		BaudRate: speed,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	// open port
	var err error
	r.port, err = serial.Open(port, mode)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *r200) Close() error {
	if r.port != nil {
		err := r.port.Close()
		return err
	} else {
		return errors.New("Serial port not opened")
	}
}
