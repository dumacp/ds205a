package rs485

import (
	"time"
)

// Config contiene la configuración para la conexión RS485
type Config struct {
	Port         string        // Puerto serial
	BaudRate     int           // Velocidad de transmisión
	DataBits     int           // Bits de datos
	StopBits     int           // Bits de parada
	Parity       string        // Paridad
	ReadTimeout  time.Duration // Timeout de lectura
	WriteTimeout time.Duration // Timeout de escritura
	Debug        bool          // Habilitar debug de comunicación serial
}

// Connection representa una conexión RS485
type Connection struct {
	config *Config
	port   SerialPort
	closed bool
}

// SerialPort interface para abstracción del puerto serial
type SerialPort interface {
	Open() error
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Flush() error
	SetReadTimeout(timeout time.Duration) error
	SetWriteTimeout(timeout time.Duration) error
}

// NewConnection crea una nueva conexión RS485
func NewConnection(config *Config) (*Connection, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	// Crear puerto serial con la configuración especificada
	port, err := NewSerialPort(config)
	if err != nil {
		return nil, err
	}

	return &Connection{
		config: config,
		port:   port,
		closed: true,
	}, nil
}

// Open abre la conexión
func (c *Connection) Open() error {
	if !c.closed {
		return nil
	}

	if err := c.port.Open(); err != nil {
		return err
	}

	// Configurar timeouts
	if err := c.port.SetReadTimeout(c.config.ReadTimeout); err != nil {
		c.port.Close()
		return err
	}

	if err := c.port.SetWriteTimeout(c.config.WriteTimeout); err != nil {
		c.port.Close()
		return err
	}

	c.closed = false
	return nil
}

// Close cierra la conexión
func (c *Connection) Close() error {
	if c.closed {
		return nil
	}

	err := c.port.Close()
	c.closed = true
	return err
}

// Read lee datos de la conexión
func (c *Connection) Read(p []byte) (int, error) {
	if c.closed {
		return 0, ErrConnectionClosed
	}

	return c.port.Read(p)
}

// Write escribe datos a la conexión
func (c *Connection) Write(p []byte) (int, error) {
	if c.closed {
		return 0, ErrConnectionClosed
	}

	return c.port.Write(p)
}

// Flush limpia los buffers
func (c *Connection) Flush() error {
	if c.closed {
		return ErrConnectionClosed
	}

	return c.port.Flush()
}

// IsOpen retorna si la conexión está abierta
func (c *Connection) IsOpen() bool {
	return !c.closed
}
