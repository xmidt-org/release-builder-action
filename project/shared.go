// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package project

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
)

func sha(fs *afero.Afero, file string) ([]byte, error) {

	f, err := fs.Open(file)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: unable to open file '%s'", err, file)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "unable to close file '%s': %v\n", file, cerr)
			return
		}
	}()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: unable to perform SHA256 against file '%s'", err, file)
	}

	return h.Sum(nil), nil
}

func generateSha256Sum(fs *afero.Afero, name, path string) error {
	files, err := fs.ReadDir(path)
	if err != nil {
		return fmt.Errorf("%w: unable to read directory '%s'", err, path)
	}

	shaFile := path + "/" + name
	var lines []string

	for _, file := range files {
		if name == file.Name() {
			continue
		}
		fn := path + "/" + file.Name()

		b, err := sha(fs, fn)
		if err != nil {
			return err
		}

		lines = append(lines, fmt.Sprintf("%x  %s", b, file.Name()))
	}

	f, err := fs.Create(shaFile)
	if err != nil {
		return fmt.Errorf("%w: unable to create file '%s'", err, shaFile)
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to close file '%s': %v\n", f, err)
			return
		}
	}()

	for _, line := range lines {
		_, err = fmt.Fprintln(f, line)
		if err != nil {
			return fmt.Errorf("%w: unable to write to file '%s'", err, shaFile)
		}
	}

	return nil
}

func mkdir(fs *afero.Afero, path string) error {
	fi, err := fs.Stat(path)
	if err == nil {
		if fi.Mode().IsDir() {
			return nil
		}

		return fmt.Errorf("%w: path '%s'", errPathNotDirectory, path)
	}

	if os.IsNotExist(err) {
		err := fs.MkdirAll(path, os.ModePerm)
		if err == nil {
			return nil
		}

		return fmt.Errorf("%w: unable to make directory '%s'", err, path)
	}

	return fmt.Errorf("%w: directory check failed for: '%s'", err, path)
}
