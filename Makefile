APP_NAME=poetrade47
BUILD_DIR=build
APP_DIR=poetrade47.AppDir
APP_PACKAGE=./src


.PHONY: all linux windows clean

# Running 'make' with no arguments builds both platforms
all: linux windows

# Build Universal Linux AppImage
linux:
	@echo "--- Building Linux AppImage ---"
	mkdir -p $(APP_DIR)/usr/bin
	cp src/resources/cheesecake47.png $(APP_DIR)/cheesecake47.png
	go build -o $(APP_DIR)/usr/bin/$(APP_NAME) $(APP_PACKAGE)
	ln -rsf $(APP_DIR)/usr/bin/$(APP_NAME) $(APP_DIR)/AppRun
	mkdir -p $(BUILD_DIR)/linux
	ARCH=x86_64 appimagetool $(APP_DIR) $(BUILD_DIR)/linux/$(APP_NAME).AppImage
	chmod +x $(BUILD_DIR)/linux/$(APP_NAME).AppImage

# Build Windows Executable
windows:
	@echo "--- Building Windows Executable ---"
	mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(APP_NAME).exe $(APP_PACKAGE)

# Wipe out build files to start fresh
clean:
	@echo "--- Cleaning Workspace ---"
	rm -rf $(BUILD_DIR)
	rm -rf squashfs-root
	rm -f $(APP_DIR)/AppRun
	rm -f $(APP_DIR)/usr/bin/$(APP_NAME)
	rm -f $(APP_DIR)/cheesecake47.png