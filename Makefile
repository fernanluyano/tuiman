BINARY     := tuiman
BUILD_DIR  := build

.PHONY: run build test vet clean

run:
	go run .

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
