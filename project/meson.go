// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package project

import (
	"fmt"
)

type Meson struct {
	Provides string
}

func (p *Project) generateMesonWrapper(path, tgzFile string) error {

	found, err := p.fs.Exists("meson.build")
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	provides := p.opts.Meson.Provides
	if provides == "none" {
		provides = p.repoName
	}

	p.opts.Log("Generating the meson wrapper file.")
	slug := p.getReleaseSlug()

	sha, err := sha(p.fs, tgzFile)
	if err != nil {
		return err
	}

	line := fmt.Sprintf(
		"[wrap-file]\n"+
			"directory = %s\n\n"+
			"source_filename = %s.tar.gz\n"+
			"source_url = https://github.com/%s/releases/download/%s/%s.tar.gz\n"+
			"source_hash = %x\n\n"+
			"[meson_provides]\n"+
			"lib%s = lib%s_dep\n",
		slug,
		slug,
		p.opts.Slug, p.nextRelease.Version, slug,
		sha,
		provides, provides)

	file := path + "/" + provides + ".wrap"
	f, err := p.fs.Create(file)
	if err != nil {
		return fmt.Errorf("%w: unable to create file '%s'", err, file)
	}

	_, err = fmt.Fprintln(f, line)
	if err != nil {
		return fmt.Errorf("%w: unable to write to file '%s'", err, file)
	}
	f.Close()

	return nil
}

func (p *Project) examineMesonProject() error {
	// TODO: To do this right we'd examine the meson file directly and validate
	// the version number matches.  However, to do that with the tool as it stands
	// today requires installing all the dependencies.
	return nil
}
