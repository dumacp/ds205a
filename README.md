# DS205A Swing Turnstile Gate

Módulo Go para controlar torniquetes tipo puerta PMR DS205A a través de protocolo RS485.

## Descripción

Este paquete proporciona una interfaz para comunicarse con torniquetes DS205A utilizando el protocolo RS485. Permite controlar y monitorear el estado del torniquete de forma programática.

## Características

- ✅ Comunicación RS485
- ✅ Control de torniquete (abrir/cerrar)
- ✅ Monitoreo de estado
- ✅ Manejo de eventos
- ✅ CLI para pruebas y administración

## Estructura del Proyecto

```
├── pkg/
│   ├── ds205a/      # API pública principal
│   └── rs485/       # Comunicación RS485
├── internal/
│   └── protocol/    # Implementación del protocolo interno
├── cmd/
│   └── ds205a-cli/  # Herramienta de línea de comandos
├── examples/        # Ejemplos de uso
├── test/           # Pruebas de integración
└── doc/            # Documentación del dispositivo
```

## Instalación

```bash
go get github.com/dumacp/ds205a
```

## Uso Básico

```go
import "github.com/dumacp/ds205a/pkg/ds205a"

// Crear una nueva instancia del torniquete
turnstile, err := ds205a.New("/dev/ttyUSB0", 0x01, 9600, 5*time.Second)
if err != nil {
    log.Fatal(err)
}
defer turnstile.Close()

// Abrir el torniquete por la izquierda
err = turnstile.LeftOpen(context.Background(), 1)
if err != nil {
    log.Printf("Error al abrir torniquete: %v", err)
}
```

## CLI Tool

### Instalación

Para compilar e instalar la herramienta de línea de comandos:

```bash
go install github.com/dumacp/ds205a/cmd/ds205a-cli@latest
```

O compilar localmente:

```bash
git clone https://github.com/dumacp/ds205a.git
cd ds205a
go build -o ds205a-cli ./cmd/ds205a-cli
```

### Uso del CLI

```bash
# Obtener estado del torniquete
ds205a-cli -cmd status

# Abrir paso izquierdo con baudrate personalizado
ds205a-cli -port /dev/ttyUSB0 -baud 115200 -cmd left-open -value1 1

# Configuracion de parametros internos value1 = Menu , value2 = 2
ds205a-cli -port /dev/ttyUSB0 -baud 115200 -cmd set-param -value1 1 -value2 1

# Deshabilitar restricciones de paso
ds205a-cli -cmd disable-restrictions

# Ver todas las opciones y comandos disponibles
ds205a-cli --help
```

## Documentación

La documentación del dispositivo está disponible en el directorio [doc/](doc/).

## Contribución

Las contribuciones son bienvenidas. Por favor, asegúrate de ejecutar las pruebas antes de enviar un pull request.

## Licencia

[Especificar licencia]