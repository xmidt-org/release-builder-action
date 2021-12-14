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
package main

import (
	"errors"
	"fmt"

	flag "github.com/spf13/pflag"
	"github.com/xmidt-org/release-builder-action/project"
)

var (
	errBoolFormatError = errors.New("the string should be 'true' or 'false'")
)

func main() {
	p, err := parseAndValidateInput()
	if err != nil {
		Err("%s", err)
		return
	}

	err = p.ExamineProject()
	if err != nil {
		Err("%s", err)
		return
	}

	err = p.Release()
	if err != nil {
		Err("%s", err)
		return
	}

	p.OutputData()
}

func parseAndValidateInput() (*project.Project, error) {
	// Github focused values
	var slug, workspace, token string
	// General project preference values
	var cl, tagPrefix, artDir, shaFile, dryrunStr string
	// Meson focused values
	var mesonProject, provides string

	flag.StringVar(&slug, "gh-repository", "", "the github.repository")
	flag.StringVar(&workspace, "gh-workspace", "", "the github.workspace")
	flag.StringVar(&token, "gh-token", "", "the GITHUB_TOKEN that will allow you to tag and release")
	flag.StringVar(&cl, "changelog", "", "the changelog file to examine")
	flag.StringVar(&tagPrefix, "tag-prefix", "", "the tag prefix to use")
	flag.StringVar(&artDir, "artifact-dir", "", "the artifact dir to use")
	flag.StringVar(&shaFile, "shasum-file", "", "the shasum filename to use")
	flag.StringVar(&mesonProject, "meson-project", "false", "if this is a meson project")
	flag.StringVar(&provides, "meson-provides", "", "the meson dependency name")
	flag.StringVar(&dryrunStr, "dry-run", "false", "if this is a dry run")
	flag.Parse()

	dryrun := false
	switch dryrunStr {
	case "true":
		dryrun = true
	case "false":
	default:
		return nil, errBoolFormatError
	}

	opts := project.ProjectOpts{
		Slug:          slug,
		BasePath:      workspace,
		Token:         token,
		TagPrefix:     tagPrefix,
		ChangelogFile: cl,
		ArtifactDir:   artDir,
		SHASumFile:    shaFile,
		Log:           Info,
	}

	switch mesonProject {
	case "true":
		opts.Meson = &project.Meson{
			Provides: provides,
		}
	case "false":
	default:
		return nil, errBoolFormatError
	}

	p, err := project.NewProject(opts, dryrun)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func Info(format string, v ...interface{}) {
	fmt.Printf("\x1b[1;34m"+format+"\x1b[0m\n", v...)
}

func Err(format string, v ...interface{}) {
	fmt.Printf("\x1b[1;31m"+format+"\x1b[0m\n", v...)
}
