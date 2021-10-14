package commands

import (
	"github.com/BurntSushi/toml"
	"github.com/persistenceOne/persistenceBridge/application/configuration"
	constants2 "github.com/persistenceOne/persistenceBridge/application/constants"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	cmd := InitCommand()
	err := cmd.Flags().Set(constants2.FlagPBridgeHome, constants2.TestHomeDir)
	require.Equal(t, nil, err)
	err = cmd.Execute()
	require.Equal(t, nil, err)

	config := configuration.GetAppConfig()
	_, err = toml.DecodeFile(filepath.Join(constants2.TestHomeDir, "config.toml"), &config)
	require.Equal(t, nil, err)

}
