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

package jobs

type Storage struct {
	Type string `yaml:"type"`
	S3   S3     `yaml:"s3"`
}

type S3 struct {
	Region      string      `yaml:"region"`
	Bucket      string      `yaml:"bucket"`
	Prefix      string      `yaml:"prefix"`
	EndpointURL string      `yaml:"endpointUrl"`
	Credentials Credentials `yaml:"credentials"`
}

type Credentials struct {
	AccessKeyID     string `yaml:"access-key-id"`
	SecretAccessKey string `yaml:"secret-access-key"`
}

type PITR struct {
	Enabled bool `yaml:"enabled"`
}

type PBMConfig struct {
	Storage Storage `yaml:"storage"`
	PITR    PITR    `yaml:"pitr"`
}

func (c *PBMConfig) Equals(config *PBMConfig) bool {
	if c.PITR.Enabled != config.PITR.Enabled {
		return false
	}

	if c.Storage.Type != config.Storage.Type {
		return false
	}

	if c.Storage.S3.Bucket != c.Storage.S3.Bucket {
		return false
	}

	if c.Storage.S3.EndpointURL != config.Storage.S3.EndpointURL {
		return false
	}

	if c.Storage.S3.Prefix != config.Storage.S3.Prefix {
		return false
	}

	if c.Storage.S3.Region != config.Storage.S3.Region {
		return false
	}

	return true
}
