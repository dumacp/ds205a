package device

import (
	"errors"
	"sync"
	"time"

	"github.com/dumacp/ds205a/pkg/rs485"
)

var (
	ErrDeviceClosed    = errors.New("device is closed")
	ErrDeviceNotOpen   = errors.New("device is not open")
	ErrTimeout         = errors.New("operation timeout")
	ErrInvalidResponse = errors.New("invalid response from device")
	ErrCommunication   = errors.New("communication error")
	ErrInvalidDeviceID = errors.New("invalid device ID")
)

// Device representa la implementación interna del dispositivo DS205A
type Device struct {
	mu     sync.RWMutex
	conn   *rs485.Connection
	config *Config
	closed bool
	logger Logger
}

// Config contiene la configuración del dispositivo DS205A
type Config struct {
	Port         string        // Puerto serial (ej: "/dev/ttyUSB0")
	BaudRate     int           // Velocidad de transmisión (default: 9600)
	DataBits     int           // Bits de datos (default: 8)
	StopBits     int           // Bits de parada (default: 1)
	Parity       string        // Paridad: "none", "odd", "even" (default: "none")
	Timeout      time.Duration // Timeout de operaciones (default: 5s)
	ReadTimeout  time.Duration // Timeout de lectura (default: 2s)
	WriteTimeout time.Duration // Timeout de escritura (default: 2s)
	DeviceID     byte          // ID del dispositivo (default: 0x01)
	RetryCount   int           // Número de reintentos (default: 3)
}

// Logger interface para logging personalizable
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// defaultLogger implementación básica de logger
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {}
func (l *defaultLogger) Info(msg string, args ...interface{})  {}
func (l *defaultLogger) Warn(msg string, args ...interface{})  {}
func (l *defaultLogger) Error(msg string, args ...interface{}) {}

// GetDefaultLogger retorna el logger por defecto
func GetDefaultLogger() Logger {
	return &defaultLogger{}
}

// Direction representa la dirección de paso
type Direction int

const (
	DirectionIn  Direction = iota // Entrada
	DirectionOut                  // Salida
)

// PassageDirection representa la dirección de paso específica del dispositivo
type PassageDirection int

const (
	PassageDirectionNone  PassageDirection = iota // Sin dirección
	PassageDirectionEntry                         // Entrada
	PassageDirectionExit                          // Salida
)

// Status representa el estado del dispositivo según respuesta de 16 bytes
type Status struct {
	MachineNumber        uint8  // Número de máquina
	VersionNumber        uint8  // Número de versión
	FaultEvent           uint8  // Evento de falla
	GateStatus           uint8  // Estado de la puerta
	AlarmEvent           uint8  // Evento de alarma
	InfraredStatus       uint8  // Estado infrarrojo
	PowerSupplyVoltage   uint8  // Voltaje de alimentación
	LeftPedestrianCount  uint32 // Contador de peatones izquierda (3 bytes convertidos a uint32)
	RightPedestrianCount uint32 // Contador de peatones derecha (3 bytes convertidos a uint32)
}

// DeviceInfo contiene información del dispositivo
type DeviceInfo struct {
	Version     [3]uint8 // Versión del firmware [major, minor, patch]
	MachineType uint8    // Tipo de máquina
}
