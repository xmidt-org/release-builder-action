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

import "github.com/stretchr/testify/mock"

type mockGit struct {
	mock.Mock
}

func (m *mockGit) IsTagPresent(tag string) (bool, error) {
	args := m.Called(tag)
	return args.Bool(0), args.Error(1)
}

func (m *mockGit) TagHead(ver, msg string) error {
	args := m.Called(ver, msg)
	return args.Error(0)
}

func (m *mockGit) PushTags(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *mockGit) CreateArchive(slug, ver, fmt, dir string) (string, error) {
	args := m.Called(slug, ver, fmt, dir)
	return args.String(0), args.Error(1)
}
