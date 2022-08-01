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

	proto "code.vegaprotocol.io/protos/vega"
)

type WalletWrapper struct {
	walletURL string
}

// UserDetails Holds wallet information for each user
type UserDetails struct {
	userName string
	token    string
	pubKey   string
}

type OrderDetails struct {
	marketID  string
	user      int
	price     int64
	size      int64
	orderType proto.Order_Type
	tif       proto.Order_TimeInForce
	expiresAt int64
}

func (od *OrderDetails) getTIFString() string {
	switch od.tif {
	case proto.Order_TIME_IN_FORCE_GTT:
		return "TIME_IN_FORCE_GTT"
	case proto.Order_TIME_IN_FORCE_GFA:
		return "TIME_IN_FORCE_GFA"
	case proto.Order_TIME_IN_FORCE_GFN:
		return "TIME_IN_FORCE_GFN"
	case proto.Order_TIME_IN_FORCE_GTC:
		return "TIME_IN_FORCE_GTC"
	case proto.Order_TIME_IN_FORCE_IOC:
		return "TIME_IN_FORCE_IOC"
	case proto.Order_TIME_IN_FORCE_FOK:
		return "TIME_IN_FORCE_FOK"
	}
	return ""
}

func Abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

// SecondsFromNowInSecs : Creates a timestamp relative to the current time in seconds
func (w WalletWrapper) SecondsFromNowInSecs(seconds int64) int64 {
	return time.Now().Unix() + seconds
}

func (w WalletWrapper) LoginWallet(username, password string) (string, error) {
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

func (w WalletWrapper) CreateWallet(username, password string) (string, error) {
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

func (w WalletWrapper) CreateKey(password, token string) (string, error) {
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

func (w WalletWrapper) ListKeys(token string) ([]string, error) {
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

func (w *WalletWrapper) CreateOrLoadWallets(number int) error {
	// We want to make or load a set of wallets, do it in a loop here
	var key string
	var newKeys = 0
	for i := 0; i < number; i++ {
		userName := fmt.Sprintf("User%04d", i)

		// Attempt to log in with that username
		token, err := w.LoginWallet(userName, "p3rfb0t")
		if err != nil {
			token, err = w.CreateWallet(userName, "p3rfb0t")
			if err != nil {
				return fmt.Errorf("unable to create a new wallet: %w", err)
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
	return nil
}

func (w *WalletWrapper) SignSubmitTx(user int, command string) error {
	err := w.SendCommand([]byte(command), users[user].token)
	return err
}

func (w *WalletWrapper) SendCommand(submission []byte, token string) error {
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

func (w *WalletWrapper) SendOrder(od OrderDetails) error {
	cmd := `{ "orderSubmission": {
      "marketId": "$MARKETID",
      $PRICE
      "size": "$SIZE",
      "side": "$SIDE",
      "timeInForce": "$TIME_IN_FORCE",
      $EXPIRES_AT
      "type": "$TYPE",
      "reference": "order_ref"
    },
    "pubKey": "$PUBKEY",
    "propagate": true
  }`

	cmd = strings.Replace(cmd, "$MARKETID", od.marketID, 1)
	cmd = strings.Replace(cmd, "$SIZE", fmt.Sprintf("%d", Abs(od.size)), 1)

	if od.orderType == proto.Order_TYPE_LIMIT {
		cmd = strings.Replace(cmd, "$PRICE", fmt.Sprintf("\"price\": \"%d\",", od.price), 1)
		cmd = strings.Replace(cmd, "$TYPE", "TYPE_LIMIT", 1)
	} else {
		cmd = strings.Replace(cmd, "$PRICE", "", 1)
		cmd = strings.Replace(cmd, "$TYPE", "TYPE_MARKET", 1)
	}

	if od.expiresAt > 0 {
		cmd = strings.Replace(cmd, "$EXPIRES_AT", fmt.Sprintf("\"expiresAt\": %d,", od.expiresAt), 1)
	} else {
		cmd = strings.Replace(cmd, "$EXPIRES_AT", "", 1)
	}

	if od.size > 0 {
		cmd = strings.Replace(cmd, "$SIDE", "SIDE_BUY", 1)
	} else {
		cmd = strings.Replace(cmd, "$SIDE", "SIDE_SELL", 1)
	}

	cmd = strings.Replace(cmd, "$TIME_IN_FORCE", od.getTIFString(), 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[od.user].pubKey, 1)

	err := w.SignSubmitTx(od.user, cmd)
	return err
}

func (w *WalletWrapper) SendNewMarketProposal(user int) {
	refStr := "PerfBotProposalRef"

	cmd := `{"proposalSubmission": {
              "reference": "$UNIQUEREF",
               "rationale": {
                "description": "some description"
              },
              "terms": {
                "validationTimestamp": $VALIDATIONTS,
                "closingTimestamp": $CLOSINGTS,
                "enactmentTimestamp": $ENACTMENTTS,
                "newMarket":{
                  "changes":  {
                    "decimalPlaces": 5,
                    "instrument": {
                      "code": "CRYPTO:BTCUSD/NOV22",
                      "name": "NOV 2022 BTC vs USD future",
                      "future": {
                        "settlementAsset": "fUSDC",
                        "quoteName":"BTCUSD",
                        "oracleSpecBinding": {
                          "settlementPriceProperty": "trading.settled",
                          "tradingTerminationProperty": "trading.termination"
                        },
                        "oracleSpecForTradingTermination": {
                          "pubKeys": ["0xDEADBEEF"],
                          "filters": [{
                            "key": {
                              "name": "trading.termination",
                              "type": "TYPE_BOOLEAN"
                            }
                          }]
                        },
                        "oracleSpecForSettlementPrice": {
                          "pubKeys": ["0xDEADBEEF"],
                          "filters": [{
                            "key": {
                              "name": "trading.settled",
                              "type": "TYPE_INTEGER"
                            }
                          }]
                        }
                      }
                    },
                    "simple" : {
                      "factorLong":           0.15,
                      "factorShort":          0.25,
                      "maxMoveUp":            10,
                      "minMoveDown":          -5,
                      "probabilityOfTrading": 0.1
                    }
                  },
                  "liquidityCommitment": {
                    "fee":              "0.01",
                    "commitmentAmount": "50000000",
                    "buys":             [{
                                          "reference" : "PEGGED_REFERENCE_BEST_BID",
                                          "proportion" : 10,
                                          "offset" : "2000"	
                    }],
                    "sells":            [{
                                          "reference" : "PEGGED_REFERENCE_BEST_ASK",
                                          "proportion" : 10,
                                          "offset" : "2000"
                    }]
                  }
                }
              }
            },
            "pubKey": "$PUBKEY",
            "propagate" : true
          }
        }
      }
    }
  }`

	/*
	   "metadata":      [{"base:BTC", "quote:USD", "class:fx/crypto", "monthly", "sector:crypto"}],
	*/

	cmd = strings.Replace(cmd, "$UNIQUEREF", refStr, 1)
	cmd = strings.Replace(cmd, "$VALIDATIONTS", fmt.Sprintf("%d", w.SecondsFromNowInSecs(1)), 1)
	cmd = strings.Replace(cmd, "$CLOSINGTS", fmt.Sprintf("%d", w.SecondsFromNowInSecs(15)), 1)
	cmd = strings.Replace(cmd, "$ENACTMENTTS", fmt.Sprintf("%d", w.SecondsFromNowInSecs(20)), 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)

	err := w.SignSubmitTx(user, cmd)
	if err != nil {
		log.Fatal("failed to submit NewMarketProposal: %w", err)
	}
}

func (w *WalletWrapper) SendCancelAll(user int, marketID string) {
	cmd := `{	"orderCancellation" :{
              "marketId": "$MARKET_ID"
            },
            "pubKey": "$PUBKEY",
            "propagate": true						
          }`

	cmd = strings.Replace(cmd, "$MARKET_ID", marketID, 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)

	err := w.SignSubmitTx(user, cmd)
	if err != nil {
		fmt.Println(cmd)
		log.Fatal("failed to submit Vote Submission: %w", err)
	}
}

func (w *WalletWrapper) SendVote(user int, proposalID string, vote bool) error {
	cmd := `{ "voteSubmission": {
              "proposal_id": "$PROPOSAL_ID",
              "value": "$VOTE"
            },
            "pubKey": "$PUBKEY",
            "propagate" : true
          }`

	cmd = strings.Replace(cmd, "$PROPOSAL_ID", proposalID, 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)
	if vote {
		cmd = strings.Replace(cmd, "$VOTE", "VALUE_YES", 1)
	} else {
		cmd = strings.Replace(cmd, "$VOTE", "VALUE_NO", 1)
	}

	err := w.SignSubmitTx(user, cmd)
	if err != nil {
		return err
	}
	return nil
}
