# wt-go: Git Worktree Manager

## Overview

`wt-go` is a command-line tool for managing Git worktrees. It is accompanied by `wt`, a Zsh wrapper script that provides enhanced interactive functionality using `fzf`.

The tool simplifies the process of creating, listing, removing, and switching between Git worktrees.

## Features

- **List**: Display all worktree branches.
- **Create/Switch**: Create a new worktree for a new or existing branch. If the worktree already exists, output its path for quick navigation.
- **Remove**: Delete a worktree and its associated branch.
- **Previous**: Switch to the last-used worktree.
- **Interactive Mode**: The `wt` wrapper uses `fzf` to provide an interactive menu for switching between worktrees.

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) (1.21 or later)
- [Git](https://git-scm.com/)
- [fzf](https://github.com/junegunn/fzf) (required for the `wt` wrapper)

### 1. Install `wtgo`

Install the `wtgo` binary from source:

```sh
go install github.com/sokinpui/wt-go/cmd/wtgo@latest
