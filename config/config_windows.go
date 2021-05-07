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

// +build windows
//go:build windows

package config

import (
	"os"
)

// IsWritable tries to open and write comment to file.
func IsWritable(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte("# Write check\n"))
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}
