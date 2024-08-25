package main

import (
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// @title LangChainGo-Ollama-RAG
// @version 1.0
// @description Ollama RAG
// @contact.name kiosk
// @contact.url http://www.swagger.io/support
// @contact.email xxxxx
func main() {
	rand.Seed(time.Now().UnixNano())
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	command := NewGoRAGServiceCommand()
	cobra.CheckErr(command.Execute())
}
