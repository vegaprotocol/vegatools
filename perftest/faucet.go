package perftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func topUpAsset(faucetURL, pubKey, asset string, amount int) error {
	postBody, _ := json.Marshal(map[string]string{
		"party":  pubKey,
		"amount": fmt.Sprintf("%d", amount),
		"asset":  asset,
	})
	responseBody := bytes.NewBuffer(postBody)

	if !strings.HasPrefix(faucetURL, "http") {
		faucetURL = "http://" + faucetURL
	}

	URL := faucetURL + "/api/v1/mint"

	resp, err := http.Post(URL, "application/json", responseBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	sb := string(body)
	if strings.Contains(sb, "error") {
		return fmt.Errorf(sb)
	}
	return nil
}
