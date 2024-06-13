// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-FileCopyrightText: 2021 Weston Schmidt
// SPDX-License-Identifier: Apache-2.0
package project

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	changelog "github.com/xmidt-org/gokeepachangelog"
)

const (
	badChangelog  = `Not a valid changelog.`
	goodChangelog = `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.2.3]
- example
`
)

func writeFile(fs *afero.Afero, name, contents string) error {
	f, err := fs.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.WriteString(contents)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}

	if n != len(contents) {
		return fmt.Errorf("file size is incorrect")
	}

	return nil
}

func TestNewProject(t *testing.T) {
	tests := []struct {
		description string
		opts        ProjectOpts
		dryrun      bool
		expectedErr error
	}{
		{
			description: "dryrun simple success",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			dryrun: true,
		},
		{
			description: "dryrun without token success",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				BasePath:      "..",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			dryrun: true,
		},
		{
			description: "missing slug",
			opts: ProjectOpts{
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: errRepoMissing,
		},
		{
			description: "missing path",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: git.ErrRepositoryNotExists,
		},
		{
			description: "missing changelog",
			opts: ProjectOpts{
				Slug:        "foo/bar",
				BasePath:    "..",
				Token:       "token",
				ArtifactDir: "artifacts",
				SHASumFile:  "sha256sum.txt",
			},
			expectedErr: errChangelogMissing,
		},
		{
			description: "artifact dir missing",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: errArtifactDirMissing,
		},
		{
			description: "sha sum missing",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
			},
			expectedErr: errSHAFileMissing,
		},
		{
			description: "slug is invalid - too many slashs",
			opts: ProjectOpts{
				Slug:          "foo/bar/goo",
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: errRepoFormatError,
		},
		{
			description: "slug is invalid - no split",
			opts: ProjectOpts{
				Slug:          "foo",
				BasePath:      "..",
				Token:         "token",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: errRepoFormatError,
		},
		{
			description: "check missing token",
			opts: ProjectOpts{
				Slug:          "foo/bar",
				BasePath:      "..",
				ChangelogFile: "CHANGELOG.md",
				ArtifactDir:   "artifacts",
				SHASumFile:    "sha256sum.txt",
			},
			expectedErr: errTokenMissing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			p, err := NewProject(tc.opts, tc.dryrun)
			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.NotNil(p)
				return
			}
			assert.True(errors.Is(err, tc.expectedErr),
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					err, tc.expectedErr),
			)
		})
	}
}

func TestProcessChangelog(t *testing.T) {
	tests := []struct {
		description   string
		changelogFile string
		expectedErr   error
	}{
		{
			description:   "success",
			changelogFile: "CHANGELOG.md",
		},
		{
			description:   "no changelog file with that name",
			changelogFile: "missing.md",
			expectedErr:   os.ErrNotExist,
		},
		{
			description:   "changelog file with invalid contents",
			changelogFile: "BAD.md",
			expectedErr:   changelog.ErrParsing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			p := &Project{
				opts: ProjectOpts{
					ChangelogFile: tc.changelogFile,
					BasePath:      ".",
				},
				fs: &afero.Afero{
					Fs: afero.NewMemMapFs(),
				},
			}

			err := writeFile(p.fs, "CHANGELOG.md", goodChangelog)
			require.NoError(t, err)
			err = writeFile(p.fs, "BAD.md", badChangelog)
			require.NoError(t, err)

			err = p.processChangelog()
			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.NotNil(p)
				return
			}
			assert.True(errors.Is(err, tc.expectedErr),
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					err, tc.expectedErr),
			)
		})
	}
}

func TestExamineTags(t *testing.T) {
	errTest := errors.New("test error")
	release := &changelog.Changelog{
		Releases: []changelog.Release{
			{
				Version: "unreleased",
			},
			{
				Version: "0.1.2",
			},
			{
				Version: "0.1.1",
			},
		},
	}
	norelease := &changelog.Changelog{
		Releases: []changelog.Release{
			{
				Version: "unreleased",
			},
			{
				Version: "0.1.1",
			},
		},
	}
	tests := []struct {
		description string
		cl          *changelog.Changelog
		release     bool
		expectedErr error
	}{
		{
			description: "success with a release",
			cl:          release,
			release:     true,
		},
		{
			description: "success with no release",
			cl:          norelease,
		},
		{
			description: "failure due to git call",
			cl:          release,
			expectedErr: errTest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			mockGit := &mockGit{}
			if tc.expectedErr != nil {
				mockGit.On("IsTagPresent", mock.Anything).Return(false, tc.expectedErr)
			} else {
				mockGit.On("IsTagPresent", "0.1.1").Return(true, nil)
				mockGit.On("IsTagPresent", "0.1.2").Return(false, nil)
			}

			p := &Project{
				changelog: tc.cl,
				git:       mockGit,
			}

			err := p.examineTags()
			if tc.expectedErr == nil {
				assert.NoError(err)
				if tc.release {
					assert.NotNil(p.nextRelease)
				} else {
					assert.Nil(p.nextRelease)
				}
				return
			}
			assert.True(errors.Is(err, tc.expectedErr),
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					err, tc.expectedErr),
			)
		})
	}
}

func TestGetReleaseSlug(t *testing.T) {
	assert := assert.New(t)

	p := &Project{
		repoName: "repo-name",
		changelog: &changelog.Changelog{
			Releases: []changelog.Release{
				{
					Version: "unreleased",
				},
				{
					Version: "0.1.2",
				},
				{
					Version: "0.1.1",
				},
			},
		},
	}
	p.nextRelease = &p.changelog.Releases[1]

	assert.Equal("repo-name-0.1.2", p.getReleaseSlug())
}
