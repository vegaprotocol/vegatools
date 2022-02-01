package pullvotes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	commandspb "code.vegaprotocol.io/protos/vega/commands/v1"
	"github.com/golang/protobuf/proto"
)

var (
	validators = map[string]string{
		"f3022974212780ea1196af08fd2e8a9c0d784d0be8e97637bd5e763ac4c219bd": "Staking facilities",
		"b861c11eb825d55f835aec898b3caae66a681a354bcb59651d5b3faf02b34844": "Commodum",
		"43697a3e911d8b70c0ce672adde17a5c38ca8f6a0486bf85ed0546e1b9a82887": "Bharvest",
		"4f69b1784656174e89eb094513b7136e88670b42517ed0e48cb6fd3062eb8478": "Nodes Guru",
		"5db9794f44c85b4b259907a00c8ea2383ad688dfef6ffb72c8743b6ae3eaefd4": "Ryabina",
		"126751c5830b50d39eb85412fb2964f46338cce6946ff455b73f1b1be3f5e8cc": "GF1",
		"efbdf943443bd7595e83b0d7e88f37b7932d487d1b94aab3d004997273bb43fc": "Chorus One",
		"5ca98e0dd81143fafea3a3abcefafee73f3886ac97053db8b446593e75c10e9d": "P2P",
		"74023df02b8afc9eaf3e3e2e8b07eab1d2122ac3e74b1b0222daf4af565ad3dd": "XPRV",
		"25794776055552a92e7b27dd8f15563ffb78defe7694d6c4da8bb258daca897c": "Lovali",
		"ac735acc9ab11cf1d8c59c2df2107e00092b4ac96451cb137a1629af5b66242a": "Figment",
		"8d33c6e06207ed5735c8b5b6c0c6234f44eb381b242a25a538ed3315369d2203": "Nala",
		"55504e9bfd914a7bbefa342c82f59a2f4dee344e5b6863a14c02a812f4fbde32": "RBF",
	}
	all map[string]map[string]struct{}
)

const (
	pathBlock = "/block"
)

type result struct {
	Block block `json:"block"`
}

type block struct {
	Data blockData `json:"data"`
}

type blockData struct {
	Txs [][]byte `json:"txs"`
}

type blockResponse struct {
	Result result `json:"result"`
}

func getTxsAtBlockHeight(nodeAddress string, height uint64) {
	// prepare the request
	req, err := http.NewRequest("GET", nodeAddress, nil)
	if err != nil {
		panic(err)
	}
	values := req.URL.Query()
	values.Add("height", fmt.Sprintf("%v", height))
	req.URL.RawQuery = values.Encode()
	req.URL.Path = path.Join(req.URL.Path, pathBlock)

	// build client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// extract the response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	blockResp := blockResponse{}
	err = json.Unmarshal(body, &blockResp)
	if err != nil {
		panic(err)
	}

	for _, v := range blockResp.Result.Block.Data.Txs {
		unpackSignedTx(v)
	}
}

func getCommand(inputData *commandspb.InputData) {
	switch cmd := inputData.Command.(type) {
	case *commandspb.InputData_NodeVote:
		m, ok := all[cmd.NodeVote.Reference]
		if !ok {
			m = map[string]struct{}{}
			all[cmd.NodeVote.Reference] = m

		}
		m[hex.EncodeToString(cmd.NodeVote.PubKey)] = struct{}{}
		all[cmd.NodeVote.Reference] = m

		fmt.Printf("%v -> %d\n", cmd.NodeVote.Reference, len(m))
		for k := range m {
			fmt.Printf("%v -> %v\n", k, validators[k])
		}
	}
}

func unpackSignedTx(rawtx []byte) {
	tx := commandspb.Transaction{}
	err := proto.Unmarshal(rawtx, &tx)
	if err != nil {
		panic(err)
	}

	inputData := commandspb.InputData{}
	err = proto.Unmarshal(tx.InputData, &inputData)
	if err != nil {
		panic(err)
	}

	getCommand(&inputData)

}

// Start ...
func Start(from, to uint64, address string) error {
	all = map[string]map[string]struct{}{}
	for i := from; i < to; i++ {
		fmt.Printf("block: %v\n", i)
		getTxsAtBlockHeight(address, i)
	}
	return nil
}
