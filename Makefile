.PHONY: build

build:
	rm -rf build
	mkdir build
	go version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./build/bakso_ayam ./cmd/server/main.go
	upx ./build/bakso_ayam --best --lzma
	cp Dockerfile ./build/Dockerfile
	cp deploy.sh ./build/deploy.sh

