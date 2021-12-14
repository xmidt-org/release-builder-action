/**
 *  Copyright (c) 2021  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth: &http.BasicAuth{
			Username: "ignored",
			Password: token,
		},
	})

	if err != nil {
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

	_, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("%w: unable to generate the %s archive", err, format)
	}

	return base + "." + format, nil
}
