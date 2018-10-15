package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	actual := Parse("this is a {{ . }}", "string")
	expected := "this is a string"
	assert.Equal(t, expected, actual)
}
