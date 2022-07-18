package perftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type UserDetails struct {
	userName string
	token    string
	pubKey   string
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
		return nil, err
	}
	sb := string(body)
	log.Println(sb)
	if strings.Contains(sb, "error") {
		return nil, fmt.Errorf(sb)
	}

	// Get the token value out from the JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(sb), &result)
	keys := result["keys"].([]interface{})
	if len(keys) == 0 {
		return nil, fmt.Errorf("No keys found")
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
				return 0, fmt.Errorf("Unable to create a new wallet: %w", err)
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

func sendCommand(walletURL string, submission []byte, token string) error {
	postBuffer := bytes.NewBuffer(submission)

	URL := "http://" + walletURL + "/api/v1/command"

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
