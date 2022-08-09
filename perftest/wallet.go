package perftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"

	proto "code.vegaprotocol.io/protos/vega"
	commandspb "code.vegaprotocol.io/protos/vega/commands/v1"
	v1 "code.vegaprotocol.io/protos/vega/oracles/v1"
	walletpb "code.vegaprotocol.io/protos/vega/wallet/v1"
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

// LoginWallet opens a wallet and logs into it using the supplied username and password
func (w walletWrapper) LoginWallet(username, password string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"wallet":     username,
		"passphrase": password,
	})
	responseBody := bytes.NewBuffer(postBody)

	URL := "http://" + w.walletURL + "/api/v1/auth/token"

	resp, err := http.Post(URL, "application/json", responseBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	if strings.Contains(sb, "error") {
		return "", fmt.Errorf(sb)
	}

	// Get the token value out from the JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(sb), &result)

	return result["token"].(string), nil
}

// CreateWallet will create a new wallet if one does not already exist
func (w walletWrapper) CreateWallet(username, password string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"wallet":     username,
		"passphrase": password,
	})
	responseBody := bytes.NewBuffer(postBody)

	URL := "http://" + w.walletURL + "/api/v1/wallets"

	resp, err := http.Post(URL, "application/json", responseBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	if strings.Contains(sb, "\"error\"") {
		return "", fmt.Errorf(sb)
	}

	// Get the token value out from the JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(sb), &result)

	return result["token"].(string), nil
}

// CreateKey will create a new key pair in the open wallet
func (w walletWrapper) CreateKey(password, token string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"passphrase": password,
	})
	postBuffer := bytes.NewBuffer(postBody)

	URL := "http://" + w.walletURL + "/api/v1/keys"

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, URL, postBuffer)
	if err != nil {
		log.Fatal(err)
	}

	bearer := "Bearer " + token

	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	if strings.Contains(sb, "error") {
		return "", fmt.Errorf(sb)
	}

	// Get the token value out from the JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(sb), &result)
	keys := result["key"].(map[string]interface{})
	return keys["pub"].(string), nil
}

// ListKeys shows all the keys associated with the open wallet.
func (w walletWrapper) ListKeys(token string) ([]string, error) {
	URL := "http://" + w.walletURL + "/api/v1/keys"

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	bearer := "Bearer " + token

	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sb := string(body)
	if strings.Contains(sb, "error") {
		return nil, fmt.Errorf(sb)
	}

	// Get the token value out from the JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(sb), &result)
	keys := result["keys"].([]interface{})
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys found")
	}

	pubKeys := []string{}
	for i := 0; i < len(keys); i++ {
		key := keys[i].(map[string]interface{})
		pubKeys = append(pubKeys, key["pub"].(string))
	}
	return pubKeys, nil
}

// CreateOrLoadWallets will first attempt to open a wallet but if that does not
// exist it will create one and create a key ready for use
func (w walletWrapper) CreateOrLoadWallets(number int) ([]UserDetails, error) {
	// We want to make or load a set of wallets, do it in a loop here
	var key string
	var newKeys = 0
	var users []UserDetails = []UserDetails{}
	for i := 0; i < number; i++ {
		userName := fmt.Sprintf("User%04d", i)

		// Attempt to log in with that username
		token, err := w.LoginWallet(userName, "p3rfb0t")
		if err != nil {
			token, err = w.CreateWallet(userName, "p3rfb0t")
			if err != nil {
				return nil, fmt.Errorf("unable to create a new wallet: %w", err)
			}
		}
		keys, _ := w.ListKeys(token)
		if len(keys) == 0 {
			key, _ = w.CreateKey("p3rfb0t", token)
			newKeys++
		} else {
			key = keys[0]
		}

		users = append(users, UserDetails{
			userName: userName,
			token:    token,
			pubKey:   key,
		})
	}
	return users, nil
}

// SignSubmitTx will sign and then submit a transaction
func (w *walletWrapper) SignSubmitTx(userToken string, command string) error {
	err := w.SendCommand([]byte(command), userToken)
	return err
}

// SendCommand will send a signed command to the wallet
func (w *walletWrapper) SendCommand(submission []byte, token string) error {
	postBuffer := bytes.NewBuffer(submission)

	URL := "http://" + w.walletURL + "/api/v1/command"

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, URL, postBuffer)
	if err != nil {
		return err
	}

	bearer := "Bearer " + token
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	sb := string(body)
	if strings.Contains(sb, "error") {
		log.Println(sb)
		return fmt.Errorf(sb)
	}
	return nil
}

// SendOrder sends a new order command to the wallet
func (w *walletWrapper) SendOrder(user UserDetails, os *commandspb.OrderSubmission) error {
	m := jsonpb.Marshaler{}

	submitTxReq := &walletpb.SubmitTransactionRequest{
		PubKey:    user.pubKey,
		Propagate: true,
		Command: &walletpb.SubmitTransactionRequest_OrderSubmission{
			OrderSubmission: os,
		},
	}

	cmd, err := m.MarshalToString(submitTxReq)
	if err != nil {
		return err
	}
	return w.SignSubmitTx(user.token, cmd)
}

// SendNewMarketProposal will build and send a new market proposal to the wallet
func (w *walletWrapper) SendNewMarketProposal(user UserDetails) error {

	m := jsonpb.Marshaler{}

	submitTxReq := &walletpb.SubmitTransactionRequest{
		PubKey:    user.pubKey,
		Propagate: true,
		Command: &walletpb.SubmitTransactionRequest_ProposalSubmission{
			ProposalSubmission: &commandspb.ProposalSubmission{
				Reference: "PerfBotProposalRef",
				Rationale: &proto.ProposalRationale{
					Description: "PerfBotRational",
				},
				Terms: &proto.ProposalTerms{
					ValidationTimestamp: w.SecondsFromNowInSecs(1),
					ClosingTimestamp:    w.SecondsFromNowInSecs(15),
					EnactmentTimestamp:  w.SecondsFromNowInSecs(20),
					Change: &proto.ProposalTerms_NewMarket{
						NewMarket: &proto.NewMarket{
							Changes: &proto.NewMarketConfiguration{
								DecimalPlaces: 5,
								Instrument: &proto.InstrumentConfiguration{
									Code: "CRYPTO:BTCUSD/NOV22",
									Name: "NOV 2022 BTC vs USD future",
									Product: &proto.InstrumentConfiguration_Future{
										Future: &proto.FutureProduct{
											SettlementAsset: "fUSDC",
											QuoteName:       "BTCUSD",
											OracleSpecBinding: &proto.OracleSpecToFutureBinding{
												SettlementPriceProperty:    "trading.settled",
												TradingTerminationProperty: "trading.termination",
											},
											OracleSpecForTradingTermination: &v1.OracleSpecConfiguration{
												PubKeys: []string{"0xDEADBEEF"},
												Filters: []*v1.Filter{
													&v1.Filter{Key: &v1.PropertyKey{
														Name: "trading.termination",
														Type: v1.PropertyKey_TYPE_BOOLEAN,
													}},
												},
											},
											OracleSpecForSettlementPrice: &v1.OracleSpecConfiguration{
												PubKeys: []string{"0xDEADBEEF"},
												Filters: []*v1.Filter{
													&v1.Filter{Key: &v1.PropertyKey{
														Name: "trading.settled",
														Type: v1.PropertyKey_TYPE_INTEGER,
													}},
												},
											},
										},
									},
								},
								RiskParameters: &proto.NewMarketConfiguration_Simple{
									Simple: &proto.SimpleModelParams{
										FactorLong:           0.15,
										FactorShort:          0.25,
										MaxMoveUp:            10,
										MinMoveDown:          -5,
										ProbabilityOfTrading: 0.1,
									},
								},
							},
							LiquidityCommitment: &proto.NewMarketCommitment{
								Fee:              "0.01",
								CommitmentAmount: "50000000",
								Buys: []*proto.LiquidityOrder{
									&proto.LiquidityOrder{
										Reference:  proto.PeggedReference_PEGGED_REFERENCE_BEST_BID,
										Proportion: 10,
										Offset:     "2000",
									},
								},
								Sells: []*proto.LiquidityOrder{
									&proto.LiquidityOrder{
										Reference:  proto.PeggedReference_PEGGED_REFERENCE_BEST_ASK,
										Proportion: 10,
										Offset:     "2000",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	cmd, err := m.MarshalToString(submitTxReq)
	if err != nil {
		return err
	}

	return w.SignSubmitTx(user.token, cmd)
}

// SendCancelAll will build and send a cancel all command to the wallet
func (w *walletWrapper) SendCancelAll(user UserDetails, marketID string) error {
	cancel := commandspb.OrderCancellation{
		MarketId: marketID,
	}

	m := jsonpb.Marshaler{}

	submitTxReq := &walletpb.SubmitTransactionRequest{
		PubKey:    user.pubKey,
		Propagate: true,
		Command: &walletpb.SubmitTransactionRequest_OrderCancellation{
			OrderCancellation: &cancel,
		},
	}
	cmd, err := m.MarshalToString(submitTxReq)
	if err != nil {
		return err
	}
	return w.SignSubmitTx(user.token, cmd)
}

// SendVote will build and send a vote command to the wallet
func (w walletWrapper) SendVote(user UserDetails, vote *commandspb.VoteSubmission) error {
	m := jsonpb.Marshaler{}

	submitTxReq := &walletpb.SubmitTransactionRequest{
		PubKey:    user.pubKey,
		Propagate: true,
		Command: &walletpb.SubmitTransactionRequest_VoteSubmission{
			VoteSubmission: vote,
		},
	}
	cmd, err := m.MarshalToString(submitTxReq)

	err = w.SignSubmitTx(user.token, cmd)
	if err != nil {
		return err
	}
	return nil
}
