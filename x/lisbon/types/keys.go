package types

// constants
const (
	// module name
	ModuleName = "lisbon"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the inflation persistent store
//const (
//	prefixPeriod = iota + 1
//	prefixEpochMintProvision
//	prefixEpochIdentifier
//	prefixEpochsPerPeriod
//	prefixSkippedEpochs
//)

// KVStore key prefixes
//var (
//	KeyPrefixPeriod             = []byte{prefixPeriod}
//	KeyPrefixEpochMintProvision = []byte{prefixEpochMintProvision}
//	KeyPrefixEpochIdentifier    = []byte{prefixEpochIdentifier}
//	KeyPrefixEpochsPerPeriod    = []byte{prefixEpochsPerPeriod}
//	KeyPrefixSkippedEpochs      = []byte{prefixSkippedEpochs}
//)

var (
	//FeePoolKey                        = []byte{0x00} // key for global distribution state
	ProposerKey = []byte{0x01} // key for the proposer operator address
	//ValidatorOutstandingRewardsPrefix = []byte{0x02} // key for outstanding rewards
	//
	//DelegatorWithdrawAddrPrefix          = []byte{0x03} // key for delegator withdraw address
	//DelegatorStartingInfoPrefix          = []byte{0x04} // key for delegator starting info
	//ValidatorHistoricalRewardsPrefix     = []byte{0x05} // key for historical validators rewards / stake
	//ValidatorCurrentRewardsPrefix        = []byte{0x06} // key for current validator rewards
	//ValidatorAccumulatedCommissionPrefix = []byte{0x07} // key for accumulated validator commission
	//ValidatorSlashEventPrefix            = []byte{0x08} // key for validator slash fraction
)
