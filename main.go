// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
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
