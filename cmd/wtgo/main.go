package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sokinpui/wt-go/internal/worktree"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wtgo",
	Short: "wtgo is a CLI tool for managing Git worktrees",
	Long: `wtgo (worktree) is a command-line interface tool designed to simplify Git worktree management. It allows you to list, create, and remove Git worktrees with ease.

Usage:
  wtgo                       List all Git worktrees
  wtgo <branch>              Create a new worktree and branch named <branch>
  wtgo -                     Switch to the previous worktree
  wtgo --rm <branch>         Remove worktree <branch> and delete branch <branch> (use with caution)
  git branch | fzf | wtgo    Create a new worktree for a branch selected via fzf
`,
	Run: func(cmd *cobra.Command, args []string) {
		if removeFlag { // Guard clause for --rm flag
			if len(args) != 1 {
				fmt.Fprintf(os.Stderr, "Error: The --rm flag requires exactly one argument (the branch name).\n")
				os.Exit(1)
			}
			worktree.RemoveWorktreeAndBranch(args[0])
			return
		}

		// If arguments are provided, process them directly.
		if len(args) == 1 {
			if args[0] == "-" {
				path, err := worktree.SwitchToPreviousWorktree()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				fmt.Print(path)
				return
			}
			worktree.CreateWorktreeAndBranch(args[0])
			return
		}

		// If no arguments, check for stdin input.
		if len(args) == 0 {
			stat, err := os.Stdin.Stat()
			// Check if stdin is piped (not a character device like a terminal)
			if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					branchName := strings.TrimSpace(scanner.Text())
					if branchName != "" {
						worktree.CreateWorktreeAndBranch(branchName)
						return
					}
				}
				if err := scanner.Err(); err != nil {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					os.Exit(1)
				}
				// If stdin was piped but provided no valid branch name, fall through to list worktrees.
			}
			// No arguments and no valid stdin input, list worktrees.
			worktree.ListWorktrees()
			return
		}

		// If more than one argument is provided (and not --rm), it's an error.
		fmt.Fprintf(os.Stderr, "Error: Too many arguments. See 'wtgo --help'.\n")
		os.Exit(1)
	},
}

// removeFlag is a persistent flag to indicate removal of a worktree.
var removeFlag bool

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// main is the entry point for the wtgogo application.
func main() {
	Execute()
}

func init() {
	// Add persistent flags here
	rootCmd.PersistentFlags().BoolVarP(&removeFlag, "rm", "", false, "Remove a Git worktree and delete its branch")
}
