package commands

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/persistenceOne/persistenceBridge/application"
	"github.com/persistenceOne/persistenceBridge/application/casp"
	"github.com/persistenceOne/persistenceBridge/application/configuration"
	constants2 "github.com/persistenceOne/persistenceBridge/application/constants"
	db2 "github.com/persistenceOne/persistenceBridge/application/db"
	"github.com/persistenceOne/persistenceBridge/application/rpc"
	"github.com/persistenceOne/persistenceBridge/application/shutdown"
	ethereum2 "github.com/persistenceOne/persistenceBridge/ethereum"
	"github.com/persistenceOne/persistenceBridge/kafka"
	"github.com/persistenceOne/persistenceBridge/kafka/utils"
	tendermint2 "github.com/persistenceOne/persistenceBridge/tendermint"
	"github.com/persistenceOne/persistenceBridge/utilities/logging"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func StartCommand() *cobra.Command {
	pBridgeCommand := &cobra.Command{
		Use:   "start",
		Short: "starts persistenceBridge",
		RunE: func(cmd *cobra.Command, args []string) error {

			homePath, err := cmd.Flags().GetString(constants2.FlagPBridgeHome)
			if err != nil {
				log.Fatalln(err)
			}

			showDebugLog, err := cmd.Flags().GetBool(constants2.FlagShowDebugLog)
			if err != nil {
				log.Fatalln(err)
			}
			logging.ShowDebugLog(showDebugLog)

			pStakeConfig := configuration.InitConfig()
			_, err = toml.DecodeFile(filepath.Join(homePath, "config.toml"), &pStakeConfig)
			if err != nil {
				log.Fatalf("Error decoding pStakeConfig file: %v\n", err.Error())
			}
			pStakeConfig = configuration.SetConfig(cmd)

			tmAddress, err := casp.GetTendermintAddress()
			if err != nil {
				log.Fatalln(err)
			}

			ethAddress, err := casp.GetEthAddress()
			if err != nil {
				log.Fatalln(err)
			}

			configuration.SetPStakeAddress(tmAddress)
			configuration.ValidateAndSeal()
			setBech32Prefixes()

			err = logging.InitializeBot()
			if err != nil {
				log.Fatalln(err)
			}

			logging.Info("Bridge (Wrap) Tendermint Address:", tmAddress.String())
			logging.Info("Bridge (Admin) Ethereum Address:", ethAddress.String())

			tmSleepTime, err := cmd.Flags().GetInt(constants2.FlagTendermintSleepTime)
			if err != nil {
				log.Fatalln(err)
			}

			tmStart, err := cmd.Flags().GetInt64(constants2.FlagTendermintStartHeight)
			if err != nil {
				log.Fatalln(err)
			}

			ethSleepTime, err := cmd.Flags().GetInt(constants2.FlagEthereumSleepTime)
			if err != nil {
				log.Fatalln(err)
			}

			ethStart, err := cmd.Flags().GetInt64(constants2.FlagEthereumStartHeight)
			if err != nil {
				log.Fatalln(err)
			}

			timeout, err := cmd.Flags().GetString(constants2.FlagTimeOut)
			if err != nil {
				log.Fatalln(err)
			}

			db, err := db2.InitializeDB(homePath+"/db", tmStart, ethStart)
			if err != nil {
				log.Fatalln(err)
			}
			defer func(db *badger.DB) {
				err := db.Close()
				if err != nil {
					log.Println("Error while closing DB: ", err.Error())
				}
			}(db)

			unboundEpochTime, err := db2.GetUnboundEpochTime()
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("unbound epoch time: %d\n", unboundEpochTime.Epoch)

			chain, err := tendermint2.InitializeAndStartChain(timeout, homePath)
			if err != nil {
				log.Fatalln(err)
			}

			ethereumClient, err := ethclient.Dial(pStakeConfig.Ethereum.EthereumEndPoint)
			if err != nil {
				log.Fatalf("Error while dialing to eth orchestrator %s: %s\n", pStakeConfig.Ethereum.EthereumEndPoint, err.Error())
			}

			encodingConfig := application.MakeEncodingConfig()
			initClientCtx := client.Context{}.
				WithJSONMarshaler(encodingConfig.Marshaler).
				WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
				WithTxConfig(encodingConfig.TransactionConfig).
				WithLegacyAmino(encodingConfig.Amino).
				WithInput(os.Stdin).
				WithAccountRetriever(authTypes.AccountRetriever{}).
				WithBroadcastMode(configuration.GetAppConfig().Tendermint.BroadcastMode).
				WithHomeDir(homePath)

			protoCodec := codec.NewProtoCodec(initClientCtx.InterfaceRegistry)
			kafkaState := utils.NewKafkaState(pStakeConfig.Kafka.Brokers, homePath, pStakeConfig.Kafka.TopicDetail)
			end := make(chan bool)
			ended := make(chan bool)
			go kafka.KafkaRoutine(kafkaState, protoCodec, chain, ethereumClient, end, ended)

			go rpc.StartServer(pStakeConfig.RPCEndpoint)

			logging.Info("Starting to listen ethereum....")
			go ethereum2.StartListening(ethereumClient, time.Duration(ethSleepTime)*time.Millisecond, pStakeConfig.Kafka.Brokers, protoCodec)

			logging.Info("Starting to listen tendermint....")
			go tendermint2.StartListening(initClientCtx, chain, pStakeConfig.Kafka.Brokers, protoCodec, time.Duration(tmSleepTime)*time.Millisecond)

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			for sig := range signalChan {
				logging.Info("STOP SIGNAL RECEIVED: " + sig.String())
				shutdown.SetBridgeStopSignal(true)
				for {
					if !shutdown.GetKafkaConsumerClosed() {
						logging.Info("Stopping Kafka Routine!!!")
						kafka.KafkaClose(kafkaState, end, ended)()
						shutdown.SetKafkaConsumerClosed(true)
					}
					if shutdown.GetTMStopped() && shutdown.GetETHStopped() && shutdown.GetKafkaConsumerClosed() {
						return nil
					}
					time.Sleep(100 * time.Millisecond)
				}
			}

			return nil
		},
	}
	pBridgeCommand.Flags().String(constants2.FlagTimeOut, constants2.DefaultTimeout, "timeout time for connecting to rpc")
	pBridgeCommand.Flags().String(constants2.FlagEthereumEndPoint, constants2.DefaultEthereumEndPoint, "ethereum orchestrator to connect")
	pBridgeCommand.Flags().String(constants2.FlagPBridgeHome, constants2.DefaultPBridgeHome, "home for pBridge")
	pBridgeCommand.Flags().Bool(constants2.FlagShowDebugLog, false, "show debug logs")
	pBridgeCommand.Flags().String(constants2.FlagEthereumEndPoint, "", "ethereum orchestrator to connect")
	pBridgeCommand.Flags().String(constants2.FlagKafkaPorts, "", "ports kafka brokers are running on, --ports 192.100.10.10:443,192.100.10.11:443")
	pBridgeCommand.Flags().Int(constants2.FlagTendermintSleepTime, constants2.DefaultTendermintSleepTime, "sleep time between block checking for tendermint in ms")
	pBridgeCommand.Flags().Int(constants2.FlagEthereumSleepTime, constants2.DefaultEthereumSleepTime, "sleep time between block checking for ethereum in ms")
	pBridgeCommand.Flags().Int64(constants2.FlagTendermintStartHeight, constants2.DefaultTendermintStartHeight, fmt.Sprintf("Start checking height on tendermint chain from this height (default %d - starts from where last left)", constants2.DefaultTendermintStartHeight))
	pBridgeCommand.Flags().Int64(constants2.FlagEthereumStartHeight, constants2.DefaultEthereumStartHeight, fmt.Sprintf("Start checking height on ethereum chain from this height (default %d - starts from where last left)", constants2.DefaultEthereumStartHeight))
	pBridgeCommand.Flags().String(constants2.FlagDenom, "", "denom name")
	pBridgeCommand.Flags().String(constants2.FlagAccountPrefix, "", "account prefix on tendermint chains")
	pBridgeCommand.Flags().String(constants2.FlagTendermintNode, constants2.DefaultTendermintNode, "tendermint rpc node url")
	pBridgeCommand.Flags().String(constants2.FlagTendermintChainID, constants2.DefaultTendermintChainId, "tendermint rpc node url")
	pBridgeCommand.Flags().Uint64(constants2.FlagEthGasLimit, 0, "Gas limit for eth txs")
	pBridgeCommand.Flags().String(constants2.FlagBroadcastMode, "", "broadcast mode for tendermint")
	pBridgeCommand.Flags().String(constants2.FlagCASPURL, "", "casp api url (with http)")
	pBridgeCommand.Flags().String(constants2.FlagCASPVaultID, "", "casp vault id")
	pBridgeCommand.Flags().String(constants2.FlagCASPApiToken, "", "casp api token (in format: Bearer ...)")
	pBridgeCommand.Flags().String(constants2.FlagCASPTMPublicKey, "", "casp tendermint public key")
	pBridgeCommand.Flags().String(constants2.FlagCASPEthPublicKey, "", "casp ethereum public key")
	pBridgeCommand.Flags().Int(constants2.FlagCASPSignatureWaitTime, -1, "csap siganture wait time")
	pBridgeCommand.Flags().Int(constants2.FlagCASPMaxGetSignatureAttempts, 0, "casp max attempts to fetch operation id")
	pBridgeCommand.Flags().String(constants2.FlagRPCEndpoint, "", "rpc Endpoint for server")
	pBridgeCommand.Flags().Int64(constants2.FlagMinimumWrapAmount, -1, "minimum amount in send coin tx to wrap onto eth")
	pBridgeCommand.Flags().String(constants2.FlagTelegramBotToken, "", "telegram bot token")
	pBridgeCommand.Flags().Int64(constants2.FlagTelegramChatID, 0, "telegram chat id")
	//This will always be used from flag
	pBridgeCommand.Flags().Bool(constants2.FlagCASPConcurrentKey, true, "allows starting multiple sign operations that specify the same key")

	return pBridgeCommand
}

func setBech32Prefixes() {
	if configuration.GetAppConfig().Tendermint.AccountPrefix == "" {
		panic("account prefix is empty")
	}
	bech32PrefixAccAddr := configuration.GetAppConfig().Tendermint.AccountPrefix
	bech32PrefixAccPub := configuration.GetAppConfig().Tendermint.AccountPrefix + sdkTypes.PrefixPublic
	bech32PrefixValAddr := configuration.GetAppConfig().Tendermint.AccountPrefix + sdkTypes.PrefixValidator + sdkTypes.PrefixOperator
	bech32PrefixValPub := configuration.GetAppConfig().Tendermint.AccountPrefix + sdkTypes.PrefixValidator + sdkTypes.PrefixOperator + sdkTypes.PrefixPublic
	bech32PrefixConsAddr := configuration.GetAppConfig().Tendermint.AccountPrefix + sdkTypes.PrefixValidator + sdkTypes.PrefixConsensus
	bech32PrefixConsPub := configuration.GetAppConfig().Tendermint.AccountPrefix + sdkTypes.PrefixValidator + sdkTypes.PrefixConsensus + sdkTypes.PrefixPublic

	bech32Configuration := sdkTypes.GetConfig()
	bech32Configuration.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)
	bech32Configuration.SetBech32PrefixForValidator(bech32PrefixValAddr, bech32PrefixValPub)
	bech32Configuration.SetBech32PrefixForConsensusNode(bech32PrefixConsAddr, bech32PrefixConsPub)
	bech32Configuration.Seal()
}
