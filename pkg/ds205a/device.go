package ds205a

import (
	"context"
	"time"

	"github.com/dumacp/ds205a/internal/device"
)

// Direction representa la dirección de paso
type Direction = device.Direction

// LogLevel representa el nivel de logging
type LogLevel = device.LogLevel

const (
	DirectionIn  = device.DirectionIn  // Entrada
	DirectionOut = device.DirectionOut // Salida
)

// Niveles de logging disponibles
const (
	LogLevelSilent = device.LogLevelSilent // Sin logs
	LogLevelError  = device.LogLevelError  // Solo errores
	LogLevelWarn   = device.LogLevelWarn   // Advertencias y errores
	LogLevelInfo   = device.LogLevelInfo   // Info, advertencias y errores
	LogLevelDebug  = device.LogLevelDebug  // Todos los logs
)

// PassageDirection representa la dirección de paso específica del dispositivo
type PassageDirection = device.PassageDirection

const (
	PassageDirectionNone  = device.PassageDirectionNone  // Sin dirección
	PassageDirectionEntry = device.PassageDirectionEntry // Entrada
	PassageDirectionExit  = device.PassageDirectionExit  // Salida
)

// Status representa el estado del dispositivo
type Status = device.Status

// DeviceInfo contiene información del dispositivo
type DeviceInfo = device.DeviceInfo

// Turnstile representa un dispositivo turnstile DS205A
type Turnstile struct {
	device *device.Device
}

// New crea una nueva instancia de Turnstile
func New(port string, machineNumber uint8, baudRate int, timeout time.Duration) (*Turnstile, error) {
	return NewWithLogLevel(port, machineNumber, baudRate, timeout, device.LogLevelSilent)
}

// NewWithLogLevel crea una nueva instancia de Turnstile con nivel de logging específico
func NewWithLogLevel(port string, machineNumber uint8, baudRate int, timeout time.Duration, logLevel device.LogLevel) (*Turnstile, error) {
	config := &device.Config{
		Port:         port,
		BaudRate:     baudRate,
		DataBits:     8,
		StopBits:     1,
		Parity:       "none",
		Timeout:      timeout,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DeviceID:     machineNumber,
		RetryCount:   3,
	}

	dev, err := device.NewWithLogger(config, device.GetLoggerWithLevel(logLevel))
	if err != nil {
		return nil, err
	}

	return &Turnstile{
		device: dev,
	}, nil
}

// Open abre la conexión con el dispositivo
func (t *Turnstile) Open() error {
	return t.device.Open()
}

// Close cierra la conexión con el dispositivo
func (t *Turnstile) Close() error {
	return t.device.Close()
}

// GetStatus obtiene el estado actual del dispositivo
func (t *Turnstile) GetStatus(ctx context.Context) (*Status, error) {
	return t.device.GetStatus(ctx)
}

// GetDeviceInfo obtiene información del dispositivo
func (t *Turnstile) GetDeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	return t.device.GetDeviceInfo(ctx)
}

// LeftOpen abre el paso por la izquierda (permite que el valor especifique parámetros)
func (t *Turnstile) LeftOpen(ctx context.Context, value uint8) error {
	return t.device.LeftOpen(ctx, value)
}

// LeftAlwaysOpen mantiene siempre abierto el paso izquierdo
func (t *Turnstile) LeftAlwaysOpen(ctx context.Context) error {
	return t.device.LeftAlwaysOpen(ctx)
}

// RightOpen abre el paso por la derecha (permite que el valor especifique parámetros)
func (t *Turnstile) RightOpen(ctx context.Context, value uint8) error {
	return t.device.RightOpen(ctx, value)
}

// RightAlwaysOpen mantiene siempre abierto el paso derecho
func (t *Turnstile) RightAlwaysOpen(ctx context.Context) error {
	return t.device.RightAlwaysOpen(ctx)
}

// CloseGate cierra la puerta/torniquete
func (t *Turnstile) CloseGate(ctx context.Context) error {
	return t.device.CloseGate(ctx)
}

// ForbiddenLeftPassage prohíbe el paso por la izquierda
func (t *Turnstile) ForbiddenLeftPassage(ctx context.Context) error {
	return t.device.ForbiddenLeftPassage(ctx)
}

// ForbiddenRightPassage prohíbe el paso por la derecha
func (t *Turnstile) ForbiddenRightPassage(ctx context.Context) error {
	return t.device.ForbiddenRightPassage(ctx)
}

// DisablePassageRestrictions deshabilita las restricciones de paso
func (t *Turnstile) DisablePassageRestrictions(ctx context.Context) error {
	return t.device.DisablePassageRestrictions(ctx)
}

// ResetLeftCounters resetea los contadores del lado izquierdo
func (t *Turnstile) ResetLeftCounters(ctx context.Context) error {
	return t.device.ResetLeftCounters(ctx)
}

// ResetRightCounters resetea los contadores del lado derecho
func (t *Turnstile) ResetRightCounters(ctx context.Context) error {
	return t.device.ResetRightCounters(ctx)
}

// Reset resetea el dispositivo
func (t *Turnstile) Reset(ctx context.Context) error {
	return t.device.Reset(ctx)
}

// SetParameters establece parámetros del dispositivo
func (t *Turnstile) SetParameters(ctx context.Context, value uint8) error {
	return t.device.SetParameters(ctx, value)
}
