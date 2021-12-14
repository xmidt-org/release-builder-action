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
package project

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/afero"
	changelog "github.com/xmidt-org/gokeepachangelog"
	"github.com/xmidt-org/release-builder-action/git"
)

var (
	errRepoMissing        = errors.New("repository must be specified")
	errTokenMissing       = errors.New("token must be specified")
	errChangelogMissing   = errors.New("changelog must be specified")
	errArtifactDirMissing = errors.New("artifact dir must be specified")
	errSHAFileMissing     = errors.New("shasum-file must be specified")
	errRepoFormatError    = errors.New("the slug format is invalid")
	errPathNotDirectory   = errors.New("path is not a directory")
	//errVersionMismatch    = errors.New("the versions do not match")
)

type ProjectOpts struct {
	Slug          string
	BasePath      string
	Token         string
	TagPrefix     string
	ChangelogFile string
	ArtifactDir   string
	SHASumFile    string
	Now           time.Time
	Log           func(string, ...interface{})
	Meson         Meson
}

type GitIF interface {
	IsTagPresent(string) (bool, error)
	TagHead(string, string) error
	PushTags(string) error
	CreateArchive(string, string, string, string) (string, error)
}

type Project struct {
	opts        ProjectOpts
	dryRun      bool
	org         string
	repoName    string
	fs          *afero.Afero
	changelog   *changelog.Changelog
	nextRelease *changelog.Release
	git         GitIF
}

func NewProject(opts ProjectOpts, dryrun bool) (*Project, error) {
	if opts.Slug == "" {
		return nil, errRepoMissing
	}
	if opts.ChangelogFile == "" {
		return nil, errChangelogMissing
	}
	if opts.ArtifactDir == "" {
		return nil, errArtifactDirMissing
	}
	if opts.SHASumFile == "" {
		return nil, errSHAFileMissing
	}

	tmp := strings.Split(opts.Slug, "/")
	if len(tmp) != 2 {
		return nil, fmt.Errorf("%w: '%s' invalid", errRepoFormatError, opts.Slug)
	}

	if !dryrun && opts.Token == "" {
		return nil, errTokenMissing
	}

	p := Project{
		opts:     opts,
		dryRun:   dryrun,
		org:      tmp[0],
		repoName: tmp[1],
		fs: &afero.Afero{
			Fs: afero.NewOsFs(),
		},
	}
	if p.opts.Log == nil {
		p.opts.Log = func(s string, v ...interface{}) {}
	}

	// Open existing git repo
	g, err := git.Open(p.opts.BasePath)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to open the git path: '%s'", err, p.opts.BasePath)
	}
	p.git = g

	return &p, nil
}

// ExamineProject
func (p *Project) ExamineProject() error {
	p.opts.Log("Processing the %s file.", p.opts.ChangelogFile)
	if err := p.processChangelog(); err != nil {
		return err
	}

	p.opts.Log("Examining the git repo tags.")
	if err := p.examineTags(); err != nil {
		return err
	}

	if !p.FoundNewRelease() {
		p.opts.Log("No new release found.")
		return nil
	}

	return p.examineMesonProject()
}

func (p *Project) FoundNewRelease() bool {
	return p.nextRelease != nil
}

func (p *Project) Release() error {
	if !p.FoundNewRelease() {
		p.opts.Log("No new release found.")
		return nil
	}

	p.opts.Log("Prepairing the release: %s.", p.nextRelease.Version)

	p.opts.Log("Tagging the repository.")
	v := p.nextRelease.Version
	if err := p.git.TagHead(v, "Releasing: "+v); err != nil {
		return err
	}

	// Make the artifact dir if needed
	p.opts.Log("Ensuring the artifact directory is present.")
	artDir := p.opts.BasePath + "/" + p.opts.ArtifactDir
	if err := mkdir(p.fs, artDir); err != nil {
		return err
	}

	slug := p.getReleaseSlug()
	p.opts.Log("Creating the zip archive.")
	_, err := p.git.CreateArchive(slug, p.nextRelease.Version, "zip", artDir)
	if err != nil {
		return err
	}
	p.opts.Log("Creating the tar.gz archive.")
	tgz, err := p.git.CreateArchive(slug, p.nextRelease.Version, "tar.gz", artDir)
	if err != nil {
		return err
	}

	if err = p.generateMesonWrapper(artDir, tgz); err != nil {
		return err
	}

	p.opts.Log("Creating the sha256sum file.")
	if err = generateSha256Sum(p.fs, p.opts.SHASumFile, artDir); err != nil {
		return err
	}

	if p.dryRun {
		p.opts.Log("This is a dry run, not pushing the tags.")
		return nil
	}

	p.opts.Log("Pushing the tags to the upstream repository.")
	return p.git.PushTags(p.opts.Token)
}

func (p *Project) OutputData() {
	if p.FoundNewRelease() {
		fmt.Printf("::set-output name=release-name::%s %s\n", p.nextRelease.Version, p.opts.Now.Format("2006-01-02"))
		fmt.Printf("::set-output name=release-body::%s\n", strings.Join(p.nextRelease.Body[1:], "\n"))
		fmt.Printf("::set-output name=artifact-dir::%s\n", p.opts.ArtifactDir)
	}
}

func (p *Project) getReleaseSlug() string {
	return p.repoName + "-" + strings.TrimPrefix(p.nextRelease.Version, p.opts.TagPrefix)
}

func (p *Project) examineTags() error {
	// Map changelog and git releases
	for _, rel := range p.changelog.Releases {
		if "unreleased" == strings.ToLower(rel.Version) {
			continue
		}

		present, err := p.git.IsTagPresent(rel.Version)
		if err != nil {
			return fmt.Errorf("%w: unable to process git repo", err)
		}

		if !present {
			p.nextRelease = &rel
			return nil
		}
	}

	return nil
}

func (p *Project) processChangelog() error {
	path := p.opts.BasePath + "/" + p.opts.ChangelogFile
	f, err := p.fs.Open(path)
	if err != nil {
		return fmt.Errorf("%w: unable to open the changelog file found here: '%s'", err, path)
	}
	defer f.Close()

	p.changelog, err = changelog.Parse(f)
	if err != nil {
		return fmt.Errorf("%w: unable to parse the changelog file found here: '%s'", err, path)
	}

	return nil
}
