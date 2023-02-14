package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/ninjasanonymous/wsl2gpggo/internal/gpgHandler"
)

var (
	gpgBasePath   = flag.String("gpgBasePath", "", `gpg config path on windows. Defaults to C:\Users\<USERNAME>\AppData\Local\gnupg`)
	gpgSocketName = flag.String("gpgSocketName", "S.gpg-agent", "gpg socket name on windows. Defaults to S.gpg-agent")
)

func main() {
	flag.Parse()

	var basePath string
	if *gpgBasePath == "" {
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to find user home directory: %v", err)
		}

		basePath = filepath.Join(homeDirectory, "AppData", "Local", "gnupg")
	}

	gpgSocketPath := filepath.Join(basePath, *gpgSocketName)

	handler, err := gpgHandler.NewGPGHandler(gpgSocketPath, nil)
	if err != nil {
		log.Fatalf("Error when attempting to NewGPGHandler(): %v", err)
	}

	err = handler.Handle()
	if err != nil {
		log.Fatalf("Error when attempting to interact with GPG Socket: %v", err)
	}
}
