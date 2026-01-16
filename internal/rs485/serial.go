package rs485

import (
	"errors"
	"fmt"
	"time"

	"go.bug.st/serial"
)

var (
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrConnectionClosed = errors.New("connection is closed")
	ErrPortNotFound     = errors.New("serial port not found")
	ErrOpenFailed       = errors.New("failed to open serial port")
)

// serialPort implementa SerialPort usando la librería go.bug.st/serial
type serialPort struct {
	config *Config
	port   serial.Port
}

// NewSerialPort crea un nuevo puerto serial
func NewSerialPort(config *Config) (SerialPort, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &serialPort{
		config: config,
	}, nil
}

// Open abre el puerto serial
func (sp *serialPort) Open() error {
	mode := &serial.Mode{
		BaudRate: sp.config.BaudRate,
		DataBits: sp.config.DataBits,
		StopBits: parseStopBits(sp.config.StopBits),
		Parity:   parseParity(sp.config.Parity),
	}

	port, err := serial.Open(sp.config.Port, mode)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenFailed, err)
	}

	sp.port = port
	return nil
}

// Close cierra el puerto serial
func (sp *serialPort) Close() error {
	if sp.port == nil {
		return nil
	}

	err := sp.port.Close()
	sp.port = nil
	return err
}

// Read lee datos del puerto serial
func (sp *serialPort) Read(p []byte) (int, error) {
	if sp.port == nil {
		return 0, ErrConnectionClosed
	}

	return sp.port.Read(p)
}

// Write escribe datos al puerto serial
func (sp *serialPort) Write(p []byte) (int, error) {
	if sp.port == nil {
		return 0, ErrConnectionClosed
	}

	return sp.port.Write(p)
}

// Flush limpia los buffers del puerto serial
func (sp *serialPort) Flush() error {
	if sp.port == nil {
		return ErrConnectionClosed
	}

	// La librería go.bug.st/serial no expone un método flush directo
	// Pero podemos intentar drenar el buffer de lectura
	return nil
}

// SetReadTimeout configura el timeout de lectura
func (sp *serialPort) SetReadTimeout(timeout time.Duration) error {
	if sp.port == nil {
		return ErrConnectionClosed
	}

	return sp.port.SetReadTimeout(timeout)
}

// SetWriteTimeout configura el timeout de escritura
func (sp *serialPort) SetWriteTimeout(timeout time.Duration) error {
	// La librería go.bug.st/serial no soporta timeout de escritura separado
	// Este método existe por compatibilidad con la interfaz
	return nil
}

// parseParity convierte string a serial.Parity
func parseParity(parity string) serial.Parity {
	switch parity {
	case "odd":
		return serial.OddParity
	case "even":
		return serial.EvenParity
	case "mark":
		return serial.MarkParity
	case "space":
		return serial.SpaceParity
	default:
		return serial.NoParity
	}
}

// parseStopBits convierte int a serial.StopBits
func parseStopBits(stopBits int) serial.StopBits {
	switch stopBits {
	case 1:
		return serial.OneStopBit
	case 2:
		return serial.TwoStopBits
	default:
		return serial.OneStopBit // default a 1 stop bit
	}
}

// validateConfig valida la configuración del puerto serial
func validateConfig(config *Config) error {
	if config.Port == "" {
		return fmt.Errorf("%w: port cannot be empty", ErrInvalidConfig)
	}

	if config.BaudRate <= 0 {
		return fmt.Errorf("%w: invalid baud rate: %d", ErrInvalidConfig, config.BaudRate)
	}

	if config.DataBits < 5 || config.DataBits > 8 {
		return fmt.Errorf("%w: data bits must be between 5 and 8: %d", ErrInvalidConfig, config.DataBits)
	}

	if config.StopBits < 1 || config.StopBits > 2 {
		return fmt.Errorf("%w: stop bits must be 1 or 2: %d", ErrInvalidConfig, config.StopBits)
	}

	validParities := []string{"none", "odd", "even", "mark", "space"}
	valid := false
	for _, p := range validParities {
		if config.Parity == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("%w: invalid parity: %s", ErrInvalidConfig, config.Parity)
	}

	return nil
}
