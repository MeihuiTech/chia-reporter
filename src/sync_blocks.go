package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"time"
)

func GetSyncedHeight(db *gorm.DB) (*ChiaBlockSyncHeight, error) {
	var blockHeight ChiaBlockSyncHeight
	r := db.Last(&blockHeight)
	if r.Error == nil {
		return &blockHeight, nil
	} else if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return nil, fmt.Errorf("error get latest block: %v", r.Error)
	}
}

func SyncBlocks(ctx context.Context, channel chan int, config *Config) {

	db, err := GetDb(config)
	if err != nil {
		fmt.Printf("error open db connection: %v \r\n", err)
		channel <- 1
		return
	}

	blockHeight, err := GetSyncedHeight(db)
	if err != nil {
		return
	}
	start := uint64(0)
	if blockHeight != nil {
		start = start + blockHeight.Height + 1
	}

	client, err := RpcClient(config.PrivateCert, config.PrivateKey, config.CaCert)
	if err == nil {
		batch := uint64(10)
		interval := 20
		for true {
			end := start + batch
			result := &GetBlocksResponse{}
			err = GetBlockRecords(client, config.RpcHost, config.FullNodeRpcPort, start, end, result)
			if err != nil {
				fmt.Printf("error GetBlockRecords: %v", err)
			} else if len(result.BlockRecords) > 0 {
				doneWithHistory := uint64(len(result.BlockRecords)) < batch
				// begin Transaction
				err = db.Transaction(func(tx *gorm.DB) error {
					for index, block := range result.BlockRecords {
						farmerAddress, err := EncodePuzzleHash(block.FarmerPuzzleHash, "xch")
						if err == nil {
							err = IncreaseTotalBlock(farmerAddress, tx)
							if err != nil {
								return fmt.Errorf("error increase total blocks")
							}

							if doneWithHistory && block.BlockTimestamp == 0 {
								result.BlockRecords[index].BlockTimestamp = uint64(time.Now().Unix())
							} else if block.BlockTimestamp == 0 {
								result.BlockRecords[index].BlockTimestamp = HeightToTimestamp(block.Height)
							}

							timestamp := result.BlockRecords[index].BlockTimestamp
							err = IncreaseDailyBlock(farmerAddress, timestamp, tx)
							if err != nil {
								return fmt.Errorf("error calculate daily blocks")
							}
							if config.SyncBlocks {
								result.BlockRecords[index].FarmerAddress = farmerAddress
								poolAddress, err := EncodePuzzleHash(block.FarmerPuzzleHash, "xch")
								if err == nil {
									result.BlockRecords[index].PoolAddress = poolAddress
								} else {
									fmt.Printf("error encode pool puzzle hash:%v \r\n", err)
								}
								result.BlockRecords[index].IsTransactionBlock = block.BlockTimestamp != 0
							}
						} else {
							return fmt.Errorf("error encode farmer puzzle hash:%v \r\n", err)
						}
					}
					if config.SyncBlocks {
						tx.Create(&result.BlockRecords)
					}
					length := uint64(len(result.BlockRecords))
					err = LogSyncHeight(start+length-1, tx)
					if err == nil {
						start += length
					} else {
						return fmt.Errorf("error log sync height: %v", err)
					}
					return nil
				})
				if len(result.BlockRecords) < 10 {
					time.Sleep(time.Duration(interval) * time.Second)
				}

			} else {
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}
	} else {
		fmt.Printf("error create rcp client: %v \r\n", err)
		channel <- 1
	}
}
func SyncAction(ctx *cli.Context) error {

	syncChannel := make(chan int)
	signalChannel := make(chan os.Signal)
	defer close(signalChannel)
	defer close(syncChannel)

	config, err := NewConfig(ctx)
	if err != nil {
		return err
	}

	go SyncBlocks(context.Background(), syncChannel, config)

	signal.Notify(signalChannel, os.Interrupt)
	select {
	case sig := <-signalChannel:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
	case code := <-syncChannel:
		fmt.Printf("sync goroutine exit with code: %d\n", code)
	}
	return nil
}
