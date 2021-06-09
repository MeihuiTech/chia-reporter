package main

import (
	"errors"
	"gorm.io/gorm"
)

type ChiaTotalFarmerBlocks struct {
	ID            uint64 `gorm:"primaryKey;<-:false" json:"id"`
	FarmerAddress string `gorm:"type:varchar(256);not null;index:idx_tfb_farmer_address" json:"farmer_address"`
	BlockCount    uint64 `gorm:"type:bigint(20);not null;"`
}

type ChiaDailyFarmerBlocks struct {
	ID            uint64 `gorm:"primaryKey;<-:false" json:"id"`
	FarmerAddress string `gorm:"type:varchar(256);not null;index:idx_dfb_farmer_address" json:"farmer_address"`
	BlockCount    uint64 `gorm:"type:bigint(20);not null;"`
	Day           string `gorm:"type:date;not null;index:idx_dfb_day"`
}

type ChiaBlockSyncHeight struct {
	ID     uint64 `gorm:"primaryKey;<-:false" json:"id"`
	Height uint64 `gorm:"type:bigint(20);not null;default:0" json:"height"`
}

func IncreaseTotalBlock(farmerAddress string, db *gorm.DB) error {
	var totalBlock ChiaTotalFarmerBlocks
	r := db.Where("farmer_address = ?", farmerAddress).Take(&totalBlock)
	if r.Error == nil {
		r = db.Model(totalBlock).Update("block_count", gorm.Expr("block_count + ?", 1))
		return r.Error
	} else if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		r = db.Create(&ChiaTotalFarmerBlocks{
			BlockCount: uint64(1),
			FarmerAddress: farmerAddress,
		})
		return r.Error
	} else {
		return r.Error
	}
}

func IncreaseDailyBlock(farmerAddress string, height uint64, db *gorm.DB) error {
	//TODO height to timestamp?
	return nil
}

func LogSyncHeight(height uint64, db *gorm.DB) error {
	blockHeight, err := GetSyncedHeight(db)
	if err != nil {
		return err
	}
	if blockHeight == nil {
		r := db.Create(&ChiaBlockSyncHeight{
			Height: height,
		})
		return r.Error
	} else {
		r := db.Model(blockHeight).Update("height", height)
		return r.Error
	}
}
