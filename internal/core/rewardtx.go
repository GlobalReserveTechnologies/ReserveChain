package core

// RewardEntry represents the payout allocated to a single operator
// for a given epoch. At the economics level this is usually computed
// from NodeWorkEpochResult and the per-epoch issuance budget.
type RewardEntry struct {
	OperatorID string  `json:"operator_id" yaml:"operator_id"`
	AmountGRC  float64 `json:"amount_grc" yaml:"amount_grc"`
}

// RewardTx is a specialised transaction-like structure that describes
// how the epoch reward budget is split across operators. In a future
// mainnet build this will either become a first-class transaction type
// or be embedded directly into the block header / body. For now it is
// defined as a pure data container so the economics and consensus
// layers have a shared shape to work against.
type RewardTx struct {
	// EpochIndex identifies the epoch this reward decision belongs to.
	// The mapping from block height / time to EpochIndex is defined by
	// the consensus / scheduling layer.
	EpochIndex uint64 `json:"epoch_index" yaml:"epoch_index"`

	// EpochStartUnix and EpochEndUnix are optional wall-clock anchors
	// (seconds since Unix epoch) for observability and audit purposes.
	// They are not strictly required for deterministic validation but
	// are helpful in operator tooling.
	EpochStartUnix int64 `json:"epoch_start_unix" yaml:"epoch_start_unix"`
	EpochEndUnix   int64 `json:"epoch_end_unix" yaml:"epoch_end_unix"`

	// LeaderValidatorID is the validator that authored this RewardTx.
	// In the long-term design this must match the deterministically
	// selected economic leader for the epoch.
	LeaderValidatorID string `json:"leader_validator_id" yaml:"leader_validator_id"`

	// TotalRewardGRC is the total reward budget for this epoch as
	// computed by the issuance curve. The sum of Entries.AmountGRC
	// should never exceed this value.
	TotalRewardGRC float64 `json:"total_reward_grc" yaml:"total_reward_grc"`

	// WorkRoot is an optional commitment to the underlying work metrics
	// used to derive the payouts (for example, a Merkle root). The
	// exact construction is defined at the economics / consensus layer.
	WorkRoot [32]byte `json:"work_root" yaml:"work_root"`

	// Entries enumerates the individual operator payouts for this epoch.
	Entries []RewardEntry `json:"entries" yaml:"entries"`
}
