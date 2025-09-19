package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sokinpui/wt.go/internal/git"
)

// CreateWorktreeAndBranch creates a new Git worktree and a new branch.
// The worktree will be created in a sibling directory to the current repository,
// named after the branch.
func CreateWorktreeAndBranch(branchName string) {
	if branchName == "" {
		fmt.Fprintf(os.Stderr, "Error: Branch name cannot be empty.\n")
		return
	}

	repoRoot, err := git.Exec("rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository or cannot determine root: %v\n", err)
		return
	}
	repoRoot = strings.TrimSpace(repoRoot)

	// Extract parent directory and repository base name
	parentDir := filepath.Dir(repoRoot)
	repoBaseName := filepath.Base(repoRoot)

	// Construct the new worktree path: ../<repo>.wt/<branch>
	worktreeCollectionDir := filepath.Join(parentDir, repoBaseName+".wt")
	newWorktreePath := filepath.Join(worktreeCollectionDir, branchName)

	fmt.Printf("Creating new worktree '%s' at '%s' and new branch '%s'...\n", branchName, newWorktreePath, branchName)

	output, err := git.Exec("worktree", "add", "-b", branchName, newWorktreePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating worktree and branch '%s': %v\n%s\n", branchName, err, output)
		return
	}

	fmt.Printf("Successfully created worktree and branch '%s'.\n", branchName)
	fmt.Print(output)
}

// RemoveWorktreeAndBranch removes a Git worktree and deletes its associated branch.
func RemoveWorktreeAndBranch(branchName string) {
	if branchName == "" {
		fmt.Fprintf(os.Stderr, "Error: Branch name cannot be empty.\n")
		return
	}

	worktreePath, err := FindWorktreePathForBranch(branchName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding worktree for branch '%s': %v\n", branchName, err)
		return
	}
	if worktreePath == "" {
		fmt.Fprintf(os.Stderr, "Error: No worktree found for branch '%s'.\n", branchName)
		return
	}

	fmt.Printf("Removing worktree at '%s' for branch '%s'...\n", worktreePath, branchName)
	output, err := git.Exec("worktree", "remove", worktreePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing worktree '%s': %v\n%s\n", worktreePath, err, output)
		return
	}
	fmt.Printf("Worktree '%s' removed successfully.\n", worktreePath)
	fmt.Print(output)

	fmt.Printf("Deleting branch '%s'...\n", branchName)
	output, err = git.Exec("branch", "-D", branchName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting branch '%s': %v\n%s\n", branchName, err, output)
		return
	}
	fmt.Printf("Branch '%s' deleted successfully.\n", branchName)
	fmt.Print(output)
}

// FindWorktreePathForBranch parses `git worktree list --porcelain` to find the path
// of the worktree associated with the given branch name.
func FindWorktreePathForBranch(branchName string) (string, error) {
	output, err := git.Exec("worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(output, "\n")
	var currentPath string
	var currentBranch string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentPath != "" && currentBranch == branchName {
				return currentPath, nil
			}
			currentPath = ""
			currentBranch = ""
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				branchRef := strings.TrimPrefix(parts[1], "refs/heads/")
				currentBranch = branchRef
			}
		}
	}

	if currentPath != "" && currentBranch == branchName {
		return currentPath, nil
	}

	return "", nil
}

// ListWorktrees lists all existing Git worktrees.
// It parses the output of `git worktree list --porcelain` to display only branch names.
func ListWorktrees() {
	output, err := git.Exec("worktree", "list", "--porcelain")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
		return
	}

	branchNames := make(map[string]struct{})
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "branch ") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				branchRef := strings.TrimPrefix(parts[1], "refs/heads/")
				if branchRef != "" {
					branchNames[branchRef] = struct{}{}
				}
			}
		}
	}

	if len(branchNames) == 0 {
		fmt.Println("No Git worktrees found.")
		return
	}

	fmt.Println("Git worktree branches:")
	for branch := range branchNames {
		fmt.Println(branch)
	}
}
