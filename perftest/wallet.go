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

type UserDetails struct {
	userName string
	token    string
	pubKey   string
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

// SecondsFromNowInSecs : Creates a timestamp relative to the current time in seconds
func SecondsFromNowInSecs(seconds int64) int64 {
	return time.Now().Unix() + seconds
}

func loginWallet(walletURL, userName, password string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"wallet":     userName,
		"passphrase": password,
	})
	responseBody := bytes.NewBuffer(postBody)

	URL := "http://" + walletURL + "/api/v1/auth/token"

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

func createWallet(walletURL, userName, password string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"wallet":     userName,
		"passphrase": password,
	})
	responseBody := bytes.NewBuffer(postBody)

	URL := "http://" + walletURL + "/api/v1/wallets"

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

func createKey(walletURL, userName, password, token string) (string, error) {
	postBody, _ := json.Marshal(map[string]string{
		"passphrase": password,
	})
	postBuffer := bytes.NewBuffer(postBody)

	URL := "http://" + walletURL + "/api/v1/keys"

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

func listKeys(walletURL, token string) ([]string, error) {
	URL := "http://" + walletURL + "/api/v1/keys"

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

func createOrLoadWallets(walletURL string, number int) (int, error) {
	// We want to make or load a set of wallets, do it in a loop here
	var key string
	var newKeys = 0
	for i := 0; i < number; i++ {
		userName := fmt.Sprintf("User%04d", i)

		// Attempt to log in with that username
		token, err := loginWallet(walletURL, userName, "p3rfb0t")
		if err != nil {
			token, err = createWallet(walletURL, userName, "p3rfb0t")
			if err != nil {
				return 0, fmt.Errorf("unable to create a new wallet: %w", err)
			}
		}
		keys, _ := listKeys(walletURL, token)
		if len(keys) == 0 {
			key, _ = createKey(walletURL, userName, "p3rfb0t", token)
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
	return newKeys, nil
}

func signSubmitTx(user int, command string) error {
	err := sendCommand([]byte(command), users[user].token)
	return err
}

func sendCommand(submission []byte, token string) error {
	postBuffer := bytes.NewBuffer(submission)

	URL := "http://" + savedWalletURL + "/api/v1/command"

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

func sendOrder(marketId string, user int, price, size int64,
	orderType string, tif proto.Order_TimeInForce, expiresAt int64) {
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

	cmd = strings.Replace(cmd, "$MARKETID", marketId, 1)
	cmd = strings.Replace(cmd, "$SIZE", fmt.Sprintf("%d", abs(size)), 1)

	if orderType == "LIMIT" {
		cmd = strings.Replace(cmd, "$PRICE", fmt.Sprintf("\"price\": \"%d\",", price), 1)
		cmd = strings.Replace(cmd, "$TYPE", "TYPE_LIMIT", 1)
	} else {
		cmd = strings.Replace(cmd, "$PRICE", "", 1)
		cmd = strings.Replace(cmd, "$TYPE", "TYPE_MARKET", 1)
	}

	if expiresAt > 0 {
		cmd = strings.Replace(cmd, "$EXPIRES_AT", fmt.Sprintf("\"expiresAt\": %d,", expiresAt), 1)
	} else {
		cmd = strings.Replace(cmd, "$EXPIRES_AT", "", 1)
	}

	if size > 0 {
		cmd = strings.Replace(cmd, "$SIDE", "SIDE_BUY", 1)
	} else {
		cmd = strings.Replace(cmd, "$SIDE", "SIDE_SELL", 1)
	}

	switch tif {
	case proto.Order_TIME_IN_FORCE_GTT:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_GTT", 1)
	case proto.Order_TIME_IN_FORCE_GFA:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_GFA", 1)
	case proto.Order_TIME_IN_FORCE_GFN:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_GFN", 1)
	case proto.Order_TIME_IN_FORCE_GTC:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_GTC", 1)
	case proto.Order_TIME_IN_FORCE_IOC:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_IOC", 1)
	case proto.Order_TIME_IN_FORCE_FOK:
		cmd = strings.Replace(cmd, "$TIME_IN_FORCE", "TIME_IN_FORCE_FOK", 1)
	}
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)
	cmd = strings.Replace(cmd, "$EXPIRESAT", fmt.Sprintf("%d", expiresAt), 1)

	/*	if orderType == LIMIT {
	  cmd.OrderSubmission.Price = strconv.FormatInt(price, 10)
	}*/

	err := signSubmitTx(user, cmd)
	if err != nil {
		fmt.Println("failed to submit Order Submission: %w", err)
	}
}

func sendNewMarketProposal(user int) {
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
	cmd = strings.Replace(cmd, "$VALIDATIONTS", fmt.Sprintf("%d", SecondsFromNowInSecs(1)), 1)
	cmd = strings.Replace(cmd, "$CLOSINGTS", fmt.Sprintf("%d", SecondsFromNowInSecs(10)), 1)
	cmd = strings.Replace(cmd, "$ENACTMENTTS", fmt.Sprintf("%d", SecondsFromNowInSecs(15)), 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)

	err := signSubmitTx(user, cmd)
	if err != nil {
		log.Fatal("failed to submit NewMarketProposal: %w", err)
	}
}

func sendCancelAll(user int, marketId string) {
	cmd := `{	"orderCancellation" :{
              "marketId": "$MARKET_ID"
            },
            "pubKey": "$PUBKEY",
            "propagate": true						
          }`

	cmd = strings.Replace(cmd, "$MARKET_ID", marketId, 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)

	err := signSubmitTx(user, cmd)
	if err != nil {
		fmt.Println(cmd)
		log.Fatal("failed to submit Vote Submission: %w", err)
	}
}
