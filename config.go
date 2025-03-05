package main

import (
	"os"
	"strconv"
)

type config struct {
	Port           int64
	ReadBuferSize  int64
	WriteBuferSize int64
	Crt            string
	Key            string
}

var cfg *config

func initConfig() error {
	port, err := strconv.ParseInt(os.Getenv(portEnvName), 10, 64)
	if err != nil {
		return err
	}

	readBS := defReadSize
	writeBS := defWriteSize

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

	cfg = &config{
		Port:           port,
		WriteBuferSize: int64(writeBS),
		ReadBuferSize:  int64(readBS),
		Crt:            os.Getenv(crtEnvName),
		Key:            os.Getenv(keyEnvName),
	}

	return nil
}
