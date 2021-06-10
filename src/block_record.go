package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
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
