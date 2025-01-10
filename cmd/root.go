package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
    org     string
    webhook string
)

var rootCmd = &cobra.Command{
    Use:   "myprogram",
    Short: "A tool to fetch updated GitHub repositories and notify a Google Chat webhook",
    // The `RunE` can be used if you want to return an error from this command
    Run: func(cmd *cobra.Command, args []string) {
        pat := os.Getenv("GITHUB_PAT")
        if pat == "" {
            log.Fatal("Please set the GITHUB_PAT environment variable")
        }

        if org == "" {
            log.Fatal("Please provide a GitHub organization name using --org (or -o)")
        }

        // if webhook == "" {
        //     log.Fatal("Please provide a Google Chat webhook URL using --webhook (or -w)")
        // }

        // Call the actual logic from another function (in run.go)
        err := run(org, webhook, pat)
        if err != nil {
            log.Fatalf("Error executing command: %v", err)
        }
    },
}

// Execute runs the root command which parses flags, etc.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.Flags().StringVarP(&org, "org", "o", "", "GitHub organization name (required)")
    rootCmd.Flags().StringVarP(&webhook, "webhook", "w", "", "Google Chat webhook URL (required)")
}