.PHONY: run dev build help

run:
	@bash run.sh

dev:
	@bash run.sh

build:
	@echo "Compilando Ana sin PortAudio (requiere GCC para habilitarlo)..."
	go build -o ana.exe .\cmd\ana\main.go

help:
	@echo "Comandos disponibles:"
	@echo "  make run   - Ejecutar Ana (carga .env automáticamente)"
	@echo "  make dev   - Lo mismo que 'make run'"
	@echo "  make build - Compilar Ana"
	@echo "  make help  - Mostrar esta ayuda"
