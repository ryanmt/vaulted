package main

import (
	"errors"
	"fmt"
	"os"
	// "runtime/pprof"
	// "runtime/trace"

	"github.com/miquella/vaulted/lib"
	"github.com/miquella/vaulted/lib/legacy"
)

const (
	EX_USAGE_ERROR     = 64
	EX_DATA_ERROR      = 65
	EX_UNAVAILABLE     = 69
	EX_TEMPORARY_ERROR = 79
)

type ErrorWithExitCode struct {
	error
	ExitCode int
}

var (
	ErrNoError           = errors.New("")
	ErrFileNotExist      = ErrorWithExitCode{os.ErrNotExist, EX_USAGE_ERROR}
	ErrNoPasswordEntered = ErrorWithExitCode{errors.New("Could not get password"), EX_UNAVAILABLE}

	vaultedErrMap = map[error]ErrorWithExitCode{
		vaulted.ErrIncorrectPassword:       ErrorWithExitCode{vaulted.ErrIncorrectPassword, EX_TEMPORARY_ERROR},
		vaulted.ErrInvalidKeyConfig:        ErrorWithExitCode{vaulted.ErrInvalidKeyConfig, EX_DATA_ERROR},
		vaulted.ErrInvalidEncryptionConfig: ErrorWithExitCode{vaulted.ErrInvalidEncryptionConfig, EX_DATA_ERROR},
	}
)

func main() {
	// Make this configurable
	// f1, err := os.Create("trace.out")
	// if err != nil {
	// panic(err)
	// }
	// defer f1.Close()
	// err = trace.Start(f1)
	// if err != nil {
	// panic(err)
	// }
	// defer trace.Stop()
	// f2, err := os.Create("pprof.out")
	// if err != nil {
	// panic(err)
	// }
	// defer f2.Close()
	// pprof.StartCPUProfile(f2)
	// defer pprof.StopCPUProfile()

	command, err := ParseArgs(os.Args[1:])
	if err == nil {
		steward := NewSteward()
		store := struct {
			vaulted.Store
			legacy.LegacyStore
		}{
			Store:       vaulted.New(steward),
			LegacyStore: legacy.New(steward),
		}

		err = command.Run(store)
	}

	if err != nil {
		if _, exists := vaultedErrMap[err]; exists {
			err = vaultedErrMap[err]
		}

		exiterr, ok := err.(ErrorWithExitCode)
		if !ok || exiterr.error != ErrNoError {
			fmt.Fprintln(os.Stderr, err)
		}
		if ok {
			os.Exit(exiterr.ExitCode)
		} else {
			os.Exit(1)
		}
	}
}
