/*
 *    Copyright 2020 bitfly gmbh
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package rpc

import (
	"coda-explorer/types"
	"coda-explorer/util"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var logger = logrus.New().WithField("module", "rpc")

// CodaClient encapsulates all methods required to communicate with a coda blockchain node via the graphql api
type CodaClient struct {
	httpClient *http.Client
	host       string
}

// NewCodaClient creates a new rpc client
func NewCodaClient(host string) *CodaClient {

	cc := &CodaClient{
		httpClient: &http.Client{Timeout: time.Second * 10},
		host:       host,
	}

	return cc
}

// Helper function for executing a graphql query and parsing its result
func (cc *CodaClient) getData(query string, target interface{}) error {
	req := "http://" + cc.host + "?query=" + url.QueryEscape(query)

	resp, err := cc.httpClient.Get(req)
	if err != nil {
		return fmt.Errorf("error retrieving graphql query response: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading graphql query response body: %w", err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return fmt.Errorf("error decoding graphql query response: %w", err)
	}

	return nil
}

// WatchNewBlocks subscription for watching for new blocks, uses the websocket protocol and will automatically reconnect on disconnects
func (cc *CodaClient) WatchNewBlocks(newBlockChan chan string) {
	for {
		time.Sleep(time.Second * 5)
		graphqlWsHeader := http.Header{}
		graphqlWsHeader["Sec-WebSocket-Protocol"] = []string{"graphql-ws"}

		c, _, err := websocket.DefaultDialer.Dial("ws://"+cc.host, graphqlWsHeader)
		if err != nil {
			logger.Errorf("error connecting to websocket at %v: %w", cc.host, err)
			continue
		}

		subscriptionName := "newBlock"

		queryMessage := fmt.Sprintf(`{
	  "id": "1",
	  "type": "start",
	  "payload": {
		"variables": {},
		"extensions": {},
		"operationName": null,
		"query": "subscription { %s { stateHash } }"
	  }
	}`, subscriptionName)

		err = c.WriteMessage(websocket.TextMessage, []byte(queryMessage))
		if err != nil {
			logger.Errorf("error subscribing to newBlock events: %w", err)
			c.Close()
			continue
		}
		logger.Infof("subscribed to %s events", subscriptionName)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Errorf("error reading from websocket subscription: %w:", err)
				break
			}
			logger.Infof("received %v via websocket subscriptions", string(message))

			var parsedNotification *newBlockNotification
			err = json.Unmarshal(message, parsedNotification)
			if err != nil {
				logger.Errorf("error parsing new block notifciation: %w:", err)
				break
			}
			newBlockChan <- parsedNotification.StateHash
		}
		c.Close()
	}
}

type newBlockNotification struct {
	StateHash string `json:"stateHash"`
}

// GetLastBlocks retrieves the last <lookback> block hashes
func (cc *CodaClient) GetLastBlocks(lookback int) ([]*types.Block, error) {

	query := `{
				blocks(last: ` + strconv.Itoa(lookback) + `) {
					nodes {
						stateHash
						protocolState {
							previousStateHash
							consensusState {
							blockchainLength
							epoch
							slot
							totalCurrency
							}
							blockchainState {
							snarkedLedgerHash
							stagedLedgerHash
							date
							}
						}
						transactions {
							coinbase
							feeTransfer {
								fee
								recipient
							}
							userCommands {
								amount
								fee
								from
								id
								isDelegation
								memo
								nonce
								to
							}
						}
						snarkJobs {
							fee
							prover
							workIds
						}
						creatorAccount {
							publicKey
						}
					}
				}
			}`

	var resp lastBlockHashesResponse
	err := cc.getData(query, &resp)
	if err != nil {
		return nil, fmt.Errorf("error executing last block hashes graphql query: %w", err)
	}

	blocks := make([]*types.Block, len(resp.Data.Blocks.Nodes))

	for i, block := range resp.Data.Blocks.Nodes {
		b := &types.Block{}

		b.StateHash = block.StateHash
		b.Epoch = util.MustParseInt(block.ProtocolState.ConsensusState.Epoch)
		b.Height = util.MustParseInt(block.ProtocolState.ConsensusState.BlockchainLength)
		b.Slot = util.MustParseInt(block.ProtocolState.ConsensusState.Slot)
		b.PreviousStateHash = block.ProtocolState.PreviousStateHash

		blocks[i] = b
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Height < blocks[j].Height
	})
	return blocks, nil
}

// Type for parsing the last block hashes graphql query response
type lastBlockHashesResponse struct {
	Data struct {
		Blocks struct {
			Nodes []struct {
				StateHash     string `json:"stateHash"`
				ProtocolState struct {
					BlockchainState struct {
						Date              string `json:"date"`
						SnarkedLedgerHash string `json:"snarkedLedgerHash"`
						StagedLedgerHash  string `json:"stagedLedgerHash"`
					} `json:"blockchainState"`
					ConsensusState struct {
						BlockchainLength string `json:"blockchainLength"`
						Epoch            string `json:"epoch"`
						Slot             string `json:"slot"`
						TotalCurrency    string `json:"totalCurrency"`
					} `json:"consensusState"`
					PreviousStateHash string `json:"previousStateHash"`
				} `json:"protocolState"`
				SnarkJobs []struct {
					Fee     string  `json:"fee"`
					Prover  string  `json:"prover"`
					WorkIds []int64 `json:"workIds"`
				} `json:"snarkJobs"`
				Transactions struct {
					Coinbase    string `json:"coinbase"`
					FeeTransfer []struct {
						Fee       string `json:"fee"`
						Recipient string `json:"recipient"`
					} `json:"feeTransfer"`
					UserCommands []struct {
						Amount       string `json:"amount"`
						Fee          string `json:"fee"`
						From         string `json:"from"`
						ID           string `json:"id"`
						IsDelegation bool   `json:"isDelegation"`
						Memo         string `json:"memo"`
						Nonce        int    `json:"nonce"`
						To           string `json:"to"`
					} `json:"userCommands"`
				} `json:"transactions"`
				CreatorAccount struct {
					PublicKey string `json:"publicKey"`
				} `json:"creatorAccount"`
			} `json:"nodes"`
		} `json:"blocks"`
	} `json:"data"`
}

// GetBlock retrieves a block by its state hash
func (cc *CodaClient) GetBlock(stateHash string) (*types.Block, error) {
	query := `
		query {
		  block(stateHash: "` + stateHash + `") {
			stateHash
			protocolState {
			  previousStateHash
			  consensusState {
				blockchainLength
				epoch
				slot
				totalCurrency
			  }
			  blockchainState {
				snarkedLedgerHash
				stagedLedgerHash
				date
			  }
			}
			transactions {
			  coinbase
			  feeTransfer {
				fee
				recipient
			  }
			  userCommands {
				amount
				fee
				from
				id
				isDelegation
				memo
				nonce
				to
			  }
			}
			snarkJobs {
			  fee
			  prover
			  workIds
			}
			creatorAccount {
			  publicKey
			}
		  }
		}
	`

	var resp getBlockResponse
	err := cc.getData(query, &resp)
	if err != nil {
		return nil, fmt.Errorf("error executing get block graphql query for block %v: %w", stateHash, err)
	}

	block := &types.Block{
		StateHash:         stateHash,
		PreviousStateHash: resp.Data.Block.ProtocolState.PreviousStateHash,
		SnarkedLedgerHash: resp.Data.Block.ProtocolState.BlockchainState.SnarkedLedgerHash,
		StagedLedgerHash:  resp.Data.Block.ProtocolState.BlockchainState.StagedLedgerHash,
		Coinbase:          util.MustParseInt(resp.Data.Block.Transactions.Coinbase),
		Creator:           resp.Data.Block.CreatorAccount.PublicKey,
		Slot:              util.MustParseInt(resp.Data.Block.ProtocolState.ConsensusState.Slot),
		Height:            util.MustParseInt(resp.Data.Block.ProtocolState.ConsensusState.BlockchainLength),
		Epoch:             util.MustParseInt(resp.Data.Block.ProtocolState.ConsensusState.Epoch),
		Ts:                util.MustParseJsTimestamp(resp.Data.Block.ProtocolState.BlockchainState.Date),
		TotalCurrency:     util.MustParseInt(resp.Data.Block.ProtocolState.ConsensusState.TotalCurrency),
		UserCommandsCount: len(resp.Data.Block.Transactions.UserCommands),
		SnarkJobsCount:    len(resp.Data.Block.SnarkJobs),
		FeeTransferCount:  len(resp.Data.Block.Transactions.FeeTransfer),
		UserJobs:          make([]*types.UserJob, len(resp.Data.Block.Transactions.UserCommands)),
		SnarkJobs:         make([]*types.SnarkJob, len(resp.Data.Block.SnarkJobs)),
		FeeTransfers:      make([]*types.FeeTransfer, len(resp.Data.Block.Transactions.FeeTransfer)),
	}

	for i, job := range resp.Data.Block.Transactions.UserCommands {
		block.UserJobs[i] = &types.UserJob{
			BlockStateHash: stateHash,
			Index:          i,
			ID:             job.ID,
			Sender:         job.From,
			Recipient:      job.To,
			Memo:           job.Memo,
			Fee:            util.MustParseInt(job.Fee),
			Amount:         util.MustParseInt(job.Amount),
			Nonce:          job.Nonce,
			Delegation:     job.IsDelegation,
		}
	}

	for i, sj := range resp.Data.Block.SnarkJobs {
		block.SnarkJobs[i] = &types.SnarkJob{
			BlockStateHash: stateHash,
			Index:          i,
			Jobids:         sj.WorkIds,
			Prover:         sj.Prover,
			Fee:            util.MustParseInt(sj.Fee),
		}
	}

	for i, ft := range resp.Data.Block.Transactions.FeeTransfer {
		block.FeeTransfers[i] = &types.FeeTransfer{
			BlockStateHash: stateHash,
			Index:          i,
			Recipient:      ft.Recipient,
			Fee:            util.MustParseInt(ft.Fee),
		}
	}

	return block, nil
}

// Type for parsing the get block graphql query response
type getBlockResponse struct {
	Data struct {
		Block struct {
			StateHash     string `json:"stateHash"`
			ProtocolState struct {
				BlockchainState struct {
					Date              string `json:"date"`
					SnarkedLedgerHash string `json:"snarkedLedgerHash"`
					StagedLedgerHash  string `json:"stagedLedgerHash"`
				} `json:"blockchainState"`
				ConsensusState struct {
					BlockchainLength string `json:"blockchainLength"`
					Epoch            string `json:"epoch"`
					Slot             string `json:"slot"`
					TotalCurrency    string `json:"totalCurrency"`
				} `json:"consensusState"`
				PreviousStateHash string `json:"previousStateHash"`
			} `json:"protocolState"`
			SnarkJobs []struct {
				Fee     string  `json:"fee"`
				Prover  string  `json:"prover"`
				WorkIds []int64 `json:"workIds"`
			} `json:"snarkJobs"`
			Transactions struct {
				Coinbase    string `json:"coinbase"`
				FeeTransfer []struct {
					Fee       string `json:"fee"`
					Recipient string `json:"recipient"`
				} `json:"feeTransfer"`
				UserCommands []struct {
					Amount       string `json:"amount"`
					Fee          string `json:"fee"`
					From         string `json:"from"`
					ID           string `json:"id"`
					IsDelegation bool   `json:"isDelegation"`
					Memo         string `json:"memo"`
					Nonce        int    `json:"nonce"`
					To           string `json:"to"`
				} `json:"userCommands"`
			} `json:"transactions"`
			CreatorAccount struct {
				PublicKey string `json:"publicKey"`
			} `json:"creatorAccount"`
		} `json:"block"`
	} `json:"data"`
}

// GetAccount retrieves account information by the account public key
func (cc *CodaClient) GetAccount(publicKey string) (*types.Account, error) {

	query := `query {
  				account(publicKey: "` + publicKey + `") {
					balance {
					  total
					}
					nonce
					receiptChainHash
					delegateAccount {
					  publicKey
					}
					votingFor
				  }
				}`

	var resp getAccountResponse
	err := cc.getData(query, &resp)
	if err != nil {
		return nil, fmt.Errorf("error executing get account graphql query: %w", err)
	}

	account := &types.Account{
		PublicKey:        publicKey,
		Balance:          util.MustParseInt(resp.Data.Account.Balance.Total),
		Nonce:            util.MustParseInt(resp.Data.Account.Nonce),
		ReceiptChainHash: resp.Data.Account.ReceiptChainHash,
		Delegate:         resp.Data.Account.DelegateAccount.PublicKey,
		VotingFor:        resp.Data.Account.VotingFor,
		TxSent:           0,
		TxReceived:       0,
		BlocksProposed:   0,
		SnarkJobs:        0,
		FirstSeen:        time.Time{},
		LastSeen:         time.Time{},
	}
	return account, nil
}

// Type for parsing the account information graphql query response
type getAccountResponse struct {
	Data struct {
		Account struct {
			Balance struct {
				Total string `json:"total"`
			} `json:"balance"`
			DelegateAccount struct {
				PublicKey string `json:"publicKey"`
			} `json:"delegateAccount"`
			Nonce            string `json:"nonce"`
			ReceiptChainHash string `json:"receiptChainHash"`
			VotingFor        string `json:"votingFor"`
		} `json:"account"`
	} `json:"data"`
}

// GetDaemonStatus retrieves the current daemon status
func (cc *CodaClient) GetDaemonStatus() (*types.DaemonStatus, error) {
	query := `query {
			  daemonStatus {
				blockchainLength
				commitId
				consensusConfiguration {
				  epochDuration
				  slotDuration
				  slotsPerEpoch
				}
				consensusMechanism
				highestBlockLengthReceived
				ledgerMerkleRoot
				numAccounts
				peers
				stateHash
				syncStatus
				uptimeSecs
			  }
			}
			`

	var resp getDaemonStatusResponse
	err := cc.getData(query, &resp)
	if err != nil {
		return nil, fmt.Errorf("error executing get daemon status graphql query: %w", err)
	}

	daemonStatus := &types.DaemonStatus{
		Ts:                         time.Now(),
		BlockchainLength:           resp.Data.DaemonStatus.BlockchainLength,
		CommitID:                   resp.Data.DaemonStatus.CommitID,
		EpochDuration:              resp.Data.DaemonStatus.ConsensusConfiguration.EpochDuration,
		SlotDuration:               resp.Data.DaemonStatus.ConsensusConfiguration.SlotDuration,
		SlotsPerEpoch:              resp.Data.DaemonStatus.ConsensusConfiguration.SlotsPerEpoch,
		ConsensusMechanism:         resp.Data.DaemonStatus.ConsensusMechanism,
		HighestBlockLengthReceived: resp.Data.DaemonStatus.HighestBlockLengthReceived,
		LedgerMerkleRoot:           resp.Data.DaemonStatus.LedgerMerkleRoot,
		NumAccounts:                resp.Data.DaemonStatus.NumAccounts,
		Peers:                      resp.Data.DaemonStatus.Peers,
		PeersCount:                 len(resp.Data.DaemonStatus.Peers),
		StateHash:                  resp.Data.DaemonStatus.StateHash,
		SyncStatus:                 resp.Data.DaemonStatus.SyncStatus,
		Uptime:                     resp.Data.DaemonStatus.UptimeSecs,
	}
	return daemonStatus, nil
}

// Type for parsing the daemon status graphql query response
type getDaemonStatusResponse struct {
	Data struct {
		DaemonStatus struct {
			BlockchainLength       int    `json:"blockchainLength"`
			CommitID               string `json:"commitId"`
			ConsensusConfiguration struct {
				EpochDuration int `json:"epochDuration"`
				SlotDuration  int `json:"slotDuration"`
				SlotsPerEpoch int `json:"slotsPerEpoch"`
			} `json:"consensusConfiguration"`
			ConsensusMechanism         string   `json:"consensusMechanism"`
			HighestBlockLengthReceived int      `json:"highestBlockLengthReceived"`
			LedgerMerkleRoot           string   `json:"ledgerMerkleRoot"`
			NumAccounts                int      `json:"numAccounts"`
			Peers                      []string `json:"peers"`
			StateHash                  string   `json:"stateHash"`
			SyncStatus                 string   `json:"syncStatus"`
			UptimeSecs                 int      `json:"uptimeSecs"`
		} `json:"daemonStatus"`
	} `json:"data"`
}
