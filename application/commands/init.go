package commands

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/persistenceOne/persistenceBridge/application/configuration"
	constants2 "github.com/persistenceOne/persistenceBridge/application/constants"
	"github.com/spf13/cobra"
)

func InitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "creates config.toml file",
		RunE: func(cmd *cobra.Command, args []string) error {

			config := configuration.SetConfig(cmd)

			var buf bytes.Buffer
			encoder := toml.NewEncoder(&buf)
			if err := encoder.Encode(config); err != nil {
				panic(err)
			}

			homeDir, err := cmd.Flags().GetString(constants2.FlagPBridgeHome)
			if err != nil {
				panic(err)
			}
			if err = os.MkdirAll(homeDir, os.ModePerm); err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(filepath.Join(homeDir, "config.toml"), buf.Bytes(), 0600); err != nil {
				panic(err)
			}
			log.Println("generated configuration file at ", filepath.Join(homeDir, "config.toml"))

			return nil
		},
	}
	//This will always be used from flag
	cmd.Flags().String(constants2.FlagPBridgeHome, constants2.DefaultPBridgeHome, "home for pBridge")

	// Tendermint
	cmd.Flags().String(constants2.FlagDenom, constants2.DefaultDenom, "denom name")
	cmd.Flags().Int64(constants2.FlagTMAvgBlockTime, constants2.DefaultTendermintAvgBlockTime.Milliseconds(), "avg block of tm chain (in ms)")
	cmd.Flags().String(constants2.FlagAccountPrefix, constants2.DefaultAccountPrefix, "account prefix on tendermint chains")
	cmd.Flags().String(constants2.FlagTendermintNode, constants2.DefaultTendermintNode, "tendermint rpc node url")
	cmd.Flags().Uint32(constants2.FlagTendermintCoinType, constants2.DefaultTendermintCoinType, "tendermint address coin type")
	cmd.Flags().String(constants2.FlagTendermintChainID, constants2.DefaultTendermintChainId, "chain id of tendermint node")
	cmd.Flags().String(constants2.FlagTMGasPrice, constants2.DefaultTendermintGasPrice, "tendermint gas price (should be a float value)")
	cmd.Flags().Float64(constants2.FlagTMGasAdjustment, constants2.DefaultTendermintGasAdjustment, "tendermint gas adjustment (should be a float value and greater than 1.0)")
	cmd.Flags().String(constants2.FlagBroadcastMode, constants2.DefaultBroadcastMode, "broadcast mode for tendermint")
	cmd.Flags().Int64(constants2.FlagMinimumWrapAmount, constants2.DefaultMinimumWrapAmount, "minimum amount in send coin tx to wrap onto eth")

	// Ethereum
	cmd.Flags().String(constants2.FlagEthereumEndPoint, constants2.DefaultEthereumEndPoint, "ethereum orchestrator to connect")
	cmd.Flags().Uint64(constants2.FlagEthGasLimit, constants2.DefaultEthGasLimit, "Gas limit for eth txs")
	cmd.Flags().Int64(constants2.FlagEthGasFeeCap, constants2.DefaultEthGasFeeCap, "Gas fee cap for eth txs")
	cmd.Flags().String(constants2.FlagTokenWrapperAddress, constants2.DefaultEthZeroAddress, "sc address of token wrapper")
	cmd.Flags().String(constants2.FlagLiquidStakingAddress, constants2.DefaultEthZeroAddress, "sc address of liquid staking")

	// Kafka
	cmd.Flags().String(constants2.FlagKafkaPorts, constants2.DefaultKafkaPorts, "ports kafka brokers are running on, --ports 192.100.10.10:443,192.100.10.11:443")

	// CASP
	cmd.Flags().String(constants2.FlagCASPURL, "", "casp api url (with http)")
	cmd.Flags().String(constants2.FlagCASPVaultID, "", "casp vault id")
	cmd.Flags().String(constants2.FlagCASPTMPublicKey, "", "casp tendermint public key")
	cmd.Flags().String(constants2.FlagCASPEthPublicKey, "", "casp ethereum public key")
	cmd.Flags().Int(constants2.FlagCASPWaitTime, int(constants2.DefaultCASPWaitTime.Seconds()), "casp wait time (in seconds)")
	cmd.Flags().Bool(constants2.FlagCASPConcurrentKey, true, "allows starting multiple sign operations that specify the same key")
	cmd.Flags().Uint(constants2.FlagCASPMaxAttempts, constants2.DefaultCASPMaxAttempts, "max attempts for getting signature for an operation id and posting data to casp for generating signature")

	// Telegram alerting service
	cmd.Flags().String(constants2.FlagTelegramBotToken, "", "telegram bot token")
	cmd.Flags().Int64(constants2.FlagTelegramChatID, 0, "telegram chat id")

	// Others
	cmd.Flags().String(constants2.FlagRPCEndpoint, constants2.DefaultRPCEndpoint, "rpc Endpoint for server")

	return cmd
}
