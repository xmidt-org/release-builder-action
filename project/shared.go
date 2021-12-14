package project

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
	defer f.Close()

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
	defer f.Close()

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
