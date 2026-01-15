package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dumacp/ds205a/pkg/ds205a"
)

// Comandos disponibles
type Command string

const (
	CmdStatus              Command = "status"
	CmdInfo                Command = "info"
	CmdLeftOpen            Command = "left-open"
	CmdLeftAlwaysOpen      Command = "left-always-open"
	CmdRightOpen           Command = "right-open"
	CmdRightAlwaysOpen     Command = "right-always-open"
	CmdCloseGate           Command = "close-gate"
	CmdForbidLeft          Command = "forbid-left"
	CmdForbidRight         Command = "forbid-right"
	CmdDisableRestrictions Command = "disable-restrictions"
	CmdResetLeftCounters   Command = "reset-left-counters"
	CmdResetRightCounters  Command = "reset-right-counters"
	CmdSetParams           Command = "set-params"
	CmdReset               Command = "reset"
)

func main() {
	var (
		port     = flag.String("port", "/dev/ttyUSB0", "Serial port")
		baudRate = flag.Int("baud", 9600, "Baud rate (9600, 19200, 38400, 57600, 115200)")
		deviceID = flag.Int("id", 1, "Device ID")
		timeout  = flag.Duration("timeout", 5*time.Second, "Operation timeout")
		command  = flag.String("cmd", "", "Command to execute (see available commands below)")
		value    = flag.Int("value", 1, "Value parameter for commands that require it")
	)

	// Personalizar la salida de ayuda
	flag.Usage = func() {
		fmt.Printf("DS205A Turnstile CLI Tool\n")
		fmt.Printf("========================\n\n")
		fmt.Printf("Usage: %s [options] -cmd <command>\n\n", os.Args[0])
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		fmt.Printf("\nAvailable Commands:\n")
		printCommandsHelp()
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  %s -cmd %s\n", os.Args[0], CmdStatus)
		fmt.Printf("  %s -port /dev/ttyUSB1 -baud 115200 -cmd %s\n", os.Args[0], CmdInfo)
		fmt.Printf("  %s -cmd %s -value 1\n", os.Args[0], CmdLeftOpen)
		fmt.Printf("  %s -cmd %s\n", os.Args[0], CmdDisableRestrictions)
		fmt.Printf("  %s -cmd %s\n\n", os.Args[0], CmdCloseGate)
	}

	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// Validar comando antes de crear dispositivo
	validCmd := Command(*command)
	if !isValidCommand(validCmd) {
		fmt.Printf("Error: Invalid command '%s'\n\n", *command)
		fmt.Printf("Available commands: %s\n\n", getAvailableCommands())
		printUsage()
		os.Exit(1)
	}

	// Crear dispositivo
	device, err := ds205a.New(*port, byte(*deviceID), *baudRate, *timeout)
	if err != nil {
		log.Fatalf("Error creating device: %v", err)
	}

	// Abrir conexión
	if err := device.Open(); err != nil {
		log.Fatalf("Error opening device: %v", err)
	}
	defer device.Close()

	ctx := context.Background()

	// Ejecutar comando
	err = executeCommand(device, Command(*command), *value, ctx)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

func executeCommand(device *ds205a.Turnstile, cmd Command, value int, ctx context.Context) error {
	switch cmd {
	case CmdStatus:
		return cmdStatus(device, ctx)
	case CmdInfo:
		return cmdInfo(device, ctx)
	case CmdLeftOpen:
		return cmdLeftOpen(device, uint8(value), ctx)
	case CmdLeftAlwaysOpen:
		return cmdLeftAlwaysOpen(device, ctx)
	case CmdRightOpen:
		return cmdRightOpen(device, uint8(value), ctx)
	case CmdRightAlwaysOpen:
		return cmdRightAlwaysOpen(device, ctx)
	case CmdCloseGate:
		return cmdCloseGate(device, ctx)
	case CmdForbidLeft:
		return cmdForbiddenLeft(device, ctx)
	case CmdForbidRight:
		return cmdForbiddenRight(device, ctx)
	case CmdDisableRestrictions:
		return cmdDisableRestrictions(device, ctx)
	case CmdResetLeftCounters:
		return cmdResetLeftCounters(device, ctx)
	case CmdResetRightCounters:
		return cmdResetRightCounters(device, ctx)
	case CmdSetParams:
		return cmdSetParameters(device, uint8(value), ctx)
	case CmdReset:
		return cmdReset(device, ctx)
	default:
		return fmt.Errorf("unknown command: %s\nUse one of: %s", cmd, getAvailableCommands())
	}
}

func cmdStatus(device *ds205a.Turnstile, ctx context.Context) error {
	status, err := device.GetStatus(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Turnstile Status:\n")
	fmt.Printf("  Machine Number: %d\n", status.MachineNumber)
	fmt.Printf("  Direction: %s\n", status.Direction)
	fmt.Printf("  Position: %d\n", status.Position)
	fmt.Printf("  Memory: %d\n", status.Memory)
	fmt.Printf("  System Voltage: %d\n", status.SystemVoltage)
	fmt.Printf("  Temperature: %d\n", status.SystemTemperature)
	return nil
}

func cmdInfo(device *ds205a.Turnstile, ctx context.Context) error {
	info, err := device.GetDeviceInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Device Information:\n")
	fmt.Printf("  Version: %d.%d.%d\n", info.Version[0], info.Version[1], info.Version[2])
	fmt.Printf("  Machine Type: %d\n", info.MachineType)
	return nil
}

func cmdLeftOpen(device *ds205a.Turnstile, value uint8, ctx context.Context) error {
	fmt.Printf("Opening left passage with value %d...\n", value)
	return device.LeftOpen(ctx, value)
}

func cmdLeftAlwaysOpen(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Setting left passage to always open...")
	return device.LeftAlwaysOpen(ctx)
}

func cmdRightOpen(device *ds205a.Turnstile, value uint8, ctx context.Context) error {
	fmt.Printf("Opening right passage with value %d...\n", value)
	return device.RightOpen(ctx, value)
}

func cmdRightAlwaysOpen(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Setting right passage to always open...")
	return device.RightAlwaysOpen(ctx)
}

func cmdCloseGate(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Closing gate...")
	return device.CloseGate(ctx)
}

func cmdForbiddenLeft(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Forbidding left passage...")
	return device.ForbiddenLeftPassage(ctx)
}

func cmdForbiddenRight(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Forbidding right passage...")
	return device.ForbiddenRightPassage(ctx)
}

func cmdDisableRestrictions(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Disabling passage restrictions...")
	return device.DisablePassageRestrictions(ctx)
}

func cmdResetLeftCounters(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Resetting left counters...")
	return device.ResetLeftCounters(ctx)
}

func cmdResetRightCounters(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Resetting right counters...")
	return device.ResetRightCounters(ctx)
}

func cmdSetParameters(device *ds205a.Turnstile, value uint8, ctx context.Context) error {
	fmt.Printf("Setting parameters with value %d...\n", value)
	return device.SetParameters(ctx, value)
}

func cmdReset(device *ds205a.Turnstile, ctx context.Context) error {
	fmt.Println("Resetting device...")
	return device.Reset(ctx)
}

// getAvailableCommands retorna la lista de comandos disponibles
func getAvailableCommands() string {
	commands := []Command{
		CmdStatus, CmdInfo, CmdLeftOpen, CmdLeftAlwaysOpen,
		CmdRightOpen, CmdRightAlwaysOpen, CmdCloseGate,
		CmdForbidLeft, CmdForbidRight, CmdDisableRestrictions,
		CmdResetLeftCounters, CmdResetRightCounters,
		CmdSetParams, CmdReset,
	}

	var cmdStrs []string
	for _, cmd := range commands {
		cmdStrs = append(cmdStrs, string(cmd))
	}
	return strings.Join(cmdStrs, ", ")
}

// isValidCommand verifica si un comando es válido
func isValidCommand(cmd Command) bool {
	validCommands := []Command{
		CmdStatus, CmdInfo, CmdLeftOpen, CmdLeftAlwaysOpen,
		CmdRightOpen, CmdRightAlwaysOpen, CmdCloseGate,
		CmdForbidLeft, CmdForbidRight, CmdDisableRestrictions,
		CmdResetLeftCounters, CmdResetRightCounters,
		CmdSetParams, CmdReset,
	}

	for _, validCmd := range validCommands {
		if cmd == validCmd {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Println()
	fmt.Println("DS205A Turnstile CLI Tool")
	fmt.Println("========================")
	fmt.Println()
	fmt.Printf("  %s [options] -cmd <command>\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Available Commands:")
	printCommandsHelp()
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Printf("  %s -cmd %s\n", os.Args[0], CmdStatus)
	fmt.Printf("  %s -port /dev/ttyUSB1 -baud 115200 -cmd %s\n", os.Args[0], CmdInfo)
	fmt.Printf("  %s -cmd %s -value 1\n", os.Args[0], CmdLeftOpen)
	fmt.Printf("  %s -cmd %s\n", os.Args[0], CmdDisableRestrictions)
	fmt.Printf("  %s -cmd %s\n", os.Args[0], CmdCloseGate)
	fmt.Println()
}

// printCommandsHelp imprime solo la sección de comandos organizados por categoría
func printCommandsHelp() {
	// Comandos organizados por categoría
	commands := map[string][]struct {
		cmd        Command
		desc       string
		needsValue bool
	}{
		"Status & Info": {
			{CmdStatus, "Get turnstile status", false},
			{CmdInfo, "Get device information", false},
		},
		"Passage Control": {
			{CmdLeftOpen, "Open left passage", true},
			{CmdLeftAlwaysOpen, "Set left passage to always open", false},
			{CmdRightOpen, "Open right passage", true},
			{CmdRightAlwaysOpen, "Set right passage to always open", false},
			{CmdCloseGate, "Close the gate/turnstile", false},
		},
		"Restrictions": {
			{CmdForbidLeft, "Forbid left passage", false},
			{CmdForbidRight, "Forbid right passage", false},
			{CmdDisableRestrictions, "Disable all passage restrictions", false},
		},
		"Counters": {
			{CmdResetLeftCounters, "Reset left side counters", false},
			{CmdResetRightCounters, "Reset right side counters", false},
		},
		"Configuration": {
			{CmdSetParams, "Set device parameters", true},
			{CmdReset, "Reset device", false},
		},
	}

	for category, cmds := range commands {
		fmt.Printf("  %s:\n", category)
		for _, cmdInfo := range cmds {
			if cmdInfo.needsValue {
				fmt.Printf("    %-20s - %s (use -value <num>)\n", cmdInfo.cmd, cmdInfo.desc)
			} else {
				fmt.Printf("    %-20s - %s\n", cmdInfo.cmd, cmdInfo.desc)
			}
		}
		fmt.Println()
	}
}
