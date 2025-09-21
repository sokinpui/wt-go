package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sokinpui/wt-go/internal/git"
)

// CreateWorktreeAndBranch handles creation and switching of Git worktrees.
// If a worktree for the given branch already exists, it prints the path to that worktree.
// This allows for easy switching, e.g., `cd $(wt <branch>)`.
// If no worktree exists, it creates a new one. If the branch doesn't exist,
// it creates the branch as well. After creation, it prints the new worktree's path.
func CreateWorktreeAndBranch(branchName string) {
	if branchName == "" {
		fmt.Fprintf(os.Stderr, "Error: Branch name cannot be empty.\n")
		return
	}

	existingPath, err := FindWorktreePathForBranch(branchName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for existing worktree for branch '%s': %v\n", branchName, err)
		return
	}

	isSwitching := false
	if existingPath != "" {
		wd, err := os.Getwd()
		if err != nil {
			// If we can't get the current directory, we can't compare.
			// To be safe, don't update the state.
			fmt.Fprintf(os.Stderr, "Warning: could not get current working directory: %v\n", err)
		} else {
			absWd, errWd := filepath.Abs(wd)
			absExistingPath, errExisting := filepath.Abs(existingPath)
			if errWd == nil && errExisting == nil && absWd != absExistingPath {
				isSwitching = true
			}
		}
	} else {
		// No existing worktree, so we are creating one, which is a switch.
		isSwitching = true
	}

	if isSwitching {
		if err := saveCurrentWorktreeState(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save current worktree state: %v\n", err)
		}
	}

	if existingPath != "" {
		fmt.Print(existingPath)
		return
	}

	repoRoot, err := git.Exec("rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository or cannot determine root: %v\n", err)
		return
	}
	repoRoot = strings.TrimSpace(repoRoot)

	parentDir := filepath.Dir(repoRoot)
	repoBaseName := filepath.Base(repoRoot)

	worktreeCollectionDir := filepath.Join(parentDir, repoBaseName+".wt")
	sanitizedBranchName := strings.ReplaceAll(branchName, "/", "_")
	newWorktreePath := filepath.Join(worktreeCollectionDir, sanitizedBranchName)

	_, err = git.Exec("rev-parse", "--verify", "--quiet", "refs/heads/"+branchName)
	branchExists := err == nil

	var gitArgs []string

	if branchExists {
		fmt.Fprintf(os.Stderr, "worktree create: %s\n", newWorktreePath)
		gitArgs = []string{"worktree", "add", newWorktreePath, branchName}
	} else {
		fmt.Fprintf(os.Stderr, "branch create: %s\n", branchName)
		fmt.Fprintf(os.Stderr, "worktree create: %s\n", newWorktreePath)
		gitArgs = []string{"worktree", "add", "-b", branchName, newWorktreePath}
	}

	output, err := git.Exec(gitArgs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating worktree for branch '%s': %v\n%s\n", branchName, err, output)
		return
	}
	// Print any informational output from the git command to stderr.
	if strings.TrimSpace(output) != "" {
		fmt.Fprint(os.Stderr, output)
	}
	fmt.Print(newWorktreePath)
}

// RemoveWorktreeAndBranch removes a Git worktree and deletes its associated branch.
func RemoveWorktreeAndBranch(branchName string, force bool) {
	if branchName == "" {
		fmt.Fprintf(os.Stderr, "Error: Branch name cannot be empty.\n")
		return
	}

	if branchName == "main" || branchName == "master" {
		fmt.Fprintf(os.Stderr, "Error: Deleting the '%s' branch is not allowed.\n", branchName)
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

	removeArgs := []string{"worktree", "remove"}
	if force {
		removeArgs = append(removeArgs, "--force")
	}
	removeArgs = append(removeArgs, worktreePath)

	output, err := git.Exec(removeArgs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing worktree '%s': %v\n%s\n", worktreePath, err, output)
		return
	}
	fmt.Fprintf(os.Stderr, "worktree remove: %s\n", worktreePath)
	if strings.TrimSpace(output) != "" {
		fmt.Fprint(os.Stderr, output)
	}

	deleteFlag := "-d"
	if force {
		deleteFlag = "-D"
	}

	output, err = git.Exec("branch", deleteFlag, branchName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting branch '%s': %v\n%s\n", branchName, err, output)

		// Branch deletion failed, attempt to restore worktree to leave the user in a consistent state.
		fmt.Fprintf(os.Stderr, "Attempting to restore worktree at '%s'...\n", worktreePath)
		recreateArgs := []string{"worktree", "add", worktreePath, branchName}
		recreateOutput, recreateErr := git.Exec(recreateArgs...)
		if recreateErr != nil {
			fmt.Fprintf(os.Stderr, "FATAL: Could not restore worktree for branch '%s'. Please check your repository state.\nError: %v\n%s\n", branchName, recreateErr, recreateOutput)
		} else {
			fmt.Fprintf(os.Stderr, "Worktree for branch '%s' restored successfully.\n", branchName)
			fmt.Fprint(os.Stderr, recreateOutput)
		}
		return
	}
	fmt.Fprintf(os.Stderr, "branch delete: %s\n", branchName)
	if strings.TrimSpace(output) != "" {
		fmt.Fprint(os.Stderr, output)
	}
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

	var orderedBranchNames []string
	seenBranches := make(map[string]bool)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "branch ") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 { // Ensure there's a branch ref part
				branchRef := strings.TrimPrefix(parts[1], "refs/heads/")
				if branchRef != "" {
					if !seenBranches[branchRef] {
						orderedBranchNames = append(orderedBranchNames, branchRef)
						seenBranches[branchRef] = true
					}
				}
			}
		}
	}

	if len(orderedBranchNames) == 0 {
		fmt.Fprintln(os.Stdout, "No Git worktrees found.")
		return
	}

	fmt.Fprintln(os.Stdout, "Git worktree branches:")
	for _, branch := range orderedBranchNames {
		fmt.Println(branch)
	}
}

// SwitchToPreviousWorktree returns the path of the previous worktree from the state file.
// It also saves the current working directory to the state file to allow toggling.
func SwitchToPreviousWorktree() (string, error) {
	stateFile, err := getStateFilePath()
	if err != nil {
		return "", fmt.Errorf("getting state file path: %w", err)
	}

	content, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no previous worktree state found")
		}
		return "", fmt.Errorf("reading state file: %w", err)
	}

	path := strings.TrimSpace(string(content))
	if path == "" {
		return "", fmt.Errorf("state file is empty")
	}

	// Before returning the path to switch to, we should save the current path.
	// This allows for toggling between two worktrees with `wt -`.
	if err := saveCurrentWorktreeState(); err != nil {
		// Not a fatal error for switching, but the user should know.
		fmt.Fprintf(os.Stderr, "Warning: could not save current worktree state: %v\n", err)
	}

	return path, nil
}

func getStateFilePath() (string, error) {
	gitCommonDir, err := git.Exec("rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository or could not determine common git directory: %w", err)
	}
	gitCommonDir = strings.TrimSpace(gitCommonDir)

	return filepath.Join(gitCommonDir, "wt.state"), nil
}

func saveCurrentWorktreeState() error {
	stateFile, err := getStateFilePath()
	if err != nil {
		return fmt.Errorf("could not get state file path: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	content, err := os.ReadFile(stateFile)
	if err == nil {
		if strings.TrimSpace(string(content)) == wd {
			return nil // Path is the same, no need to update.
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not read state file for comparison: %w", err)
	}

	return os.WriteFile(stateFile, []byte(wd), 0644)
}
