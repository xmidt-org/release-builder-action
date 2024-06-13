// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Git encapsulates the difficult to test go-git code.
type Git struct {
	repo *git.Repository
}

// Open opens a path as a git repository or errors out.
func Open(path string) (*Git, error) {
	g := &Git{}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	g.repo = repo

	return g, nil
}

// IsTagPresent returns true if the specified tag is present, false otherwise.
func (g *Git) IsTagPresent(tag string) (bool, error) {
	_, err := g.repo.Tag(tag)
	if err == nil {
		return true, nil
	}

	if err == git.ErrTagNotFound {
		return false, nil
	}

	return false, fmt.Errorf("%w: unable to process git repo", err)
}

// TagHead adds the specified tag to the head of the repo.
func (g *Git) TagHead(tag, msg string) error {
	head, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("%w: repo.Head() error", err)
	}
	hash := head.Hash()
	commit, err := g.repo.CommitObject(hash)
	if err != nil {
		return fmt.Errorf("%w: repo.CommitObject() error", err)
	}
	_, err = g.repo.CreateTag(tag, hash, &git.CreateTagOptions{
		Tagger:  &commit.Committer,
		Message: msg,
	})
	if err != nil {
		return fmt.Errorf("%w: repo.CreateTag() error for tag '%s'", err, tag)
	}

	return nil
}

// PushTags pushes the tags to the upstream/remote repo.
func (g *Git) PushTags(token string) error {
	opts := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth: &http.BasicAuth{
			Username: "ignored",
			Password: token,
		},
	}

	if err := opts.Validate(); err != nil {
		return fmt.Errorf("%w: failed opts.PushOptions.Validate()", err)
	}

	if err := g.repo.Push(opts); err != nil {
		return fmt.Errorf("%w: failed repo.Push()", err)
	}

	return nil
}

// CreateArchive creates the archive file based on the naming conventions.
func (g *Git) CreateArchive(slug, version, format, path string) (string, error) {
	base := path + "/" + slug

	args := []string{
		"archive",
		"--format=" + format,
		"-o", base + "." + format,
		"--prefix=" + slug + "/",
		version,
	}

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s... unable to generate the %s archive", err, string(out), format)
	}

	return base + "." + format, nil
}
