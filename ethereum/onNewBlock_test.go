package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/persistenceOne/persistenceBridge/application/casp"
	"github.com/persistenceOne/persistenceBridge/application/configuration"
	"github.com/persistenceOne/persistenceBridge/application/db"
	"github.com/persistenceOne/persistenceBridge/application/outgoingTx"
	"github.com/persistenceOne/persistenceBridge/kafka/utils"
	test "github.com/persistenceOne/persistenceBridge/utilities/testing"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func TestOnNewBlock(t *testing.T) {
	pStakeConfig := configuration.InitConfig()
	appconfig := test.GetCmdWithConfig()
	configuration.SetConfig(&appconfig)
	tmAddress, err := casp.GetTendermintAddress()
	require.Equal(t, nil, err)

	configuration.SetPStakeAddress(tmAddress)

	ethereumClient, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/b21966541db246d398fb31402eec2c14")
	require.Equal(t, nil, err)
	ctx := context.Background()
	kafkaProducer := utils.NewProducer(pStakeConfig.Kafka.Brokers, utils.SaramaConfig())
	latestEthHeight, err := ethereumClient.BlockNumber(ctx)
	dirname, err := os.UserHomeDir()


	database, err := db.OpenDB(filepath.Join(dirname, "/persistence/persistenceBridge/application") + "/db")
	defer database.Close()

	fmt.Println(latestEthHeight)
	TxhashFail := common.HexToHash("0xb96560e8ef15a0d86f8156daddf6f2421d962f5a37dd8e2ba212b05eddaf0b59")

	Address := common.BytesToAddress([]byte("0xce3f57a8de9aa69da3289871b5fee5e77ffcf480"))
	amt := new(big.Int)
	amt.SetInt64(1000)
	wrapTokenMsg := outgoingTx.WrapTokenMsg{
		Address: Address,
		Amount:  amt,
	}
	txd := []outgoingTx.WrapTokenMsg{wrapTokenMsg}

	ethTransaction := db.EthereumBroadcastedWrapTokenTransaction{
		TxHash:   TxhashFail,
		Messages: txd,
	}

	err = db.SetBroadcastedEthereumTx(ethTransaction)
	require.Equal(t, nil, err)
	err = onNewBlock(ctx, latestEthHeight, ethereumClient, &kafkaProducer)
	require.Equal(t, nil, err)

	TxhashSuccess :=  common.HexToHash("0x8e08d80c37c884467b9b48a77e658711615a5cfde43f95fccfb3b95ee66cd6ea")
	Address = common.BytesToAddress([]byte("0x477573f212a7bdd5f7c12889bd1ad0aa44fb82aa"))
	amt = new(big.Int)
	amt.SetInt64(1000)
	wrapTokenMsg = outgoingTx.WrapTokenMsg{
		Address: Address,
		Amount:  amt,
	}
	txd = []outgoingTx.WrapTokenMsg{wrapTokenMsg}

	ethTransaction = db.EthereumBroadcastedWrapTokenTransaction{
		TxHash:   TxhashSuccess,
		Messages: txd,
	}

	err = db.SetBroadcastedEthereumTx(ethTransaction)
	require.Equal(t, nil, err)
	err = onNewBlock(ctx, latestEthHeight, ethereumClient, &kafkaProducer)
	require.Equal(t, nil, err)


}
