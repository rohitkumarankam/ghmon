package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var listUsers bool
var listRepos bool
// var pat string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List users or repositories",
	Long:  "Use the --users or --repos flags to list users or repositories.",
	Run: func(cmd *cobra.Command, args []string) {
		if listUsers {
		fmt.Println("Listing users...")
		// list users
		if(cmd.Parent().Flag("org").Value.String() == ""){
			fmt.Println("Please specify --org")
		}
		pat = cmd.Parent().Flag("pat").Value.String();
		if(pat == ""){
			pat = os.Getenv("GITHUB_PAT")
			if pat == "" {
				fmt.Println("Please specify --pat or set GITHUB_PAT environment variable")
				os.Exit(1)
			}
		}

		listMembers(cmd.Parent().Flag("org").Value.String(), cmd.Parent().Flag("pat").Value.String())
		} else if listRepos {
			fmt.Println("Listing all repositories...")
			// Add your logic here
			if(cmd.Parent().Flag("org").Value.String() == ""){
				fmt.Println("Please specify --org")
			}
			pat = cmd.Parent().Flag("pat").Value.String();
			if(pat == ""){
				pat = os.Getenv("GITHUB_PAT")
				if pat == "" {
					fmt.Println("Please specify --pat or set GITHUB_PAT environment variable")
					os.Exit(1)
				}
			}
			listAllRepos(cmd.Parent().Flag("org").Value.String(), cmd.Parent().Flag("pat").Value.String())
			
		} else {
			// fmt.Println("Please specify --users or --repos")
			// print all parent command parameter values
			// fmt.Printf("org: %s\n", cmd.Parent().Flag("org").Value)
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&listUsers, "users", "u", false, "List all members of the organization")
	listCmd.Flags().BoolVarP(&listRepos, "repos", "r", false, "List all repositories in the organization and members' public repositories")
}
