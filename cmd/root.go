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
    pat     string
)

var rootCmd = &cobra.Command{
    Use: "ghmon",
    Short: "A tool to monitor GitHub organization and its members’ public repositories and notify on Google Chat",
    Run: func(cmd *cobra.Command, args []string) {
        if pat == "" {
            pat = os.Getenv("GITHUB_PAT")
        }
        if org == "" {
            cmd.Help()
            return
        }
        // log.Printf("Using PAT: %q", pat)
        err := run(org, webhook, pat)
        if err != nil {
            log.Fatalf("Error executing command: %v", err)
        }
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    log.SetFlags(0)
    rootCmd.Flags().StringVarP(&org, "org", "o", "", "GitHub organization name (required)")
    rootCmd.Flags().StringVarP(&webhook, "webhook", "w", "", "Google Chat webhook URL")
    rootCmd.Flags().StringVarP(&pat, "pat", "p", "", "GitHub personal access token (default: tries to read from GITHUB_PAT environment variable)")
}