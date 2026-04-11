APP_NAME := tcl

.PHONY: build clean test vet

build:
	go build -o $(APP_NAME) ./cmd/tcl

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(APP_NAME) $(APP_NAME).exe
	rm -rf dist/
