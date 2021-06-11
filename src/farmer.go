package main

import (
	"math"
)

const StandWallet = 0

var CoinUnit = map[string]float64{
	"chia":         math.Pow10(12),
	"mojo":         1,
	"colouredcoin": math.Pow10(3),
}

type Farmer struct {
	MinerId             string
	PuzzleHash          string
	PowerAvailable      uint64
	TotalBlockAward     float64
	BalanceMinerAccount float64
}

type FarmedAmount struct {
	TotalFarmedAmount  float64 `json:"farmed_amount"`
	PoolRewardAmount   float64 `json:"pool_reward_amount"`
	FarmerRewardAmount float64 `json:"farmer_reward_amount"`
	FeeAmount          float64 `json:"fee_amount"`
	LastHeightFarmed   uint64  `json:"last_height_farmed"`
}

type WalletAddress struct {
	WalletId uint   `json:"wallet_id"`
	Address  string `json:"address"`
}

type Wallet struct {
	ID   uint
	Type int64
}
type WalletStats struct {
	WalletId           uint    `json:"wallet_id"`
	Address            string  `json:"address"`
	Balance            float64 `json:"balance"`
	PuzzleHash         string  `json:"puzzle_hash"`
	FarmerRewardAmount float64 `json:"farmer_reward_amount"`
	PoolRewardAmount   float64 `json:"pool_reward_amount"`
}
type Balances struct {
	WalletId                 uint    `json:"wallet_id"`
	ConfirmedWalletBalance   float64 `json:"confirmed_wallet_balance"`
	UnConfirmedWalletBalance float64 `json:"un_confirmed_wallet_balance"`
	SpendableBalance         float64 `json:"spendable_balance"`
	PendingChange            float64 `json:"pending_change"`
	MaxSendAmount            float64 `json:"max_send_amount"`
}

type WalletResponse struct {
	Wallets []Wallet `json:"wallets"`
}

type Plot struct {
	Filename               string `json:"filename"`
	Size                   uint64 `json:"size"`
	PlotSeed               uint64 `json:"plot-seed"`
	PoolPublicKey          string `json:"pool_public_key"`
	PoolContractPuzzleHash string `json:"pool_contract_puzzle_hash"`
	PlotPublicKey          string `json:"plot_public_key"`
	FileSize               uint64 `json:"file_size"`
	TimeModified           string `json:"time_modified"`
}
type PlotsResponse struct {
	Plots                 []Plot   `json:"plots"`
	FailedToOpenFileNames []string `json:"failed_to_open_file_names"`
	NotFoundFilenames     []string `json:"not_found_filenames"`
}
