
build: build-gpg-handler

build-gpg-handler:
	go generate ./...
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build \
			-a \
			-v \
			-trimpath='true' \
			-buildvcs='true' \
			-compiler='gc' \
			-o gpg-handler.exe \
			cmd/gpg/main.go

install: build-gpg-handler
	mv gpg-handler.exe /usr/local/bin/

listen-gpg: build
	socat UNIX-LISTEN:gpg.sock,fork EXEC:./gpg-handler.exe
