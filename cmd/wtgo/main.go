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
	Long: `wt (worktree) is a command-line interface tool designed to simplify Git worktree management.
It allows you to list, create, and remove Git worktrees with ease.

Usage:
  wt                       List all Git worktrees
  wt <branch>              Create a new worktree and branch named <branch>
  wt rm <branch>           Remove worktree <branch> and delete branch <branch>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			worktree.ListWorktrees()
			return
		}

		worktree.CreateWorktreeAndBranch(args[0])
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm <branch>",
	Short: "Remove a Git worktree and delete its branch",
	Long:  `Removes the specified Git worktree and deletes the corresponding branch.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		worktree.RemoveWorktreeAndBranch(args[0])
	},
}

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
	rootCmd.AddCommand(rmCmd)
}
