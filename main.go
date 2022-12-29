package main

import (
	"log"

	"github.com/github/gh-projects/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "projects",
	}
	rootCmd.AddCommand(cmd.NewListCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
