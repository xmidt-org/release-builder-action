# Release Builder Action
An action that builds release source artifacts and notes from the changelog file.

## Motivation

### What it does?

- Collect the snapshot of the repository as a tarball and zip file as artifacts.
- Optionally generates a [Meson](https://mesonbuild.com/) wrap file to associate with the release.
- Generates release notes based on the [changelog](https://keepachangelog.com/en/1.0.0/) file present and the tag.
- Generates sha256sum values for assets.
- Uploads the collection of source artifacts and sha256sum value with release notes as a release.

### Why do all this work?

1. It is really handy with the changelog based tag (TBD) action to be able
   to release artifacts automatically by simply updating the project CHANGELOG.md
   and merging the change to main.
2. If you are using the Meson build based project the automatic generation of
   the meson wrap file makes inclusion into other projects simple.
1. Github has a long standing bug that the yocto community has run into regarding
   the checksum values and source tarball/zip files not being 100% stable at all
   times.  A workaround to this is to build your own source artifacts, calculate
   the sha256sum and upload them yourself.

## Action Inputs

- **changelog**: An optional parameter that links to the changelog.md file.  Defaults to `CHANGELOG.md`.
- **artifact-dir**: An optional parameter that specifies the directory to place the artifacts.  Defaults to `artifacts`.
- **meson-wrap**: An optional parameter that indicates if a meson wrap file should be created for the project.  Defaults to `false`.
- **meson-provides**: An optional parameter that overwrites the dependency name output in the meson wrap file.  Defaults to `none` indicating to use the project name.  For example if the repository name is `foobar` but the library should be called `cat` set this parameter to `cat` to have `libcat_dep` specified in the wrap file.
- **verbose**: An optional parameter that indicates the debugging output that should be present in the logs.  Defaults to `0` for no debug. `1` enables all extra debug output.

## Action Outputs

- **release-name**: The release name based on the input.
- **release-body**: The release body based on the input.
- **artifact-dir**: The directory containing the artifacts.

## Example
This example will build the artifacts when a versioned tag is pushed:

```yml
name: Release

on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+*'

jobs:
  release:
    runs-on: ubuntu-latest
    name: Release Job
    steps:
      - uses: actions/checkout@v2
      - uses: xmidt-org/release-builder-action@v1
        name: Generate Release
        id: generate_release
      - name: Create Release
        id: create_release
        uses: ncipollo/release-action@v1
        with:
          name: ${{ steps.generate_release.outputs.release-name }}
          draft: false
          prerelease: false
          bodyFile: ${{ steps.generate_release.outputs.release-body }}
          artifacts: "${{ steps.generate_release.outputs.artifact-dir }}/*"
          token: ${{ secrets.GITHUB_TOKEN }}
```
