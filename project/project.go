// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package project

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gh "github.com/sethvargo/go-githubactions"
	"github.com/spf13/afero"
	changelog "github.com/xmidt-org/gokeepachangelog"
	"github.com/xmidt-org/release-builder-action/git"
)

const (
	releaseBodyFile = ".release-body.md"
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

	if p.dryRun {
		p.opts.Log("This is a dry run, do not alter the repo or create artifacts.")
		return nil
	}

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

	p.opts.Log("Pushing the tags to the upstream repository.")
	return p.git.PushTags(p.opts.Token)
}

func (p *Project) OutputData() error {
	if p.FoundNewRelease() {
		v := p.nextRelease.Version
		now := time.Now().Format("2006-01-02")

		if !p.dryRun {
			f, err := p.fs.Create(releaseBodyFile)
			if err != nil {
				return fmt.Errorf("%w: unable to create file '%s'", err, releaseBodyFile)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil {
					p.opts.Log("unable to close file '%s': %v", releaseBodyFile, cerr)
					return
				}
			}()
			for _, line := range p.nextRelease.Body[1:] {
				_, err = fmt.Fprintln(f, line)
				if err != nil {
					return fmt.Errorf("%w: unable to write to file '%s'", err, releaseBodyFile)
				}
			}
		}

		gh.SetOutput("release-tag", v)
		gh.SetOutput("release-name", v+" "+now)
		gh.SetOutput("release-body-file", releaseBodyFile)
		gh.SetOutput("artifact-dir", p.opts.ArtifactDir)
	}
	return nil
}

func (p *Project) getReleaseSlug() string {
	return p.repoName + "-" + strings.TrimPrefix(p.nextRelease.Version, p.opts.TagPrefix)
}

func (p *Project) examineTags() error {
	// Map changelog and git releases
	for _, rel := range p.changelog.Releases {
		if strings.ToLower(rel.Version) == "unreleased" {
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
		break
	}

	return nil
}

func (p *Project) processChangelog() error {
	path := p.opts.BasePath + "/" + p.opts.ChangelogFile
	f, err := p.fs.Open(path)
	if err != nil {
		return fmt.Errorf("%w: unable to open the changelog file found here: '%s'", err, path)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			p.opts.Log("unable to close file '%s': %v", releaseBodyFile, cerr)
			return
		}
	}()

	p.changelog, err = changelog.Parse(f)
	if err != nil {
		return fmt.Errorf("%w: unable to parse the changelog file found here: '%s'", err, path)
	}

	return nil
}
