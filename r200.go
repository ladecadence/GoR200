package r200

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"time"

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
	CMD_GetWorkArea                      = 0x08
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

	MIX_Gain_0dB  = 0x00
	MIX_Gain_3dB  = 0x01
	MIX_Gain_6dB  = 0x02
	MIX_Gain_9dB  = 0x03
	MIX_Gain_12dB = 0x04
	MIX_Gain_15dB = 0x05
	MIX_Gain_16dB = 0x06

	IF_AMP_Gain_12dB = 0x00
	IF_AMP_Gain_18dB = 0x01
	IF_AMP_Gain_21dB = 0x02
	IF_AMP_Gain_24dB = 0x03
	IF_AMP_Gain_27dB = 0x04
	IF_AMP_Gain_30dB = 0x05
	IF_AMP_Gain_36dB = 0x06
	IF_AMP_Gain_40dB = 0x07
)

type R200Response struct {
	Type       uint8
	Command    uint8
	Checksum   uint8
	ChecksumOK bool
	Params     []uint8
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

type R200ErrorResponse struct {
	Error   []uint8
	Message string
}

func (r *R200ErrorResponse) Parse() string {
	switch r.Error[0] {
	case ERR_InventoryFail:
		r.Message = "No tags detected"
	case ERR_CommandError:
		r.Message = "Can't execute command"
	case ERR_ReadFail:
		r.Message = "Read failed"
	default:
		r.Message = fmt.Sprintf("Error: 0x%0x", r.Error)
	}
	return r.Message
}

type R200 interface {
	Close() error
	SendCommand(uint8, []uint8) error
	Receive() ([]R200Response, error)
	ReadTags() ([]R200PoolResponse, error)
}

type r200 struct {
	port      serial.Port
	num_reads uint8
	debug     bool
}

func New(port string, speed int, num_reads uint8, debug bool) (R200, error) {
	r := r200{debug: debug, num_reads: num_reads}

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
	r.port.SetReadTimeout(time.Second * 1)
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

func (r *r200) SendCommand(command uint8, parameters []uint8) error {
	out := []uint8{}
	// command packet
	out = append(out, R200_FrameHeader)
	out = append(out, FrameType_Command)
	out = append(out, command)
	// param length  and params
	param_len := len(parameters)
	out = append(out, uint8(param_len>>8), uint8(param_len))
	if len(parameters) > 0 {
		out = append(out, parameters...)
	}
	// checksum (from frame type to last parameter)
	sum := 0
	for i := R200_TypePos; i < R200_ParamPos+param_len; i++ {
		sum = sum + int(out[i])
	}
	out = append(out, uint8(sum))
	out = append(out, R200_FrameEnd)
	if r.debug {
		fmt.Printf("Sent: %s\n", hex.EncodeToString(out))
	}
	_, err := r.port.Write(out)
	return err
}

func (r *r200) Receive() ([]R200Response, error) {
	var responses []R200Response

	// try to read all data
	var buffer []uint8
	temp := make([]uint8, 512)
	num, err := r.port.Read(temp)
	buffer = append(buffer, temp[:num]...)
	if err != nil {
		return nil, err
	}
	for num > 0 {
		num, err = r.port.Read(temp)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, temp[:num]...)
	}

	// ok, try to parse all responses
	for len(buffer) > 0 {
		if r.debug {
			fmt.Printf("Buffer: %s\n", hex.EncodeToString(buffer))
		}
		resp := R200Response{}
		var param_len uint16 = 0
		if buffer[R200_HeaderPos] == R200_FrameHeader {
			// looks like a packet, but there can be seveal ones
			// check frame type
			if buffer[R200_TypePos] == FrameType_Response || buffer[R200_TypePos] == FrameType_Notification {
				resp.Type = buffer[R200_TypePos]
				resp.Command = buffer[R200_CommandPos]
				// ok, get param len and params
				param_len = (uint16(buffer[R200_ParamLengthMSBPos]) << 8) + uint16(buffer[R200_ParamLengthLSBPos])
				resp.Params = append(resp.Params, buffer[R200_ParamPos:R200_ParamPos+param_len]...)
				// checksum
				sum := 0
				for i := R200_TypePos; i < R200_ParamPos+int(param_len); i++ {
					sum += int(buffer[i])
				}
				if uint8(sum) == buffer[R200_ParamPos+param_len] {
					resp.ChecksumOK = true
					resp.Checksum = uint8(sum)
				}
				responses = append(responses, resp)
			}
		}
		// ok, remove this packet from array
		buffer = buffer[R200_ParamPos+param_len+2:]
	}

	return responses, nil
}

func (r *r200) ReadTags() ([]R200PoolResponse, error) {
	pool := []R200PoolResponse{}
	r.SendCommand(CMD_MultiplePollInstruction, []uint8{0x22, 0x00, r.num_reads})
	resp, err := r.Receive()
	if err != nil {
		return nil, err
	}
	// we can't compare the full struct, so compare the EDC hex string
	ids := []string{}
	for _, r := range resp {
		switch r.Command {
		case CMD_SinglePollInstruction:
			item := R200PoolResponse{}
			item.Parse(r.Params)
			// compare EPCs
			if !slices.Contains(ids, hex.EncodeToString(item.EPC)) {
				ids = append(ids, hex.EncodeToString(item.EPC))
				pool = append(pool, item)
			}
		case CMD_ExecutionFailure:
			errorData := R200ErrorResponse{Error: r.Params}
			errorData.Parse()
			err = fmt.Errorf("Error reading RFID: %s", errorData.Message)
		default:
			err = errors.New("Undefined error")
		}
	}
	return pool, err
}
