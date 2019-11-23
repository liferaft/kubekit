# KubeKit Manifest

This Go package is to manage the KubeKit Manifest and Version. It's used by the KubeKit and KubekitOS teams to define the dependencies (software, packages or applications) that should exists in the KubekitOS image or VM.

## Requirements

Either **Go** or **Docker** is required to validate the changes to the Manifest.

## Adding a New Relase to the Manifest

**DO NOT MODIFY THE `MANIFEST` FILE MANUALLY**

To modify the Manifest to add a new release, for example the release  `1.2.3`, execute the following instructions:

1. Create a new branch for the new release, example: `git checkout -b 'release-1.2.3'`
2. Execute `make release NR=<new release number>`
3. Edit the new `releaseX_Y_Z.go` with the changes to the manifest.
4. Verify all the new changes are correct, executing `make test` or `make test-in-docker` if Go is not installed.
5. Commit the changes to GitHub, example `git add . && git commit "release 1.2.3"`
6. Create the Pull Request from this branch and follow to the process to merge it to the master branch

In summary, these are the commands to execute to create a new release, in this example it's the release `1.2.3`:

    git checkout -b 'release-1.2.3'
    make release NR=1.2.3
    # Modify the new release:
    vim release1_2_3.go
    make test-in-docker
    git add .
    git commit "release 1.2.3"

If something went wrong after executing `make release NR=x.y.z` you can rollback to the previous version with `make rollback-release`. If for some reason this doesn't work, then dispose the new git branch and go back to the master branch.

**IMPORTANT**: There is no need to have a `MANIFEST` file for KubeKit but it is generated when you execute  `make test`  or `make test-in-docker`.
