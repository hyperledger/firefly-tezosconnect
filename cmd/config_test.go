package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigMarkdown(t *testing.T) {
	rootCmd.SetArgs([]string{"docs"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
