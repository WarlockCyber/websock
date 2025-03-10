package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	Port           int64
	PprofPort      int64
	ReadBuferSize  int64
	WriteBuferSize int64
	PingTimeout    int64
	Crt            string
	Key            string
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

	readBS := defReadSize
	writeBS := defWriteSize
	pingTimeout := defPingTimeout

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

	cfg = &config{
		Port:           port,
		PprofPort:      pprofPort,
		WriteBuferSize: int64(writeBS),
		ReadBuferSize:  int64(readBS),
		PingTimeout:    int64(pingTimeout),
		Crt:            os.Getenv(crtEnvName),
		Key:            os.Getenv(keyEnvName),
	}

	return nil
}
