package main

import (
	"fmt"
	"os"

	emu "github.com/billylindeman/ion-client-emulator/pkg"
	"github.com/faiface/mainthread"
)

func run() {
	if err := emu.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	mainthread.Run(run)
}
