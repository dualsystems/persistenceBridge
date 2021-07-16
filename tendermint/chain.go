package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/cosmos/relayer/helpers"
	"github.com/cosmos/relayer/relayer"
	"github.com/persistenceOne/persistenceBridge/utilities/logging"
	tendermintService "github.com/tendermint/tendermint/libs/service"
)

func InitializeAndStartChain(chainConfigJsonPath, timeout, homePath string) (*relayer.Chain, error) {
	chain, err := fileInputAdd(chainConfigJsonPath)
	to, err := time.ParseDuration(timeout)
	if err != nil {
		return chain, err
	}

	err = chain.Init(homePath, to, nil, true)
	if err != nil {
		return chain, err
	}

	if chain.KeyExists(chain.Key) {
		logging.Info("deleting old key", chain.Key)
		err = chain.Keybase.Delete(chain.Key)
		if err != nil {
			return chain, err
		}
	}

	//118 is not being used anywhere
	ko, err := helpers.KeyAddOrRestore(chain, chain.Key, uint32(118))
	if err != nil {
		return chain, err
	}

	logging.Warn("Chain Keys added  [NOT TO BE USED]:", ko.Address)

	if err = chain.Start(); err != nil {
		if err != tendermintService.ErrAlreadyStarted {
			chain.Error(err)
			return chain, err
		}
	}
	return chain, nil
}

func fileInputAdd(file string) (*relayer.Chain, error) {
	// If the user passes in a file, attempt to read the chain configuration from that file
	c := &relayer.Chain{}
	if _, err := os.Stat(file); err != nil {
		return c, err
	}

	byt, err := ioutil.ReadFile(file)
	if err != nil {
		return c, err
	}

	if err = json.Unmarshal(byt, c); err != nil {
		return c, err
	}

	return c, nil
}
