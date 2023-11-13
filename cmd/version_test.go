package cmd

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	testCases := []struct {
		name                 string
		buildDate            string
		buildCommit          string
		buildVersionOverride string
		output               string
		shortened            bool
		errMsg               string
	}{
		{
			name:                 "error",
			buildDate:            time.Now().String(),
			buildCommit:          "243413535",
			buildVersionOverride: "0.0.1",
			output:               "unknown",
			errMsg:               "FF23016: Invalid output type: unknown",
		},
		{
			name:                 "yaml output",
			buildDate:            time.Now().String(),
			buildCommit:          "243413535",
			buildVersionOverride: "0.0.1",
			output:               "yaml",
		},
		{
			name:                 "json output",
			buildDate:            time.Now().String(),
			buildCommit:          "243413535",
			buildVersionOverride: "0.0.1",
			output:               "json",
		},
		{
			name:                 "shortened",
			buildDate:            time.Now().String(),
			buildCommit:          "243413535",
			buildVersionOverride: "0.0.1",
			shortened:            true,
		},
		{
			name:                 "version is empty",
			buildDate:            time.Now().String(),
			buildCommit:          "243413535",
			buildVersionOverride: "",
			output:               "json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command := versionCommand()
			BuildVersionOverride = tc.buildVersionOverride
			BuildCommit = tc.buildCommit
			BuildDate = tc.buildDate
			command.Flags().Set("output", tc.output)
			command.Flags().Set("short", strconv.FormatBool(tc.shortened))
			err := command.RunE(command, []string{"arg1"})

			if tc.errMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
