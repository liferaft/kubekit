# Contributing Guide
Before contributing, refer to the Kraken team guidelines in https://github.com/kraken/goblets. Then, refer to the project specific contributing guide in this document.

## Scope Creep
Pull requests, or PR's, should be kept as small as possible to prevent merge conflicts from other members on the team. Scope creep should be mitigated by creating more stories of smaller size. When large PR's are unavoidable, such as during a large refactor, make sure to coordinate with other team members who may be touching the same code.

## Code Generation
The configurator project is mainly made up of yaml configuration files for configuring a Kubernetes cluster node. All of the configuration code is packaged into a single Go source code file so that the yaml files can be embedded into a Go application. Since all of the yaml files funnel into a single Go file, this can frequently cause merge conflicts when more than one contributor is trying to merge a PR into configurator. Since the generated file is **always a product of the yaml files**, we can safely regenerate the file and override a conflicting version as long as all other files with merge conflicts are resolved. In other words, you can always destroy and rebuild `code.go`.

To prevent merge conflicts when submitting a PR, make sure to rebase your changes onto master and resolve any conflicts prior to generating the code and merging. This can be done via the following commands:

```bash
# While in the project directory with your feature branch checked out...

# Fetch the latest changes in remote master
git fetch origin
# Rebase changes from current feature branch onto head of remote master
git rebase origin/master
```

When you run the rebase command, you may end up with a conflict in the `code.go` file:

```
$ git rebase origin/master
First, rewinding head to replay your work on top of it...
Applying: UKS-XXXX: my changes that might contain a conflict
Using index info to reconstruct a base tree...
M	code.go
Falling back to patching base and 3-way merge...
Auto-merging code.go
CONFLICT (content): Merge conflict in code.go
error: Failed to merge in the changes.
Patch failed at 0003 conflict against master
Use 'git am --show-current-patch' to see the failed patch

Resolve all conflicts manually, mark them as resolved with
"git add/rm <conflicted_files>", then run "git rebase --continue".
You can instead skip this commit: run "git rebase --skip".
To abort and get back to the state before "git rebase", run "git rebase --abort".
```

If your code.go file has rebase conflicts (as indicated by the error above), regenerate the file and continue rebasing:

```bash
# Generate the code.go file again
make generate
# Add file to rebase operation and continue addressing conflicts
git add code.go
git rebase --continue
# Push rebased branch to Github to retry PR
git push origin --force
```
