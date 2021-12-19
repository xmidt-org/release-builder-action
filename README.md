# Release Builder Action
An action that builds release source artifacts, generates a sha256sum for all artifacts,
and extracts release notes from the changelog file, placing them in the github release.

## Motivation

### What it does?

- Collect the snapshot of the repository as a tarball and zip file as artifacts.
- Generates release notes based on the [changelog](https://keepachangelog.com/en/1.0.0/) file present and the tag.
- Generates sha256sum values for all assets.
- Uploads the collection of source artifacts and sha256sum value with release notes as a release.
- Optionally generates a [Meson](https://mesonbuild.com/) wrap file to associate with the release.

### Why do this?

1. It is really handy to create a quality release based on simply updating the
   CHANGELOG.md file.
2. If you are using the Meson build based project the automatic generation of
   the meson wrap file makes inclusion into other projects simple.
3. Github has a long standing bug that the yocto community has run into regarding
   the checksum values and source tarball/zip files not being 100% stable at all
   times.  A workaround to this is to build your own source artifacts, calculate
   the sha256sum and upload them yourself.

## Action Inputs

- **gh-token**: (optional) Provides permissions for pushing the new tag.  Generally this should be `${{ secrets.GITHUB_TOKEN }}`.
- **changelog**: (optional) Name of the to the changelog.md file.  Defaults to `CHANGELOG.md`.
- **artifact-dir**: (optional) The directory to place the artifacts.  Defaults to `artifacts`.
- **tag-prefix**: (optional) The prefix for the tag used.  Defaults to `v`.
- **artifact-dir**: (optional) The name of the artifacts directory to work in and with.  Defaults to `artifacts`.
- **shasum-file**: (optional) The checksum file name to use.  Defaults to `sha256sum.txt`.
- **meson-provides**: (optional) The name of the meson artifact provided.  The name defaults to the repository name if not specified.
- **dry-run**: (optional) If `true` the tag is not pushed.  Defaults to `false`.

## Action Outputs

- **release-tag**: The release tag based on the input.
- **release-name**: The release name based on the input.
- **release-body-file**: The release body filename based on the input.
- **artifact-dir**: The directory containing the artifacts.

## Example
This example will build the artifacts when a versioned tag is pushed:

```yml
name: Release

on:
  push:
    paths:
      - "CHANGELOG.md"
    branches:
      - main

jobs:
  release:
    runs-on: [ ubuntu-latest ]
    name: Release Job
    steps:
      - uses: actions/checkout@v2

      - name: Generate Release Bundle
        uses: xmidt-org/release-builder-action@v3
        id: bundle
        with:
          gh-token: ${{ secrets.TOKEN }}

      - name: Upload Release
        uses: ncipollo/release-action@v1
        with:
          name: ${{ steps.bundle.outputs.release-name }}
          tag: ${{ steps.bundle.outputs.release-tag }}
          draft: false
          prerelease: false
          bodyFile: ${{ steps.bundle.outputs.release-body-file }}
          artifacts: "${{ steps.bundle.outputs.artifact-dir }}/*"
          token: ${{ secrets.TOKEN }}
```

**Note:** In the example we show using [ncipollo/release-action](https://github.com/ncipollo/release-action).  These work well together.
