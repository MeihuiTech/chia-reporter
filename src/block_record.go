package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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
	contentType := "application/json"
	data := fmt.Sprintf(`{"start": %d, "end": %d}`, start, end)
	fmt.Printf("reading blocks... start: %d, end: %d \r\n", start, end)
	resp, err := client.Post(url, contentType, strings.NewReader(data))
	if err == nil {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error read response: %v", err)
		}
		err = json.Unmarshal(b, result)
		if err != nil {
			return fmt.Errorf("error parsing response: %v", err)
		}
	} else {
		return fmt.Errorf("error on get_block_records: %v", err)
	}
	return nil
}

func RpcClient(certFile string, keyFile string, caFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certs")
	}

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ca file")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true, // Not actually skipping, we check the cert in VerifyPeerCertificate
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// Code copy/pasted and adapted from
			// https://github.com/golang/go/blob/81555cb4f3521b53f9de4ce15f64b77cc9df61b9/src/crypto/tls/handshake_client.go#L327-L344, but adapted to skip the hostname verification.
			// See https://github.com/golang/go/issues/21971#issuecomment-412836078.

			// If this is the first handshake on a connection, process and
			// (optionally) verify the server's certificates.
			certs := make([]*x509.Certificate, len(rawCerts))
			for i, asn1Data := range rawCerts {
				cert, err := x509.ParseCertificate(asn1Data)
				if err != nil {
					return fmt.Errorf("failed to parse certificate from server: %v", err)
				}
				certs[i] = cert
			}

			opts := x509.VerifyOptions{
				Roots:         caCertPool,
				CurrentTime:   time.Now(),
				DNSName:       "", // <- skip hostname verification
				Intermediates: x509.NewCertPool(),
			}

			for i, cert := range certs {
				if i == 0 {
					continue
				}
				opts.Intermediates.AddCert(cert)
			}
			_, err := certs[0].Verify(opts)
			return err
		},
	}
	tlsConfig.BuildNameToCertificate()
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{Transport: tr}
	return client, nil
}

func GetSyncedHeight(db *gorm.DB) (*ChiaBlockSyncHeight, error)  {
	var blockHeight ChiaBlockSyncHeight
	r := db.Last(&blockHeight)
	if r.Error == nil {
		return &blockHeight, nil
	} else if errors.Is(r.Error, gorm.ErrRecordNotFound){
		return nil, nil
	} else {
		return nil, fmt.Errorf("error get latest block: %v", r.Error)
	}
}

func SyncBlocks(ctx context.Context, channel chan int, host string, port uint, privateCert string, privateKey string, caCert string, syncBlocks bool, db *gorm.DB) {

	blockHeight, err := GetSyncedHeight(db)
	if err != nil {
		return
	}
	start := uint64(0)
	if blockHeight != nil {
		start = start + blockHeight.Height + 1
	}

	client, err := RpcClient(privateCert, privateKey, caCert)
	if err == nil {
		batch := uint64(10)
		interval := 20
		for true {
			end := start + batch
			result := &GetBlocksResponse{}
			err = GetBlockRecords(client, host, port, start, end, result)
			if err != nil {
				fmt.Printf("error GetBlockRecords: %v", err)
			} else if len(result.BlockRecords) > 0 {
				// begin Transaction
				err = db.Transaction(func(tx *gorm.DB) error {
					for index, block := range result.BlockRecords{
						farmerAddress, err := EncodePuzzleHash(block.FarmerPuzzleHash, "xch")
						if err == nil {
							err = IncreaseTotalBlock(farmerAddress, tx)
							if err != nil {
								return fmt.Errorf("error increase total blocks")
							}
							err = IncreaseDailyBlock(farmerAddress, block.Height, tx)
							if err != nil {
								return fmt.Errorf("error calculate daily blocks")
							}
							if syncBlocks {
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
					if syncBlocks {
						tx.Create(&result.BlockRecords)
					}
					length := uint64(len(result.BlockRecords))
					err = LogSyncHeight(start + length - 1 ,tx)
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
