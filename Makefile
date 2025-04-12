APP_NAME=chatapp

.PHONY: all build run clean lint

# Build the application
all: build

build:
	@echo "Tidying modules..."
	go mod tidy
	@echo "Building the application..."
	go build -o $(APP_NAME) main.go

# Run the application after building it
run: build
	@echo "Launching the application..."
	./$(APP_NAME)

# Clean the build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)
	
	# Lint the project
	lint:
		@echo "Linting the project with golangci-lint..."
		golangci-lint run ./...