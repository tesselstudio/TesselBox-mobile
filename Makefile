.PHONY: all build desktop mobile-apk ios clean

BINARY_NAME=tesselbox

all: desktop

build: desktop

desktop:
go build -tags desktop -o $(BINARY_NAME) ./cmd/main.go

run: desktop
./$(BINARY_NAME)

mobile-apk:
gomobile build -target=android -androidapi=21 -o TesselBox.apk ./cmd

ios:
gomobile build -target=ios -o TesselBox.app ./cmd

clean:
rm -f $(BINARY_NAME) TesselBox.apk

test:
go test ./...
