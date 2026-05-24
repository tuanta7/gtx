# GTX

Helper tool for Git

```shell
go install github.com/tuanta7/gtx@latest
```

```txt
Git Extensions

Usage:
  gtx [flags]
  gtx [command]

Available Commands:
  auth        Authenticate with GitHub
  back        Undo the last commit (soft reset to HEAD~1)
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  profile     Show authentication and profile info
  prune       Reset a branch history to a single fresh commit
  shadow      Commit with a temporary profile for shadowing

Flags:
  -h, --help     help for gtx
  -t, --toggle   Help message for toggle

Use "gtx [command] --help" for more information about a command.

```

The primary purpose of this tool is to make these actions easier:

- [How to delete all commit history in github?](https://stackoverflow.com/questions/13716658/how-to-delete-all-commit-history-in-github)
- [Make the current commit the only (initial) commit in a Git repository?](https://stackoverflow.com/questions/9683279/make-the-current-commit-the-only-initial-commit-in-a-git-repository)
- [Override configured user for a single git commit](https://stackoverflow.com/questions/19840921/override-configured-user-for-a-single-git-commit)
- [How do I revert a Git repository to a previous commit?](https://stackoverflow.com/questions/4114095/how-do-i-revert-a-git-repository-to-a-previous-commit)