package protocol

import (
	"fmt"
)

// CommandType representa los tipos de comandos disponibles
type CommandType byte

const (
	// Comandos según documentación CSV
	CmdGetStatus                  CommandType = 0x10 // Status
	CmdResetLeftCounters          CommandType = 0x20 // Reset contador izquierda
	CmdResetRightCounters         CommandType = 0x21 // Reset contador derecha
	CmdRestartDevice              CommandType = 0x35 // Restart device (requiere 0x60)
	CmdLeftOpen                   CommandType = 0x80 // Abrir izquierda (Value)
	CmdLeftAlwaysOpen             CommandType = 0x81 // Siempre abierto izquierda
	CmdRightOpen                  CommandType = 0x82 // Abrir derecha (Value)
	CmdRightAlwaysOpen            CommandType = 0x83 // Siempre abierto derecha
	CmdCloseGate                  CommandType = 0x84 // Cerrar puerta
	CmdForbiddenLeftPassage       CommandType = 0x88 // Prohibir paso izquierda
	CmdForbiddenRightPassage      CommandType = 0x89 // Prohibir paso derecha
	CmdDisablePassageRestrictions CommandType = 0x8F // Deshabilitar restricciones
	CmdSetParameters              CommandType = 0x96 // Establecer parámetros (Data1 = Value)
)

// ResponseCode representa los códigos de respuesta del dispositivo
type ResponseCode byte

const (
	RespSuccess      ResponseCode = 0x00 // Comando ejecutado exitosamente
	RespError        ResponseCode = 0x01 // Error general
	RespInvalidCmd   ResponseCode = 0x02 // Comando inválido
	RespInvalidParam ResponseCode = 0x03 // Parámetro inválido
	RespDeviceBusy   ResponseCode = 0x04 // Dispositivo ocupado
	RespTimeout      ResponseCode = 0x05 // Timeout
)

// Command representa un comando para el torniquete
type Command struct {
	DeviceID byte        // ID del dispositivo destino
	Command  CommandType // Tipo de comando
	Data     []byte      // Datos del comando
}

// Response representa una respuesta del torniquete
type Response struct {
	DeviceID     byte         // ID del dispositivo que responde
	Command      CommandType  // Comando original
	ResponseCode ResponseCode // Código de respuesta
	Data         []byte       // Datos de respuesta
}

// Protocol constants según CSV
const (
	FrameHeader    = 0x7E // Starting Position
	FrameUndefined = 0x00 // Campo undefined
	FrameSize      = 8    // Tamaño fijo del frame
	DataSize       = 3    // 3 bytes de datos (Data 0, Data 1, Data 2)
	RestartParam   = 0x60 // Parámetro requerido para restart
)

// calculateTxChecksum implementa el algoritmo TX del documento
// Suma todos los bytes y aplica NOT (~ret)
func CalculateTxChecksum(data []byte) byte {
	var ret byte = 0
	for i := 0; i < len(data); i++ {
		ret += data[i]
	}
	ret = ^ret // NOT operation
	return ret
}

// validateRxChecksum implementa el algoritmo RX del documento
// Suma todos los bytes + 1, debe ser 0 si los datos son correctos
func ValidateRxChecksum(data []byte) bool {
	var ret byte = 0
	for i := 0; i < len(data); i++ {
		ret += data[i]
	}
	return (ret + 1) == 0
}

// BuildCommand construye un frame de comando según especificación CSV
func BuildCommand(deviceID byte, cmd CommandType, data []byte) ([]byte, error) {
	if len(data) > DataSize {
		return nil, fmt.Errorf("data too large: %d bytes (max %d)", len(data), DataSize)
	}

	// Frame structure: [Header][Undefined][MachineNumber][Command][Data0][Data1][Data2][Checksum]
	frame := make([]byte, FrameSize)

	frame[0] = FrameHeader    // 0x7E - Starting Position
	frame[1] = FrameUndefined // 0x00 - Undefined
	frame[2] = deviceID       // Machine Number
	frame[3] = byte(cmd)      // Command Value

	// Data bytes (3 bytes, pad with 0x00 if less)
	for i := 0; i < DataSize; i++ {
		if i < len(data) {
			frame[4+i] = data[i]
		} else {
			frame[4+i] = 0x00
		}
	}

	// Calculate checksum using algorithm from doc (exclude header and checksum position)
	checksum := CalculateTxChecksum(frame[1:7])
	frame[7] = checksum

	return frame, nil
}

// ParseResponse parsea una respuesta del dispositivo según formato CSV
func ParseResponse(data []byte) (*Response, error) {
	if len(data) < FrameSize {
		return nil, fmt.Errorf("frame too small: %d bytes (expected %d)", len(data), FrameSize)
	}

	// Verificar header
	if data[0] != FrameHeader {
		return nil, fmt.Errorf("invalid header: 0x%02X (expected 0x%02X)", data[0], FrameHeader)
	}

	// Verificar checksum usando algoritmo RX
	if !ValidateRxChecksum(data[1:]) {
		return nil, fmt.Errorf("checksum validation failed")
	}

	// Extract fields según estructura CSV
	// [Header][Undefined][MachineNumber][Command][Data0][Data1][Data2][Checksum]
	deviceID := data[2]             // Machine Number
	command := CommandType(data[3]) // Command Value

	// Para respuestas, asumimos que el command es el echo y ResponseCode está en Data0
	responseCode := ResponseCode(data[4]) // Data0 como código de respuesta

	// Extract response data (Data1 y Data2)
	responseData := make([]byte, 2)
	responseData[0] = data[5] // Data1
	responseData[1] = data[6] // Data2

	return &Response{
		DeviceID:     deviceID,
		Command:      command,
		ResponseCode: responseCode,
		Data:         responseData,
	}, nil
}

// String methods for better debugging
func (ct CommandType) String() string {
	switch ct {
	case CmdGetStatus:
		return "GetStatus"
	case CmdResetLeftCounters:
		return "ResetLeftCounters"
	case CmdResetRightCounters:
		return "ResetRightCounters"
	case CmdRestartDevice:
		return "RestartDevice"
	case CmdLeftOpen:
		return "LeftOpen"
	case CmdLeftAlwaysOpen:
		return "LeftAlwaysOpen"
	case CmdRightOpen:
		return "RightOpen"
	case CmdRightAlwaysOpen:
		return "RightAlwaysOpen"
	case CmdCloseGate:
		return "CloseGate"
	case CmdForbiddenLeftPassage:
		return "ForbiddenLeftPassage"
	case CmdForbiddenRightPassage:
		return "ForbiddenRightPassage"
	case CmdDisablePassageRestrictions:
		return "DisablePassageRestrictions"
	case CmdSetParameters:
		return "SetParameters"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(ct))
	}
}

func (rc ResponseCode) String() string {
	switch rc {
	case RespSuccess:
		return "Success"
	case RespError:
		return "Error"
	case RespInvalidCmd:
		return "InvalidCommand"
	case RespInvalidParam:
		return "InvalidParameter"
	case RespDeviceBusy:
		return "DeviceBusy"
	case RespTimeout:
		return "Timeout"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(rc))
	}
}
