package device

import (
	"context"
	"fmt"
	"time"

	"github.com/dumacp/ds205a/internal/protocol"
	"github.com/dumacp/ds205a/internal/rs485"
)

// New crea una nueva instancia del dispositivo DS205A
func New(config *Config) (*Device, error) {
	// Validar configuración
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	device := &Device{
		config: config,
		closed: true,
		logger: GetDefaultLogger(),
	}

	return device, nil
}

// NewWithLogger crea una nueva instancia con logger personalizado
func NewWithLogger(config *Config, logger Logger) (*Device, error) {
	device, err := New(config)
	if err != nil {
		return nil, err
	}
	device.logger = logger
	return device, nil
}

// Open abre la conexión con el dispositivo
func (d *Device) Open() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.closed {
		return nil // Ya está abierto
	}

	// Crear conexión RS485
	conn, err := rs485.NewConnection(&rs485.Config{
		Port:         d.config.Port,
		BaudRate:     d.config.BaudRate,
		DataBits:     d.config.DataBits,
		StopBits:     d.config.StopBits,
		Parity:       d.config.Parity,
		ReadTimeout:  d.config.ReadTimeout,
		WriteTimeout: d.config.WriteTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to open RS485 connection: %w", err)
	}

	if err := conn.Open(); err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	d.conn = conn
	d.closed = false

	d.logger.Info("Device opened successfully", "port", d.config.Port)
	return nil
}

// Close cierra la conexión con el dispositivo
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	var err error
	if d.conn != nil {
		err = d.conn.Close()
		d.conn = nil
	}

	d.closed = true
	d.logger.Info("Device closed")

	return err
}

// IsOpen retorna si el dispositivo está abierto
func (d *Device) IsOpen() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return !d.closed && d.conn != nil
}

// Write envía datos al dispositivo
func (d *Device) Write(data []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.conn == nil {
		return ErrDeviceNotOpen
	}

	d.logger.Debug("TX:", "data", fmt.Sprintf("[% 02X]", data))

	_, err := d.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// Read lee datos del dispositivo manejando fragmentación de tramas
func (d *Device) Read(ctx context.Context, buffer []byte) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.conn == nil {
		return 0, ErrDeviceNotOpen
	}

	// Buffer para acumular datos
	var accumulated []byte
	tempBuffer := make([]byte, 32) // Leer chunks más grandes

	// Leer datos hasta encontrar trama completa o timeout
	maxReadAttempts := 30

	initialByte := false

	for attempt := 0; attempt < maxReadAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
		n, err := d.conn.Read(tempBuffer)
		if err != nil {
			if n <= 0 && len(accumulated) == 0 {
				return len(accumulated), err
			}
		}

		if n > 0 {
			accumulated = append(accumulated, tempBuffer[:n]...)
			d.logger.Debug("Read chunk:", "bytes", n, "total", len(accumulated), "data", fmt.Sprintf("[% 02X]", tempBuffer[:n]))

			// Buscar header en los datos acumulados
			headerPos := -1
			if !initialByte {
				for i, b := range accumulated {
					if b == protocol.ResponseHeader {
						headerPos = i
						initialByte = true
						break
					}
				}
			}

			if headerPos >= 0 {
				// Encontramos el header, descartar datos anteriores
				if headerPos > 0 {
					d.logger.Debug("Discarding bytes before header:", "count", headerPos)
					accumulated = accumulated[headerPos:]
				}
			}

			// Verificar si tenemos la trama completa
			if initialByte && len(accumulated) >= protocol.ResponseSize {
				copy(buffer, accumulated[:protocol.ResponseSize])
				d.logger.Debug("Complete frame received:", "data", fmt.Sprintf("[% 02X]", buffer[:protocol.ResponseSize]))
				return protocol.ResponseSize, nil
			}
		}
	}

	// Si llegamos aquí, no se completó la trama
	if len(accumulated) > 0 {
		copy(buffer, accumulated)
		d.logger.Debug("Timeout with incomplete frame:", "received", len(accumulated), "expected", protocol.ResponseSize)
		return len(accumulated), fmt.Errorf("timeout: incomplete frame received %d bytes, expected %d", len(accumulated), protocol.ResponseSize)
	}

	d.logger.Debug("No data received")
	return 0, fmt.Errorf("timeout: no data received")
}

// SendCommand envía un comando y espera respuesta
func (d *Device) SendCommand(ctx context.Context, cmd protocol.CommandType, data []byte) (*protocol.Response, error) {
	if !d.IsOpen() {
		return nil, ErrDeviceNotOpen
	}

	// Construir comando
	frame, err := protocol.BuildCommand(d.config.DeviceID, cmd, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}

	// Enviar comando con reintentos
	var response *protocol.Response
	for attempt := 0; attempt <= d.config.RetryCount; attempt++ {
		if attempt > 0 {
			d.logger.Debug("Retrying command", "attempt", attempt, "command", cmd)
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}

		// Escribir comando
		if err := d.Write(frame); err != nil {
			d.logger.Warn("Failed to write command", "error", err)
			if attempt == d.config.RetryCount {
				return nil, fmt.Errorf("failed to send command after %d attempts: %w",
					d.config.RetryCount+1, err)
			}
			continue
		}

		// Leer respuesta
		responseBuffer := make([]byte, protocol.ResponseSize)
		n, err := d.Read(ctx, responseBuffer)
		if err != nil {
			if attempt == d.config.RetryCount {
				return nil, fmt.Errorf("failed to read response after %d attempts: %w",
					d.config.RetryCount+1, err)
			}
			continue
		}

		// Parsear respuesta con validación de Machine ID
		response, err = protocol.ParseResponse(responseBuffer[:n], d.config.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response after %d attempts: %w",
				d.config.RetryCount+1, err)
		}

		// Comando exitoso
		break
	}

	if response == nil {
		return nil, fmt.Errorf("failed to get valid response after %d attempts",
			d.config.RetryCount+1)
	}

	// La validación del código de respuesta ya se hace en ParseResponse
	return response, nil
}

// GetConfig retorna una copia de la configuración actual
func (d *Device) GetConfig() *Config {
	d.mu.RLock()
	defer d.mu.RUnlock()
	configCopy := *d.config
	return &configCopy
}

// validateConfig valida la configuración del dispositivo
func validateConfig(config *Config) error {
	if config.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	if config.BaudRate <= 0 {
		return fmt.Errorf("baud rate must be positive")
	}

	if config.DataBits < 5 || config.DataBits > 8 {
		return fmt.Errorf("data bits must be between 5 and 8")
	}

	if config.StopBits < 1 || config.StopBits > 2 {
		return fmt.Errorf("stop bits must be 1 or 2")
	}

	if config.Parity != "none" && config.Parity != "odd" && config.Parity != "even" {
		return fmt.Errorf("parity must be 'none', 'odd', or 'even'")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// GetStatus obtiene el estado actual del dispositivo
func (d *Device) GetStatus(ctx context.Context) (*Status, error) {
	response, err := d.SendCommand(ctx, protocol.CmdGetStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Convertir contadores de bytes a uint32
	leftCount := uint32(response.LeftPedestrianCount[0])<<16 |
		uint32(response.LeftPedestrianCount[1])<<8 |
		uint32(response.LeftPedestrianCount[2])

	rightCount := uint32(response.RightPedestrianCount[0])<<16 |
		uint32(response.RightPedestrianCount[1])<<8 |
		uint32(response.RightPedestrianCount[2])

	status := &Status{
		MachineNumber:        response.MachineNumber,
		VersionNumber:        response.VersionNumber,
		FaultEvent:           response.FaultEvent,
		GateStatus:           response.GateStatus,
		AlarmEvent:           response.AlarmEvent,
		InfraredStatus:       response.InfraredStatus,
		PowerSupplyVoltage:   response.PowerSupplyVoltage,
		LeftPedestrianCount:  leftCount,
		RightPedestrianCount: rightCount,
	}

	return status, nil
}

// LeftOpen abre el paso por la izquierda
func (d *Device) LeftOpen(ctx context.Context, value uint8) error {
	_, err := d.SendCommand(ctx, protocol.CmdLeftOpen, []byte{value})
	if err != nil {
		return fmt.Errorf("failed to open left passage: %w", err)
	}
	return nil
}

// LeftAlwaysOpen mantiene siempre abierto el paso izquierdo
func (d *Device) LeftAlwaysOpen(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdLeftAlwaysOpen, nil)
	if err != nil {
		return fmt.Errorf("failed to set left always open: %w", err)
	}
	return nil
}

// RightOpen abre el paso por la derecha
func (d *Device) RightOpen(ctx context.Context, value uint8) error {
	_, err := d.SendCommand(ctx, protocol.CmdRightOpen, []byte{value})
	if err != nil {
		return fmt.Errorf("failed to open right passage: %w", err)
	}
	return nil
}

// RightAlwaysOpen mantiene siempre abierto el paso derecho
func (d *Device) RightAlwaysOpen(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdRightAlwaysOpen, nil)
	if err != nil {
		return fmt.Errorf("failed to set right always open: %w", err)
	}
	return nil
}

// CloseGate cierra la puerta/torniquete
func (d *Device) CloseGate(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdCloseGate, nil)
	if err != nil {
		return fmt.Errorf("failed to close gate: %w", err)
	}
	return nil
}

// ForbiddenLeftPassage prohíbe el paso por la izquierda
func (d *Device) ForbiddenLeftPassage(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdForbiddenLeftPassage, nil)
	if err != nil {
		return fmt.Errorf("failed to forbid left passage: %w", err)
	}
	return nil
}

// ForbiddenRightPassage prohíbe el paso por la derecha
func (d *Device) ForbiddenRightPassage(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdForbiddenRightPassage, nil)
	if err != nil {
		return fmt.Errorf("failed to forbid right passage: %w", err)
	}
	return nil
}

// DisablePassageRestrictions deshabilita las restricciones de paso
func (d *Device) DisablePassageRestrictions(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdDisablePassageRestrictions, nil)
	if err != nil {
		return fmt.Errorf("failed to disable passage restrictions: %w", err)
	}
	return nil
}

// ResetLeftCounters resetea los contadores del lado izquierdo
func (d *Device) ResetLeftCounters(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdResetLeftCounters, nil)
	if err != nil {
		return fmt.Errorf("failed to reset left counters: %w", err)
	}
	return nil
}

// ResetRightCounters resetea los contadores del lado derecho
func (d *Device) ResetRightCounters(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdResetRightCounters, nil)
	if err != nil {
		return fmt.Errorf("failed to reset right counters: %w", err)
	}
	return nil
}

// GetDeviceInfo obtiene información del dispositivo
func (d *Device) GetDeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	// Usando el comando de status para obtener información básica
	response, err := d.SendCommand(ctx, protocol.CmdGetStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	info := &DeviceInfo{
		Version:     [3]uint8{response.VersionNumber, 0, 0}, // Usar VersionNumber de la respuesta
		MachineType: response.MachineNumber,                 // Usar el número de máquina como tipo
	}

	return info, nil
}

// Reset resetea el dispositivo
func (d *Device) Reset(ctx context.Context) error {
	_, err := d.SendCommand(ctx, protocol.CmdRestartDevice, []byte{0x60})
	if err != nil {
		return fmt.Errorf("failed to reset device: %w", err)
	}
	return nil
}

// SetParameters establece parámetros del dispositivo
func (d *Device) SetParameters(ctx context.Context, value uint8) error {
	_, err := d.SendCommand(ctx, protocol.CmdSetParameters, []byte{value})
	if err != nil {
		return fmt.Errorf("failed to set parameters: %w", err)
	}
	return nil
}
