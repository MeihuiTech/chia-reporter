package main

import (
	"fmt"
	"net/http"
)

type ChiaBlockRecord struct {
	ID                         uint64 `gorm:"primaryKey;<-:false" json:"id"`
	ChallengeBlockInfoHash     string `gorm:"type:varchar(256);not null;default:unknown" json:"challenge_block_info_hash"`
	Deficit                    uint64 `gorm:"type:bigint(20);not null;default:0" json:"deficit"`
	FarmerPuzzleHash           string `gorm:"type:varchar(256);not null;default:unknown;idx_bc_farmer_puzzle_hash" json:"farmer_puzzle_hash"`
	Fees                       uint64 `gorm:"type:bigint(20);not null;default:0" json:"fees"`
	HeaderHash                 string `gorm:"type:varchar(256);not null;default:unknown;index:idx_bc_header_hash" json:"header_hash"`
	Height                     uint64 `gorm:"type:bigint(20);not null;default:0;index:idx_bc_height" json:"height"`
	Overflow                   bool   `gorm:"type:bool;not null;default:false" json:"overflow"`
	PoolPuzzleHash             string `gorm:"type:varchar(256);not null;default:unknown" json:"pool_puzzle_hash"`
	PrevHash                   string `gorm:"type:varchar(256);not null;default:unknown" json:"prev_hash"`
	PrevTransactionBlockHash   string `gorm:"type:varchar(256);not null;default:unknown" json:"prev_transaction_block_hash"`
	PrevTransactionBlockHeight uint64 `gorm:"type:bigint(20);not null;default:0" json:"prev_transaction_block_height"`
	RequiredIters              uint64 `gorm:"type:bigint(20);not null;default:0" json:"required_iters"`
	RewardInfusionNewChallenge string `gorm:"type:varchar(256);not null;default:unknown" json:"reward_infusion_new_challenge"`
	SignagePointIndex          uint64 `gorm:"type:bigint(20);not null;default:0" json:"signage_point_index"`
	SubSlotIters               uint64 `gorm:"type:bigint(20);not null;default:0" json:"sub_slot_iters"`
	BlockTimestamp             uint64 `gorm:"type:int;index:idx_bc_block_timestamp" json:"timestamp"`
	TotalIters                 uint64 `gorm:"type:bigint(20);not null;default:0" json:"total_iters"`
	Weight                     uint64 `gorm:"type:bigint(20);not null;default:0" json:"weight"`
	FarmerAddress              string `gorm:"type:varchar(256);not null;default:unknown;index:idx_bc_farmer_address_itb" json:"farmer_address"`
	PoolAddress                string `gorm:"type:varchar(256);not null;default:unknown;index:idx_bc_pool_address" json:"pool_address"`
	IsTransactionBlock         bool   `gorm:"type:bool;not null;default:false;index:idx_bc_farmer_address_itb" json:"is_transaction_block"`
}

type GetBlocksResponse struct {
	BlockRecords []ChiaBlockRecord `json:"block_records"`
}

func GetBlockRecords(client *http.Client, host string, port uint, start uint64, end uint64, result *GetBlocksResponse) error {
	url := fmt.Sprintf("https://%s:%d/get_block_records?start=%d&end=%d", host, port, start, end)
	data := fmt.Sprintf(`{"start": %d, "end": %d}`, start, end)
	fmt.Printf("reading blocks... start: %d, end: %d \r\n", start, end)
	return RpcFetch(client, url, data, result)
}