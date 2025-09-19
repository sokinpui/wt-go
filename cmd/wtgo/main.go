package main

import (
	"fmt"
	"os"

	"github.com/sokinpui/wt.go/internal/worktree"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "wt is a CLI tool for managing Git worktrees",
	Long: `wt (worktree) is a command-line interface tool designed to simplify Git worktree management. It allows you to list, create, and remove Git worktrees with ease.

Usage:
  wt                       List all Git worktrees
  wt <branch>              Create a new worktree and branch named <branch>
  wt --rm <branch>         Remove worktree <branch> and delete branch <branch>
`, // Updated usage for --rm flag
	Run: func(cmd *cobra.Command, args []string) {
		if removeFlag { // Guard clause for --rm flag
			if len(args) != 1 {
				fmt.Fprintf(os.Stderr, "Error: The --rm flag requires exactly one argument (the branch name to remove).\n")
				os.Exit(1)
			}
			worktree.RemoveWorktreeAndBranch(args[0])
			return
		}

		// Handle cases when --rm is not set
		switch len(args) {
		case 0:
			worktree.ListWorktrees()
		case 1:
			worktree.CreateWorktreeAndBranch(args[0])
		default:
			fmt.Fprintf(os.Stderr, "Error: Too many arguments. See 'wt --help'.\n")
			os.Exit(1)
		}
	},
}

// removeFlag is a persistent flag to indicate removal of a worktree.
var removeFlag bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// main is the entry point for the wtgo application.
func main() {
	Execute()
}

func init() {
	// Add persistent flags here
	rootCmd.PersistentFlags().BoolVarP(&removeFlag, "rm", "", false, "Remove a Git worktree and delete its branch")
}
