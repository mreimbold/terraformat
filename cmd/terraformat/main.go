package main

import (
	"os"

	"github.com/mreimbold/terraformat/internal/cli"
)

func main() {
	os.Exit(cli.Run("terraformat"))
}
