package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const rpcUser = "bootcamp"
const rpcPass = "bootcamp123"
const rpcURL = "http://127.0.0.1:18443"

var httpClient = &http.Client{}

// RPCRequest is what we SEND to Bitcoin Core
type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

// RPCResponse is what Bitcoin Core SENDS BACK
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID any `json:"id"`
}


// ============================================================================
// SECTION 3: CHALLENGE 1 - BLOCKCHAIN INFO
// ============================================================================

// BlockchainInfo struct matches the JSON response from getblockchaininfo
type BlockchainInfo struct {
	Chain      string  `json:"chain"`
	Blocks     int     `json:"blocks"`
	Difficulty float64 `json:"difficulty"`
}

func showBlockchainInfo() error {
	fmt.Println("\nBLOCKCHAIN INFO")

	var info BlockchainInfo

	// RPC call: getblockchaininfo with no params, no wallet
	if err := rpc("getblockchaininfo", nil, "", &info); err != nil {
		return err
	}

	fmt.Println("\n=== Blockchain Info ===")
	fmt.Printf("Chain: %s\n", info.Chain)
	fmt.Printf("Blocks: %d\n", info.Blocks)
	fmt.Printf("Difficulty: %v\n", info.Difficulty)

	return nil
}

// ============================================================================
// SECTION 4: CHALLENGE 2 - WALLET BALANCE
// ============================================================================

func showWalletBalance(wallet string) error {
	fmt.Println("\nWALLET BALANCE")

	// Try to load the wallet (ignore error if already loaded)
	_ = rpc("loadwallet", []any{wallet}, "", nil)

	var balance float64

	// RPC call: getbalance with wallet parameter
	if err := rpc("getbalance", nil, wallet, &balance); err != nil {
		return err
	}

	fmt.Println("\n=== Wallet: " + wallet + " ===")
	fmt.Printf("Balance: %.8f BTC\n", balance)

	return nil
}

// ============================================================================
// SECTION 5: CHALLENGE 3 - LIST TRANSACTIONS
// ============================================================================

type Transaction struct {
	Category      string  `json:"category"`
	Amount        float64 `json:"amount"`
	TxID          string  `json:"txid"`
	Confirmations int     `json:"confirmations"`
}

func listTransactions(wallet string, count int) error {
	fmt.Println("\nLIST TRANSACTIONS")

	// Load wallet
	_ = rpc("loadwallet", []any{wallet}, "", nil)

	var txs []Transaction

	// RPC call: listtransactions with parameters
	if err := rpc("listtransactions", []any{"*", count}, wallet, &txs); err != nil {
		return err
	}

	fmt.Printf("\n=== Recent Transactions (%s) ===\n", wallet)
	fmt.Printf("(Showing last %d transactions)\n\n", count)

	for i, tx := range txs {
		direction := "OUT"

		// Check category to determine direction
		switch tx.Category {
		case "receive", "generate", "immature":
			direction = "IN "
		}

		// Truncate txid to first 20 chars
		txidShort := tx.TxID
		if len(txidShort) > 20 {
			txidShort = txidShort[:20]
		}

		fmt.Printf("[%d] %s %+.8f BTC - %s... [%d confs]\n",
			i+1, direction, tx.Amount, txidShort, tx.Confirmations)
	}

	return nil
}

// ============================================================================
// SECTION 6: CHALLENGE 4 - DECODE TRANSACTION
// ============================================================================

type Vin struct {
	Coinbase string `json:"coinbase"`
	TxID     string `json:"txid"`
	Vout     int    `json:"vout"`
}

type ScriptPubKey struct {
	Address string `json:"address"`
}

type Vout struct {
	Value        float64      `json:"value"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type RawTransaction struct {
	Size int    `json:"size"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}

func decodeTransaction(txid string) error {
	fmt.Println("DECODE TRANSACTION")

	var tx RawTransaction

	// RPC call: getrawtransaction with verbose=true
	if err := rpc("getrawtransaction", []any{txid, true}, "", &tx); err != nil {
		return err
	}

	// Truncate txid for display
	txidDisplay := txid
	if len(txidDisplay) > 20 {
		txidDisplay = txidDisplay[:20]
	}

	fmt.Printf("\n--- Transaction %s... ---\n", txidDisplay)
	fmt.Printf("Size: %d bytes\n", tx.Size)

	// Print inputs
	fmt.Println("\nInputs:")
	for i, vin := range tx.Vin {
		if vin.Coinbase != "" {
			fmt.Printf(" [%d] COINBASE (mining reward)\n", i)
		} else {
			// Truncate previous txid
			prevTxid := vin.TxID
			if len(prevTxid) > 20 {
				prevTxid = prevTxid[:20]
			}
			fmt.Printf(" [%d] From: %s... (output #%d)\n", i, prevTxid, vin.Vout)
		}
	}

	// Print outputs
	fmt.Println("\nOutputs:")
	for i, vout := range tx.Vout {
		address := vout.ScriptPubKey.Address
		if address == "" {
			address = "(no address)"
		}
		fmt.Printf(" [%d] To: %s (%.8f BTC)\n", i, address, vout.Value)
	}

	return nil
}

// ============================================================================
// SECTION 7: CHALLENGE 5 - BLOCK DETAILS
// ============================================================================

type Block struct {
	Height int      `json:"height"`
	Hash   string   `json:"hash"`
	Time   int64    `json:"time"`
	NTx    int      `json:"nTx"`
	Tx     []string `json:"tx"`
}

func showBlock(blockhash string) error {
	fmt.Println("\nBLOCK DETAILS")

	// If no hash provided, get the best block
	if blockhash == "" {
		if err := rpc("getbestblockhash", nil, "", &blockhash); err != nil {
			return err
		}
		fmt.Println("(Using latest block)")
	}

	var block Block

	// RPC call: getblock with verbose=1
	if err := rpc("getblock", []any{blockhash, 1}, "", &block); err != nil {
		return err
	}

	// Truncate hash for display
	hashDisplay := block.Hash
	if len(hashDisplay) > 32 {
		hashDisplay = hashDisplay[:32]
	}

	fmt.Printf("\n=== Block #%d ===\n", block.Height)
	fmt.Printf("Hash: %s...\n", hashDisplay)
	fmt.Printf("Time: %d\n", block.Time)
	fmt.Printf("Transactions: %d\n", block.NTx)

	return nil
}

// ============================================================================
// SECTION 8: RPC HELPER FUNCTION (THE MAGIC!)
// ============================================================================

// rpc makes a JSON-RPC call to Bitcoin Core
func rpc(method string, params []any, wallet string, out any) error {

	// STEP 1: Build the URL
	url := rpcURL
	if wallet != "" {
		url = fmt.Sprintf("%s/wallet/%s", rpcURL, wallet)
	}

	// STEP 2: Create the RPC request struct
	req := RPCRequest{
		JSONRPC: "1.0",
		ID:      "explorer",
		Method:  method,
		Params:  params,
	}

	// STEP 3: Convert request struct to JSON bytes
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// STEP 4: Create HTTP POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// STEP 5: Add authentication (Basic Auth)
	httpReq.SetBasicAuth(rpcUser, rpcPass)
	httpReq.Header.Set("Content-Type", "application/json")

	// STEP 6: Send the request to Bitcoin Core
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// STEP 7: Read the response body
	respBody, err := io.ReadAll(httpResp.Body)
    if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
    }

    fmt.Printf("\nHTTP Status: %s\n", httpResp.Status)
    fmt.Printf("Response Body: %q\n", string(respBody))

	// STEP 8: Parse the JSON response
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// STEP 9: Check for RPC errors
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error (%d): %s",
			rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// STEP 10: Decode the result into the output parameter
	if out != nil {
		if err := json.Unmarshal(rpcResp.Result, out); err != nil {
			return fmt.Errorf("failed to decode result: %w", err)
		}
	}

	return nil
}

// ============================================================================
// SECTION 9: MAIN FUNCTION (RUN ALL CHALLENGES)
// ============================================================================

func main() {


	// ===== CHALLENGE 1 =====
	if err := showBlockchainInfo(); err != nil {
		fmt.Printf("❌ Error in showing block chain info: %v\n", err)
	}

	// ===== CHALLENGE 2 =====
	if err := showWalletBalance("alice"); err != nil {
		fmt.Printf("❌ Error in in showing wallet balance: %v\n", err)
	}

	// ===== CHALLENGE 3 =====
	if err := listTransactions("alice", 5); err != nil {
		fmt.Printf("❌ Error in listing transactions: %v\n", err)
	}

	// ===== CHALLENGE 4 =====
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Getting a transaction to decode...")

	var bestBlockHash string
	if err := rpc("getbestblockhash", nil, "", &bestBlockHash); err != nil {
		fmt.Printf("Error getting best block hash: %v\n", err)
	} else {
		var block Block
		if err := rpc("getblock", []any{bestBlockHash, 1}, "", &block); err != nil {
			fmt.Printf("Error getting block: %v\n", err)
		} else if len(block.Tx) > 0 {
			if err := decodeTransaction(block.Tx[0]); err != nil {
				fmt.Printf("❌ Error in decoding the transaction: %v\n", err)
			}
		}
	}

	// ===== CHALLENGE 5 =====
	if err := showBlock(""); err != nil {
		fmt.Printf("❌ Error in showing block: %v\n", err)
	}

}
