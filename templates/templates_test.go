// Package templates is based on https://github.com/moby/moby/blob/503b1a9b6f24488db6a67f7ba24258e4ff5ea2a7/daemon/logger/templates/templates.go
package templates

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParse(t *testing.T) {
	tm, err := NewParse("foo", "this is a {{ . }}")
	assert.NoError(t, err)

	var b bytes.Buffer
	assert.NoError(t, tm.Execute(&b, "string"))
	assert.Equal(t, "this is a string", b.String())
}
