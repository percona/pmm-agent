package jobs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	// How many times check if backup/restore operation was started
	maxBackupChecks     = 10
	maxRestoreChecks    = 10
	cmdTimeout          = time.Minute
	resyncTimeout       = 5 * time.Minute
	statusCheckInterval = 3 * time.Second
)

type pbmSeverity int

const (
	Fatal pbmSeverity = iota
	Error
	Warning
	Info
	Debug
)

func (s pbmSeverity) String() string {
	switch s {
	case Fatal:
		return "F"
	case Error:
		return "E"
	case Warning:
		return "W"
	case Info:
		return "I"
	case Debug:
		return "D"
	default:
		return ""
	}
}

type pbmLogEntry struct {
	TS         int64 `json:"ts"`
	pbmLogKeys `json:",inline"`
	Msg        string `json:"msg"`
}

type pbmLogKeys struct {
	Severity pbmSeverity `json:"s"`
	RS       string      `json:"rs"`
	Node     string      `json:"node"`
	Event    string      `json:"e"`
	ObjName  string      `json:"eobj"`
	OPID     string      `json:"opid,omitempty"`
}

type pbmBackup struct {
	Name    string `json:"name"`
	Storage string `json:"storage"`
}

type pbmSnapshot struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Error      string `json:"error"`
	CompleteTS int    `json:"completeTS"`
	PbmVersion string `json:"pbmVersion"`
}

type pbmList struct {
	Snapshots []pbmSnapshot `json:"snapshots"`
	Pitr      struct {
		On     bool        `json:"on"`
		Ranges interface{} `json:"ranges"`
	} `json:"pitr"`
}

type pbmStatus struct {
	Backups struct {
		Type       string        `json:"type"`
		Path       string        `json:"path"`
		Region     string        `json:"region"`
		Snapshot   []pbmSnapshot `json:"snapshot"`
		PitrChunks struct {
			Size int `json:"size"`
		} `json:"pitrChunks"`
	} `json:"backups"`
	Cluster []struct {
		Rs    string `json:"rs"`
		Nodes []struct {
			Host  string `json:"host"`
			Agent string `json:"agent"`
			Ok    bool   `json:"ok"`
		} `json:"nodes"`
	} `json:"cluster"`
	Pitr struct {
		Conf bool `json:"conf"`
		Run  bool `json:"run"`
	} `json:"pitr"`
	Running struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		StartTS int    `json:"startTS"`
		Status  string `json:"status"`
		OpID    string `json:"opID"`
	} `json:"running"`
}

func getPBMOutput(ctx context.Context, dbURL *url.URL, to interface{}, args ...string) error {
	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	args = append(args, "--out=json", "--mongodb-uri="+dbURL.String())
	cmd := exec.CommandContext(nCtx, pbmBin, args...) // #nosec G204

	b, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return errors.New(string(exitErr.Stderr))
		}
		return err
	}

	return json.Unmarshal(b, to)
}

func pbmSetupS3(ctx context.Context, l logrus.FieldLogger, dbURL *url.URL, prefix string, s3Config *S3LocationConfig, resync bool) error {
	l.Info("Configuring S3 location.")
	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	confFile, err := writePBMConfigFile(prefix, s3Config)
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.Remove(confFile) //nolint:errcheck

	output, err := exec.CommandContext( //nolint:gosec
		nCtx,
		pbmBin,
		"config",
		"--mongodb-uri="+dbURL.String(),
		"--file="+confFile,
	).CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "pbm config error: %s", string(output))
	}

	if resync {
		nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
		defer cancel()

		output, err = exec.CommandContext( //nolint:gosec
			nCtx,
			pbmBin,
			"config",
			"--mongodb-uri="+dbURL.String(),
			"--force-resync",
		).CombinedOutput()

		if err != nil {
			return errors.Wrapf(err, "pbm config error: %s", string(output))
		}
	}

	return nil
}

type pbmStatusCondition func(s pbmStatus) (bool, error)

func pbmNoRunningOperations(s pbmStatus) (bool, error) {
	return s.Running.Status == "", nil
}

func pbmBackupFinished(name string) pbmStatusCondition {
	started := false
	checks := 0
	return func(s pbmStatus) (bool, error) {
		checks++
		if s.Running.Type == "backup" && s.Running.Name == name && s.Running.Status != "" {
			started = true
		}
		if !started && checks > maxBackupChecks {
			return false, errors.New("failed to start backup")
		}
		var snapshot *pbmSnapshot
		for _, snap := range s.Backups.Snapshot {
			if snap.Name == name {
				snapshot = &snap
				break
			}
		}
		if snapshot == nil || s.Running.Status != "" {
			return false, nil
		}

		if snapshot.Status == "error" {
			return false, errors.New(snapshot.Error)
		}
		return snapshot.Status == "done", nil
	}
}

func pbmRestoreFinished(name string) pbmStatusCondition {
	started := false
	checks := 0
	return func(s pbmStatus) (bool, error) {
		checks++
		if s.Running.Type == "restore" && s.Running.Name == name && s.Running.Status != "" {
			started = true
		}
		if !started {
			if checks > maxRestoreChecks {
				return false, errors.New("failed to start backup")
			}
			return false, nil
		}
		return s.Running.Status == "", nil
	}
}

func waitForPBMState(ctx context.Context, l logrus.FieldLogger, dbURL *url.URL, cond pbmStatusCondition) error {
	l.Info("Waiting for pbm operations completion.")

	ticker := time.NewTicker(statusCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var status pbmStatus
			if err := getPBMOutput(ctx, dbURL, &status, "status"); err != nil {
				return errors.Wrapf(err, "pbm status error")
			}
			done, err := cond(status)
			if err != nil {
				return errors.Wrapf(err, "condition failed")
			}
			if done {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func writePBMConfigFile(prefix string, s3Config *S3LocationConfig) (string, error) {
	tmp, err := ioutil.TempFile("", "pbm-config-*.yml")
	if err != nil {
		return "", errors.Wrap(err, "failed to create pbm configuration file")
	}

	var conf struct {
		Storage struct {
			Type string `yaml:"type"`
			S3   struct {
				Region      string `yaml:"region"`
				Bucket      string `yaml:"bucket"`
				Prefix      string `yaml:"prefix"`
				EndpointURL string `yaml:"endpointUrl"`
				Credentials struct {
					AccessKeyID     string `yaml:"access-key-id"`
					SecretAccessKey string `yaml:"secret-access-key"`
				}
			} `yaml:"s3"`
		} `yaml:"storage"`
	}

	conf.Storage.Type = "s3"
	conf.Storage.S3.EndpointURL = s3Config.Endpoint
	conf.Storage.S3.Region = s3Config.BucketRegion
	conf.Storage.S3.Bucket = s3Config.BucketName
	conf.Storage.S3.Prefix = prefix
	conf.Storage.S3.Credentials.AccessKeyID = s3Config.AccessKey
	conf.Storage.S3.Credentials.SecretAccessKey = s3Config.SecretKey

	bytes, err := yaml.Marshal(&conf)
	if err != nil {
		tmp.Close() //nolint:errcheck
		return "", errors.Wrap(err, "failed to marshall pbm configuration")
	}

	if _, err := tmp.Write(bytes); err != nil {
		tmp.Close() //nolint:errcheck
		return "", errors.Wrap(err, "failed to write pbm configuration file")
	}

	return tmp.Name(), tmp.Close()
}
