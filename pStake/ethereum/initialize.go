package ethereum

import (
	"github.com/persistenceOne/persistenceCore/pStake/constants"
	"github.com/persistenceOne/persistenceCore/pStake/ethereum/contracts"
)

func init() {
	contracts.LiquidStaking.SetABI(constants.LiquidStakingABI)
	contracts.TokenWrapper.SetABI(constants.TokenWrapperABI)
	contracts.STokens.SetABI(constants.STokensABI)
}
