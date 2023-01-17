package perftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	proto "code.vegaprotocol.io/vega/protos/vega"
	commandspb "code.vegaprotocol.io/vega/protos/vega/commands/v1"
	walletpb "code.vegaprotocol.io/vega/protos/vega/wallet/v1"
)

// WalletWrapper holds details about the wallet
type walletWrapper struct {
	walletURL string
}

// UserDetails Holds wallet information for each user
type UserDetails struct {
	userName string
	token    string
	pubKey   string
}

// SecondsFromNowInSecs : Creates a timestamp relative to the current time in seconds
func (w walletWrapper) SecondsFromNowInSecs(seconds int64) int64 {
	return time.Now().Unix() + seconds
}

type keys struct {
	Keys []key
}
type key struct {
	Name      string
	PublicKey string
}
type listKeysResult struct {
	Jsonrpc string
	Result  keys
	ID      string
}

func (w walletWrapper) sendTransaction(user UserDetails, subType string, subData interface{}) ([]byte, error) {
	transaction, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "client.send_transaction",
		"id":      "1",
		"params": map[string]interface{}{
			"publicKey":   user.pubKey,
			"sendingMode": "TYPE_SYNC",
			"transaction": map[string]interface{}{
				subType: subData,
			},
		},
	})

	return w.sendRequest(transaction, user.token)
}

func (w walletWrapper) sendRequest(request []byte, token string) ([]byte, error) {
	postBody := bytes.NewBuffer(request)

	URL := "http://" + w.walletURL + "/api/v2/requests"

	req, err := http.NewRequest(http.MethodPost, URL, postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("origin", "perfbot")
	req.Header.Add("Authorization", fmt.Sprintf("VWT %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	reply, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (w walletWrapper) NewMarket(offset int, user UserDetails) error {
	newMarket := map[string]interface{}{
		"rationale": map[string]interface{}{
			"description": "desc",
			"title":       "title",
		},
		"terms": map[string]interface{}{
			"closingTimestamp":   w.SecondsFromNowInSecs(15),
			"enactmentTimestamp": w.SecondsFromNowInSecs(30),
			"newMarket": map[string]interface{}{
				"changes": map[string]interface{}{
					"lpPriceRange":          "10",
					"decimalPlaces":         "5",
					"positionDecimalPlaces": "5",
					"instrument": map[string]interface{}{
						"code": "CRYPTO:BTCUSD/NOV22",
						"name": "NOV 2022 BTC vs USD future",
						"future": map[string]interface{}{
							"settlementAsset": "fUSDC",
							"quoteName":       "BTCUSD",
							"dataSourceSpecForSettlementData": map[string]interface{}{
								"external": map[string]interface{}{
									"oracle": map[string]interface{}{
										"signers": []interface{}{
											map[string]interface{}{
												"ethAddress": map[string]interface{}{
													"address": "0xfCEAdAFab14d46e20144F48824d0C09B1a03F2BC",
												},
											},
										},
										"filters": []interface{}{
											map[string]interface{}{
												"key": map[string]interface{}{
													"name": "trading.settled",
													"type": "TYPE_INTEGER",
												},
												"conditions": []interface{}{
													map[string]interface{}{
														"operator": "OPERATOR_GREATER_THAN",
														"value":    "0",
													},
												},
											},
										},
									},
								},
							},
							"dataSourceSpecForTradingTermination": map[string]interface{}{
								"external": map[string]interface{}{
									"oracle": map[string]interface{}{
										"signers": []interface{}{
											map[string]interface{}{
												"ethAddress": map[string]interface{}{
													"address": "0xfCEAdAFab14d46e20144F48824d0C09B1a03F2BC",
												},
											},
										},
										"filters": []interface{}{
											map[string]interface{}{
												"key": map[string]interface{}{
													"name": "trading.terminated",
													"type": "TYPE_BOOLEAN",
												},
											},
										},
									},
								},
							},
							"dataSourceSpecBinding": map[string]interface{}{
								"settlementDataProperty":     "trading.settled",
								"tradingTerminationProperty": "trading.terminated",
							},
						},
					},
					"simple": map[string]interface{}{
						"factorLong":           "0.15",
						"factorShort":          "0.25",
						"maxMoveUp":            "10",
						"minMoveDown":          "-5",
						"probabilityOfTrading": "0.1",
					},
				},
			},
		},
	}

	_, err := w.sendTransaction(user, "proposalSubmission", newMarket)
	return err
}

// GetFirstKey gives us the first public key linked to our wallet
func (w walletWrapper) GetFirstKey(longLivedToken string) (string, error) {
	post, _ := json.Marshal(map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "client.list_keys",
	})
	body, err := w.sendRequest(post, longLivedToken)

	var values listKeysResult = listKeysResult{}
	err = json.Unmarshal(body, &values)
	if err != nil {
		fmt.Println(err)
	}

	if len(values.Result.Keys) > 0 {
		return values.Result.Keys[0].PublicKey, nil
	}
	return "", nil
}

// SendBatchOrders sends a set of new order commands to the wallet
func (w *walletWrapper) SendBatchOrders(user UserDetails,
	cancels []*commandspb.OrderCancellation,
	amends []*commandspb.OrderAmendment,
	orders []*commandspb.OrderSubmission) error {

	command := &walletpb.SubmitTransactionRequest_BatchMarketInstructions{
		BatchMarketInstructions: &commandspb.BatchMarketInstructions{},
	}
	command.BatchMarketInstructions.Cancellations = cancels
	command.BatchMarketInstructions.Amendments = amends
	command.BatchMarketInstructions.Submissions = orders

	_, err := w.sendTransaction(user, "orderSubmission", command)
	return err
}

// SendOrder sends a new order command to the wallet
func (w *walletWrapper) SendOrder(user UserDetails, os *commandspb.OrderSubmission) error {
	_, err := w.sendTransaction(user, "orderSubmission", os)
	return err
}

func (w *walletWrapper) SendLiquidityProvision(user UserDetails, marketID string, orderCount int) error {
	lp := commandspb.LiquidityProvisionSubmission{
		MarketId:         marketID,
		Reference:        "MarketLiquidity",
		Fee:              "0.01",
		CommitmentAmount: "1000000000",
	}

	// Generate the buy and sell side LP orders
	buys := make([]*proto.LiquidityOrder, orderCount)
	sells := make([]*proto.LiquidityOrder, orderCount)
	for i := 0; i < orderCount; i++ {
		buys[i] = &proto.LiquidityOrder{
			Reference:  proto.PeggedReference_PEGGED_REFERENCE_BEST_BID,
			Proportion: 10,
			Offset:     strconv.FormatInt(int64(1000+(i*10)), 10),
		}
		sells[i] = &proto.LiquidityOrder{
			Reference:  proto.PeggedReference_PEGGED_REFERENCE_BEST_ASK,
			Proportion: 10,
			Offset:     strconv.FormatInt(int64(1000+(i*10)), 10),
		}
	}

	lp.Buys = buys
	lp.Sells = sells

	_, err := w.sendTransaction(user, "liquidityProvisionSubmission", &lp)

	return err
}

// SendCancelAll will build and send a cancel all command to the wallet
func (w *walletWrapper) SendCancelAll(user UserDetails, marketID string) error {
	cancel := commandspb.OrderCancellation{
		MarketId: marketID,
	}

	_, err := w.sendTransaction(user, "orderCancellation", &cancel)
	if err != nil {
		return err
	}
	return err
}

// SendVote will build and send a vote command to the wallet
func (w walletWrapper) SendVote(user UserDetails, propID string) error {
	vote := commandspb.VoteSubmission{
		ProposalId: propID,
		Value:      proto.Vote_VALUE_YES,
	}

	_, err := w.sendTransaction(user, "voteSubmission", &vote)
	return err
}
