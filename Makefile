GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

BINARY_NAME=ds205a-cli
BINARY_PATH=./cmd/ds205a-cli
BUILD_DIR=./bin

# Colores para output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: all build clean test coverage deps help

all: clean deps test build

## build: Compila el binario principal
build:
	@echo "$(GREEN)Compilando $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

## clean: Limpia archivos de build y cache
clean:
	@echo "$(YELLOW)Limpiando archivos...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## test: Ejecuta todas las pruebas
test:
	@echo "$(GREEN)Ejecutando pruebas...$(NC)"
	$(GOTEST) -v ./...

## test-coverage: Ejecuta pruebas con coverage
test-coverage:
	@echo "$(GREEN)Ejecutando pruebas con coverage...$(NC)"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generado en coverage.html$(NC)"

## deps: Instala dependencias
deps:
	@echo "$(GREEN)Instalando dependencias...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify

## tidy: Limpia dependencias no utilizadas
tidy:
	@echo "$(YELLOW)Limpiando dependencias...$(NC)"
	$(GOMOD) tidy

## fmt: Formatea el código
fmt:
	@echo "$(GREEN)Formateando código...$(NC)"
	$(GOCMD) fmt ./...

## vet: Ejecuta go vet
vet:
	@echo "$(GREEN)Ejecutando go vet...$(NC)"
	$(GOCMD) vet ./...

## lint: Ejecuta golangci-lint (requiere instalación previa)
lint:
	@echo "$(GREEN)Ejecutando linter...$(NC)"
	golangci-lint run

## install: Instala el binario en GOPATH/bin
install: build
	@echo "$(GREEN)Instalando $(BINARY_NAME)...$(NC)"
	$(GOCMD) install $(BINARY_PATH)

## run-example: Ejecuta un ejemplo
run-example:
	@echo "$(GREEN)Ejecutando ejemplo básico...$(NC)"
	$(GOCMD) run ./examples/basic/main.go

## help: Muestra esta ayuda
help:
	@echo "Comandos disponibles:"
	@sed -n 's/^##//p' Makefile