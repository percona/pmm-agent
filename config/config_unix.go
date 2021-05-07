// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides access to pmm-agent configuration.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris
//go:build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package config

import (
	"golang.org/x/sys/unix"
)

// IsWritable checks if specified path is writable.
func IsWritable(path string) error {
	return unix.Access(path, unix.W_OK)
}
