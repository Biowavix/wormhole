# Nombre del ejecutable
BINARY_NAME=wh
# Ruta de instalación local
INSTALL_PATH=$(HOME)/.local/bin

.PHONY: build symlink clean help

help: ## Muestra esta ayuda estilizada
	@echo "\033[1;35m          .-------."
	@echo "      ..-'         '-.."
	@echo "    .'                 '."
	@echo "   /    .-----------.    \\"
	@echo "  |    /             \\    |"
	@echo "  |   |   WORMHOLE    |   |"
	@echo "  |    \\             /    |"
	@echo "   \\    '-----------'    /"
	@echo "    '.                 .'"
	@echo "      ''-..         ..-''"
	@echo "           '-------'\033[0m"
	@echo "\033[1;36mWORMHOLE DEVELOPMENT TOOLS\033[0m \033[0;90m(Usa 'make <target>')\033[0m"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;33m➜ %-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila el proyecto y genera el ejecutable wh
	@echo "Compilando $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go
	@echo "¡Hecho! Ejecutable generado como ./$(BINARY_NAME)"

symlink: build ## Crea un enlace simbólico en ~/.local/bin para acceso global
	@echo "Creando enlace simbólico en $(INSTALL_PATH)/$(BINARY_NAME)..."
	@mkdir -p $(INSTALL_PATH)
	ln -sf $(CURDIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Enlace creado. Ahora puedes usar '$(BINARY_NAME)' desde cualquier ubicación."

clean: ## Borra el binario compilado
	@echo "Limpiando..."
	rm -f $(BINARY_NAME)
	@echo "Limpieza completada."
