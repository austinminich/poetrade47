APP_NAME=poetrade47
BUILD_DIR=build
APP_DIR=poetrade47.AppDir

.PHONY: all linux windows clean

# Running 'make' with no arguments builds both platforms
all: linux windows

# Build Universal Linux AppImage
linux:
	@echo "--- Building Linux AppImage ---"
	mkdir -p $(APP_DIR)/usr/bin
	go build -o $(APP_DIR)/usr/bin/$(APP_NAME) main.go
	ln -rsf $(APP_DIR)/usr/bin/$(APP_NAME) $(APP_DIR)/AppRun
	mkdir -p $(BUILD_DIR)/linux
	ARCH=x86_64 appimagetool $(APP_DIR) $(BUILD_DIR)/linux/$(APP_NAME)-x86_64.AppImage
	chmod +x $(BUILD_DIR)/linux/$(APP_NAME)-x86_64.AppImage

# Build Windows Executable
windows:
	@echo "--- Building Windows Executable ---"
	mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(APP_NAME).exe main.go

# Wipe out build files to start fresh
clean:
	@echo "--- Cleaning Workspace ---"
	rm -rf $(BUILD_DIR)
	rm -rf squashfs-root
	rm -f $(APP_DIR)/AppRun
	rm -f $(APP_DIR)/usr/bin/$(APP_NAME)
