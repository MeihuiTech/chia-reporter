package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func ExportAction(ctx *cli.Context) error {
	rChannel := make(chan int)
	signalChannel := make(chan os.Signal)
	defer close(signalChannel)
	defer close(rChannel)

	config, err := NewConfig(ctx)
	if err != nil {
		return err
	}
	if config.WalletRpcPort == 0 {
		return fmt.Errorf("error config: wallet_rpc_port can not be empty")
	}

	go ExportFarmer(context.Background(), rChannel, config)

	signal.Notify(signalChannel, os.Interrupt)
	select {
	case sig := <-signalChannel:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
	case code := <-rChannel:
		fmt.Printf("sync goroutine exit with code: %d\n", code)
	}
	return nil
}

func ExportFarmer(ctx context.Context, channel chan int, config *Config)  {
	client, err := RpcClient(config.PrivateCert, config.PrivateKey, config.CaCert)
	if err == nil {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("shutdown reporter service.")
				return
			case <-time.After(time.Duration(5) * time.Second):
				{
					walletStats, err := GetWalletsStats(client, config.RpcHost, config.WalletRpcPort, config.WalletId)
					if err != nil {
						fmt.Printf("error get wallet stats: %v \r\n", err)
						continue
					}
					fmt.Printf("%v", walletStats)
					plotSize ,err := GetPlotSize(client, config.RpcHost, config.HarvesterRpcPort)
					if err != nil {
						fmt.Printf("error get plot size: %v", err)
						continue
					}
					farmer  := Farmer{
						MinerId: walletStats.Address,
						PuzzleHash: walletStats.PuzzleHash,
						PowerAvailable: plotSize,
						TotalBlockAward: (walletStats.FarmerRewardAmount + walletStats.PoolRewardAmount) / CoinUnit["chia"],
						BalanceMinerAccount: walletStats.Balance,
					}
					fmt.Printf("%v", farmer)
				}
			}
		}

	} else {
		fmt.Printf("error create rcp client: %v \r\n", err)
		channel <- 1
	}
}

func GetPlotSize(client *http.Client, host string, port uint) (uint64, error) {
	var result PlotsResponse
	err := GetPlots(client, host, port, &result)
	if err != nil {
		return 0, err
	}
	fileSize := uint64(0)
	for _, plot := range result.Plots{
		fileSize += plot.FileSize
	}
	return fileSize, nil
}
func GetPlots(client *http.Client, host string, port uint, result *PlotsResponse) error {
	url := fmt.Sprintf("https://%s:%d/get_plots?", host, port)
	data := "{}"
	return RpcFetch(client, url, data, result)
}

func GetWalletsStats(client *http.Client, host string, port uint, walletId uint) (*WalletStats, error)  {
	var walletResponse WalletResponse
	err := GetWallets(client,host, port, &walletResponse)
	if err != nil {
		return nil, err
	}
	balance := float64(0)
	for _, wallet := range walletResponse.Wallets{
		if wallet.ID == walletId {
			var balances Balances
			err = GetWalletBalance(client, host, port, walletId, &balances)
			if err != nil {
				return nil, err
			}

			if wallet.Type == StandWallet {
				balance = balances.SpendableBalance / CoinUnit["colouredcoin"]
			} else {
				balance = balances.SpendableBalance / CoinUnit["chia"]
			}
			break
		}
	}
	var farmedAmount FarmedAmount
	err = GetFarmedAmount(client, host, port, &farmedAmount)
	if err != nil {
		return nil, err
	}
	var walletAddress WalletAddress
	err = GetNextAddress(client, host, port, walletId, &walletAddress)
	if err != nil {
		return nil, err
	}
	_, puzzleHash, err := DecodePuzzleHash("xch1aeyz9l6zau9342knur2kx58cguqj0vydl40f9zuyvkw9k6dnk2vshzwmcc")
	if err != nil {
		return nil, err
	}

	return &WalletStats{
		WalletId: walletId,
		Address: walletAddress.Address,
		Balance: balance,
		PuzzleHash: hex.EncodeToString(puzzleHash),
		FarmerRewardAmount: farmedAmount.FarmerRewardAmount,
		PoolRewardAmount: farmedAmount.PoolRewardAmount,
	} ,nil
}

func GetWallets(client *http.Client, host string, port uint, result *WalletResponse) error {
	url := fmt.Sprintf("https://%s:%d/get_wallets?", host, port)
	data := "{}"
	return RpcFetch(client, url, data, result)
}

func GetWalletBalance(client *http.Client, host string, port uint, walletId uint, result *Balances) error {
	url := fmt.Sprintf("https://%s:%d/get_wallet_balance?", host, port)
	data := fmt.Sprintf(`{"wallet_id": %d}`, walletId)
	return RpcFetch(client, url, data, result)
}

func GetFarmedAmount(client *http.Client, host string, port uint, result *FarmedAmount) error {
	url := fmt.Sprintf("https://%s:%d/get_farmed_amount?", host, port)
	data := "{}"
	return RpcFetch(client, url, data, result)
}

func GetNextAddress(client *http.Client, host string, port uint, walletId uint, result *WalletAddress) error {
	url := fmt.Sprintf("https://%s:%d/get_next_address?", host, port)
	data := fmt.Sprintf(`{"wallet_id": %d, "new_address": %t}`, walletId, false)
	return RpcFetch(client, url, data, result)
}
