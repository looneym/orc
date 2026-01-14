# ORC Installation

## System-Level Installation

ORC installs to your `$GOPATH/bin` directory for system-wide access.

### Installation Location

```bash
$GOPATH/bin/orc  # Usually: ~/go/bin/orc
```

Make sure `$GOPATH/bin` is in your `$PATH`:

```bash
# Add to ~/.zshrc or ~/.bashrc if not already there
export PATH="$HOME/go/bin:$PATH"
```

### Manual Installation

To build and install the current version:

```bash
cd ~/src/orc
go build -o $GOPATH/bin/orc ./cmd/orc
```

### Automatic Installation (Git Hook)

ORC includes a `post-merge` git hook that automatically rebuilds and installs the binary when you merge to `master`.

**Hook Location**: `.git/hooks/post-merge.old`

**Trigger**: Automatically runs after `git merge` on `master` branch

**Behavior**:
- Detects if current branch is `master`
- Builds binary: `go build -o $GOPATH/bin/orc ./cmd/orc`
- Installs to system location
- Shows success/failure message

**Hook Chain**:
The existing `.git/hooks/post-merge` (beads) calls `post-merge.old` (ORC) first, then runs beads logic. This allows both hooks to coexist.

### Verification

Check that `orc` is accessible:

```bash
which orc
# Should show: ~/go/bin/orc (or your $GOPATH/bin/orc)

orc --help
# Should display ORC help

orc summary
# Should show mission/operation summary
```

### Troubleshooting

**`orc: command not found`**
- Ensure `$GOPATH/bin` is in your `$PATH`
- Run `go env GOPATH` to find your GOPATH
- Add `export PATH="$(go env GOPATH)/bin:$PATH"` to your shell RC file

**Old binary version after merge**
- The post-merge hook only runs on `master` branch
- Manually run: `go build -o $GOPATH/bin/orc ./cmd/orc`
- Check hook is executable: `ls -l .git/hooks/post-merge.old`

**Build fails in hook**
- Ensure you're in the ORC repository directory
- Ensure Go is properly installed: `go version`
- Check for syntax errors: `go build ./cmd/orc`

## Development Workflow

When working on ORC features:

1. **Make changes** to code
2. **Commit changes** to feature branch
3. **Merge to master**: `git checkout master && git merge feature-branch`
4. **Hook auto-installs** - Binary updates automatically
5. **Test**: `orc <your-command>`

No manual build step needed! The post-merge hook handles installation.

## Uninstallation

To remove ORC:

```bash
rm $GOPATH/bin/orc
```

To disable auto-install hook:

```bash
rm .git/hooks/post-merge.old
```
