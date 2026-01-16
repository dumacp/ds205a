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
	RespSuccess      ResponseCode = 0x55 // Comando ejecutado exitosamente (Command Execution)
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

// Response representa una respuesta del torniquete según reponse.csv
type Response struct {
	StartPosition        byte    // Starting Position (0x7F)
	VersionNumber        byte    // Version Number
	MachineNumber        byte    // Machine Number (debe coincidir con comando)
	FaultEvent           byte    // Fault Event
	GateStatus           byte    // Gate Status
	AlarmEvent           byte    // Alarm Event
	LeftPedestrianCount  [3]byte // Cumulative Number of Pedestrians on the Left (3 bytes)
	RightPedestrianCount [3]byte // Cumulative Number of Pedestrians on the Right (3 bytes)
	InfraredStatus       byte    // Infrared Status
	CommandExecution     byte    // Command Execution (0x55 = success)
	PowerSupplyVoltage   byte    // Power Supply Voltage
	Undefined1           byte    // Undefined
	Undefined2           byte    // Undefined
	Checksum             byte    // Checksum
}

// GetLeftCount convierte los 3 bytes del contador izquierdo a uint32
func (r *Response) GetLeftCount() uint32 {
	return uint32(r.LeftPedestrianCount[0])<<16 |
		uint32(r.LeftPedestrianCount[1])<<8 |
		uint32(r.LeftPedestrianCount[2])
}

// GetRightCount convierte los 3 bytes del contador derecho a uint32
func (r *Response) GetRightCount() uint32 {
	return uint32(r.RightPedestrianCount[0])<<16 |
		uint32(r.RightPedestrianCount[1])<<8 |
		uint32(r.RightPedestrianCount[2])
}

// IsSuccess verifica si la respuesta indica éxito
func (r *Response) IsSuccess() bool {
	return r.CommandExecution == byte(RespSuccess)
}

// Protocol constants según CSV
const (
	FrameHeader      = 0x7E // Starting Position para comandos
	ResponseHeader   = 0x7F // Starting Position para respuestas
	FrameUndefined   = 0x00 // Campo undefined
	FrameSize        = 8    // Tamaño fijo del frame de comando
	ResponseSize     = 18   // Tamaño fijo del frame de respuesta
	DataSize         = 3    // 3 bytes de datos (Data 0, Data 1, Data 2)
	RestartParam     = 0x60 // Parámetro requerido para restart
	SuccessExecution = 0x55 // Command Execution value para éxito
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
	frame := make([]byte, 0)

	frame = append(frame, FrameHeader)    // 0x7E - Starting Position
	frame = append(frame, FrameUndefined) // 0x00 - Undefined
	frame = append(frame, deviceID)       // Machine Number
	frame = append(frame, byte(cmd))      // Command Value

	// Data bytes (3 bytes, pad with 0x00 if less)
	for i := 0; i < DataSize; i++ {
		if i < len(data) {
			frame = append(frame, data[i])
		} else {
			frame = append(frame, 0x00)
		}
	}

	// Calculate checksum using algorithm from doc (exclude header and checksum position)
	checksum := CalculateTxChecksum(frame[0:])
	frame = append(frame, checksum)

	return frame, nil
}

// ParseResponse parsea una respuesta del dispositivo según reponse.csv
func ParseResponse(data []byte, expectedMachineID byte) (*Response, error) {
	if len(data) < ResponseSize {
		return nil, fmt.Errorf("response frame too small: %d bytes (expected %d)", len(data), ResponseSize)
	}

	// Verificar header de respuesta
	if data[0] != ResponseHeader {
		return nil, fmt.Errorf("invalid response header: 0x%02X (expected 0x%02X)", data[0], ResponseHeader)
	}

	// // Verificar checksum usando algoritmo RX (todos los bytes excepto el primer header)
	// if !ValidateRxChecksum(data[1:]) {
	// 	return nil, fmt.Errorf("checksum validation failed")
	// }

	// Extraer campos según reponse.csv
	response := &Response{
		StartPosition:      data[0],  // Starting Position (0x7F)
		VersionNumber:      data[1],  // Version Number
		MachineNumber:      data[2],  // Machine Number
		FaultEvent:         data[3],  // Fault Event
		GateStatus:         data[4],  // Gate Status
		AlarmEvent:         data[5],  // Alarm Event
		InfraredStatus:     data[12], // Infrared Status (posición 12)
		CommandExecution:   data[13], // Command Execution (posición 13)
		PowerSupplyVoltage: data[14], // Power Supply Voltage (posición 14)
		Undefined1:         data[15], // Placeholder para mantener compatibilidad
		Undefined2:         data[16], // Placeholder para mantener compatibilidad
		Checksum:           data[17], // Checksum (último byte del frame de 18)
	}

	// Extraer contadores de 3 bytes cada uno (6 bytes contiguos: posiciones 6-11)
	copy(response.LeftPedestrianCount[:], data[6:9])   // Bytes 6,7,8
	copy(response.RightPedestrianCount[:], data[9:12]) // Bytes 9,10,11

	// Verificar que el Machine Number coincida
	if response.MachineNumber != expectedMachineID {
		return nil, fmt.Errorf("machine ID mismatch: got 0x%02X, expected 0x%02X",
			response.MachineNumber, expectedMachineID)
	}

	// Verificar que el comando se ejecutó exitosamente
	if response.CommandExecution != SuccessExecution {
		return nil, fmt.Errorf("command execution failed: 0x%02X (expected 0x%02X)",
			response.CommandExecution, SuccessExecution)
	}

	return response, nil
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
