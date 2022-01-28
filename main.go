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
	"os"

	"github.com/xmidt-org/release-builder-action/project"
)

var (
	errBoolFormatError = errors.New("the string should be 'true' or 'false'")
)

func main() {
	os.Exit(run())
}

func run() int {
	p, err := parseAndValidateInput()
	if err != nil {
		Err("Error validating input: %s", err)
		return 1
	}

	err = p.ExamineProject()
	if err != nil {
		Err("Error examining project: %s", err)
		return 1
	}

	err = p.Release()
	if err != nil {
		Err("Error releasing: %s", err)
		return 1
	}

	err = p.OutputData()
	if err != nil {
		Err("Error outputing: %s", err)
		return 1
	}

	return 0
}

func parseAndValidateInput() (*project.Project, error) {
	dryrun := false
	switch os.Getenv("INPUTS_DRY_RUN") {
	case "true":
		dryrun = true
	case "false":
	default:
		return nil, errBoolFormatError
	}

	opts := project.ProjectOpts{
		Slug:          os.Getenv("INPUTS_SLUG"),
		BasePath:      os.Getenv("INPUTS_WORKSPACE"),
		Token:         os.Getenv("INPUTS_TOKEN"),
		TagPrefix:     os.Getenv("INPUTS_TAG_PREFIX"),
		ChangelogFile: os.Getenv("INPUTS_CHANGELOG"),
		ArtifactDir:   os.Getenv("INPUTS_ARTIFACT_DIR"),
		SHASumFile:    os.Getenv("INPUTS_SHASUM_FILE"),
		Log:           Info,
		Meson: project.Meson{
			Provides: os.Getenv("INPUTS_MESON_PROVIDES"),
		},
	}

	Info("BasePath:      '" + opts.BasePath + "'")
	Info("ChangelogFile: '" + opts.ChangelogFile + "'")
	Info("workspace:     '" + workspace + "'")
	Info("cl:            '" + cl + "'")

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
