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
	"encoding/json"
	"fmt"
	"os/exec"
)

type Meson struct {
	Provides string
	setup    bool
	info     mesonInfo
}

type mesonInfo struct {
	Version         string `json:"version"`
	DescriptiveName string `json:"descriptive_name"`
}

func (p *Project) generateMesonWrapper(path, tgzFile string) error {
	found, err := p.fs.Exists("meson.build")
	if err != nil {
		return err
	}

	if !found {
		return nil
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
		p.opts.Meson.Provides, p.opts.Meson.Provides)

	file := path + "/" + p.repoName + ".wrap"
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
	found, err := p.fs.Exists("meson.build")
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	p.opts.Log("Examining the meson project.")

	if !p.opts.Meson.setup {
		p.opts.Log("Setting up a meson project directory for examination. (%s)", mesonDirPath)
		setup := exec.Command(mesonCmd, "setup", mesonDirPath)
		if err := setup.Run(); err != nil {
			return err
		}
		p.opts.Meson.setup = true
	}

	p.opts.Log("Asking meson for project information.")
	cmd := exec.Command(mesonCmd, "introspect", mesonDirPath, "--projectinfo")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(out, &p.opts.Meson.info); err != nil {
		return err
	}

	p.opts.Log("Validating that the release version in the meson.build file matches the version in %s.", p.opts.ChangelogFile)

	v := p.opts.TagPrefix + p.opts.Meson.info.Version

	if p.nextRelease.Version != v {
		return fmt.Errorf("%w: %s declared '%s' but meson.build declared '%s'.  They must match.",
			errVersionMismatch, p.opts.ChangelogFile, p.nextRelease.Version, v)
	}

	return nil
}
