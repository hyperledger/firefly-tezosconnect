package cmd

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmdDefault(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersionCmdYAML(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "-o", "yaml"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersionCmdJSON(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "-o", "json"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersionCmdInvalidType(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "-o", "wrong"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.Regexp(t, "FF23016", err)
}

func TestVersionCmdShorthand(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "-s"})
	defer rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestSetBuildInfoWithBI(t *testing.T) {
	info := &Info{}
	setBuildInfo(info, &debug.BuildInfo{Main: debug.Module{Version: "12345"}}, true)
	assert.Equal(t, "12345", info.Version)
}
