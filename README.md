# CIer

Tool for collecting and rolling out GitHub Actions workflows across projects.

## Idea

- `scan` searches repositories for `.github/workflows/*.yml|*.yaml` and adds new files to the SQLite database.
- Each workflow belongs to a user-defined group.
- `roll` lets you pick workflows by group and opens them in `nvim` with a prefix comment so you can save with a new file name.
- During `scan`, you can open a workflow in read-only mode before deciding how to categorize it.

## Installation

```bash
# in the project root
GOCACHE=/tmp/go-cache go build ./cmd/cier
```

## Usage

```bash
# scan a single repository
./cier scan /path/to/repo

# scan multiple repositories
./cier scan /path/to/repo1 /path/to/repo2

# roll selected workflows into the current project
./cier roll

# remove workflows from a group and add to the blacklist
./cier remove

# move workflows to another group
./cier move

# manage the blacklist
./cier blacklist list
./cier blacklist add /path/to/workflow.yml
./cier blacklist restore
```

## Database

By default, the database is created at `~/.config/cier/cier.db`.
You can provide an explicit path:

```bash
./cier --db /path/to/cier.db scan /path/to/repo
```

Or via environment variable:

```bash
CIER_DB=/path/to/cier.db ./cier scan /path/to/repo
```

## Notes

- `roll` creates `.github/workflows` in the current directory if it does not exist.
- For each selected workflow, a new unnamed buffer is opened in `nvim`.
- A comment is inserted at the top with the group, source, and project.
