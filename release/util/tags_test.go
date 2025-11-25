package util_test

import (
	"testing"

	"github.com/distribution/reference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferences(t *testing.T) {
	named, err := reference.ParseNamed("example.com:454/my/repo:v1.2.3@sha256:40c70689234e535d783a744b5a870fb1fb5b2f6c2ae19a34f25258d6ea72723b")
	require.NoError(t, err)
	assert.Equal(t, "example.com:454/my/repo", named.Name())
}
