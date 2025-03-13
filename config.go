package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	Port           int64
	PprofPort      int64
	APIPort        int64
	ReadBuferSize  int64
	WriteBuferSize int64
	PingTimeout    int64
	Crt            string
	Key            string
	PProfEnabled   bool
}

var cfg *config

func initConfig() error {

	envFile, err := godotenv.Read(".env")
	if err == nil {
		for n, v := range envFile {
			os.Setenv(n, v)
		}
	}

	port, err := strconv.ParseInt(os.Getenv(portEnvName), 10, 64)
	if err != nil {
		return err
	}

	pprofPort, err := strconv.ParseInt(os.Getenv(pprofPortEnvName), 10, 64)
	if err != nil {
		return err
	}

	apiPort, err := strconv.ParseInt(os.Getenv(apiPortEnvName), 10, 64)
	if err != nil {
		return err
	}

	readBS := defReadSize
	writeBS := defWriteSize
	pingTimeout := defPingTimeout
	pprofEnabled := false

	if os.Getenv(readBuferSizeEnvName) != "" {
		r, err := strconv.ParseInt(os.Getenv(readBuferSizeEnvName), 10, 64)
		if err != nil {
			return err
		}

		readBS = int(r)
	}

	if os.Getenv(writeBuferSizeEnvName) != "" {
		r, err := strconv.ParseInt(os.Getenv(writeBuferSizeEnvName), 10, 64)
		if err != nil {
			return err
		}

		writeBS = int(r)
	}

	if os.Getenv(pingTimeoutEnvName) != "" {
		r, err := strconv.ParseInt(os.Getenv(pingTimeoutEnvName), 10, 64)
		if err != nil {
			return err
		}

		pingTimeout = int(r)
	}

	if os.Getenv(pprofEnabledEnvName) != "" {
		r, err := strconv.ParseInt(os.Getenv(pprofEnabledEnvName), 10, 64)
		if err != nil {
			return err
		}

		if r == 1 {
			pprofEnabled = true
		}
	}

	cfg = &config{
		Port:           port,
		PprofPort:      pprofPort,
		APIPort:        apiPort,
		WriteBuferSize: int64(writeBS),
		ReadBuferSize:  int64(readBS),
		PingTimeout:    int64(pingTimeout),
		Crt:            os.Getenv(crtEnvName),
		Key:            os.Getenv(keyEnvName),
		PProfEnabled:   pprofEnabled,
	}

	return nil
}
