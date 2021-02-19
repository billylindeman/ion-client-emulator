package pkg

import (
	"math/rand"
	"time"

	log "github.com/pion/ion-log"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	rootCmd = &cobra.Command{
		Use:   "ion-client-emulator",
		Short: "connect nanoarch to ion-sfu",
		Long:  "ion-client-emulator connects to an ion-sfu JSONRPC endpoint and publishes a nanoarch emulated videogame ",
	}
)

// Execute main entrypoint into application
func Execute() error {
	rand.Seed(time.Now().UTC().UnixNano())

	fixByFile := []string{"asm_amd64.s", "proc.go", "icegatherer.go"}
	fixByFunc := []string{}
	log.Init("debug", fixByFile, fixByFunc)

	return rootCmd.Execute()
}
