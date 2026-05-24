# GTX - Git extensions.

```shell
go install github.com/tuanta7/gtx@latest
```

The primary purpose of this tool is to automate these actions:

- [How to delete all commit history in github?](https://stackoverflow.com/questions/13716658/how-to-delete-all-commit-history-in-github)
- [Make the current commit the only (initial) commit in a Git repository?](https://stackoverflow.com/questions/9683279/make-the-current-commit-the-only-initial-commit-in-a-git-repository)


## Usage

### Redo Last Commit

```shell
git reset --soft HEAD~1
git add .
git commit -m <new commit message>
git push -f 
```

### Replace Origin Remote

```shell
git remote 
```

