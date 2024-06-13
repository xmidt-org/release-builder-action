// SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
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
