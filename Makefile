.PHONY: build install uninstall clean

BINARY_NAME=aws-tui
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)

build:
	go build -o $(BINARY_NAME) .

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_PATH)"

uninstall:
	sudo rm -f $(INSTALL_PATH)
	@echo "Uninstalled $(BINARY_NAME) from $(INSTALL_PATH)"

clean:
	rm -f $(BINARY_NAME)
	@echo "Cleaned build artifacts"
