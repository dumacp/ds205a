package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dumacp/ds205a/pkg/ds205a"
)

func main() {
	fmt.Println("DS205A Turnstile Basic Example")
	fmt.Println("==============================")

	// Crear dispositivo con nueva API (puerto, ID, baudrate, timeout)
	// Para habilitar debug, usar: ds205a.NewWithDebug("/dev/ttyUSB0", 0x01, 9600, 5*time.Second, true)
	device, err := ds205a.New("/dev/ttyUSB0", 0x01, 9600, 5*time.Second)
	if err != nil {
		log.Fatalf("Error creating device: %v", err)
	}

	// Abrir conexión
	fmt.Println("Opening connection...")
	if err := device.Open(); err != nil {
		log.Fatalf("Error opening device: %v", err)
	}
	defer device.Close()

	ctx := context.Background()

	// Obtener información del dispositivo
	fmt.Println("\nGetting device info...")
	deviceInfo, err := device.GetDeviceInfo(ctx)
	if err != nil {
		log.Printf("Warning: Could not get device info: %v", err)
	} else {
		fmt.Printf("Version: %d.%d.%d\n", deviceInfo.Version[0], deviceInfo.Version[1], deviceInfo.Version[2])
		fmt.Printf("Machine Type: %d\n", deviceInfo.MachineType)
	}

	// Obtener estado inicial
	fmt.Println("\nGetting initial status...")
	status, err := device.GetStatus(ctx)
	if err != nil {
		log.Printf("Warning: Could not get status: %v", err)
	} else {
		fmt.Printf("Machine Number: %d\n", status.MachineNumber)
		fmt.Printf("Version Number: %d\n", status.VersionNumber)
		fmt.Printf("Fault Event: 0x%02X\n", status.FaultEvent)
		fmt.Printf("Gate Status: 0x%02X\n", status.GateStatus)
		fmt.Printf("Alarm Event: 0x%02X\n", status.AlarmEvent)
		fmt.Printf("Infrared Status: 0x%02X\n", status.InfraredStatus)
		fmt.Printf("Power Supply Voltage: %d\n", status.PowerSupplyVoltage)
		fmt.Printf("Left Pedestrian Count: %d\n", status.LeftPedestrianCount)
		fmt.Printf("Right Pedestrian Count: %d\n", status.RightPedestrianCount)
	}

	// Deshabilitar restricciones (permitir paso libre)
	fmt.Println("\nDisabling passage restrictions...")
	if err := device.DisablePassageRestrictions(ctx); err != nil {
		log.Printf("Warning: Could not disable restrictions: %v", err)
	} else {
		fmt.Println("Passage restrictions disabled")
	}

	// Abrir paso izquierdo (entrada) con valor 1
	fmt.Println("\nOpening left passage...")
	if err := device.LeftOpen(ctx, 0x01); err != nil {
		log.Printf("Warning: Could not open left passage: %v", err)
	} else {
		fmt.Println("Left passage opened")
	}

	// Esperar un poco
	time.Sleep(20 * time.Second)

	// Abrir paso derecho (salida) con valor 1
	fmt.Println("\nOpening right passage...")
	if err := device.RightOpen(ctx, 0x01); err != nil {
		log.Printf("Warning: Could not open right passage: %v", err)
	} else {
		fmt.Println("Right passage opened")
	}

	// Esperar un poco
	time.Sleep(20 * time.Second)

	// Cerrar la puerta
	fmt.Println("\nClosing gate...")
	if err := device.CloseGate(ctx); err != nil {
		log.Printf("Warning: Could not close gate: %v", err)
	} else {
		fmt.Println("Gate closed")
	}

	// Obtener estado final
	fmt.Println("\nGetting final status...")
	finalStatus, err := device.GetStatus(ctx)
	if err != nil {
		log.Printf("Warning: Could not get final status: %v", err)
	} else {
		fmt.Printf("Final Machine Number: %d\n", finalStatus.MachineNumber)
	}

	fmt.Println("\nExample completed successfully!")
}
