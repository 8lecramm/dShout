package main

const (
	DAEMON_BLOCK              = "DERO.GetBlock"
	DAEMON_GET_SC             = "DERO.GetSC"
	DAEMON_GET_RANDOM_ADDRESS = "DERO.GetRandomAddress"
	DAEMON_GAS_ESTIMATE       = "DERO.GetGasEstimate"
	DAEMON_NAME_TO_ADDRESS    = "DERO.NameToAddress"
	WALLET_QUERY_KEY          = "QueryKey"
	WALLET_SC_INVOKE          = "scinvoke"
	WALLET_TRANSFER           = "transfer"
)

const (
	DataString DataType = "S"
	DataHash   DataType = "H"
)

type DataType string

type Argument struct {
	Name     string      `json:"name"`
	DataType DataType    `json:"datatype"`
	Value    interface{} `json:"value"`
}

type Arguments []Argument

type (
	JSONRPCRequest struct {
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		ID      string      `json:"id"`
	}
	JSONRPCResponse struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      string      `json:"id"`
	}
)

type (
	GetBlock_Params struct {
		Hash   string `json:"hash,omitempty"`
		Height uint64 `json:"height,omitempty"`
	} // no params
	GetBlock_Result struct {
		Blob         string            `json:"blob"`
		Json         string            `json:"json"`
		Block_Header BlockHeader_Print `json:"block_header"`
		Balance_Tree string            `json:"balance_tree"`
		Meta_Tree    string            `json:"meta_tree"`
		Merkle_Hash  string            `json:"merkle_hash"`
		Status       string            `json:"status"`
	}
)

type BlockHeader_Print struct {
	Depth         int64    `json:"depth"`
	Difficulty    string   `json:"difficulty"`
	Hash          string   `json:"hash"`
	Height        int64    `json:"height"`
	TopoHeight    int64    `json:"topoheight"`
	Major_Version uint64   `json:"major_version"`
	Minor_Version uint64   `json:"minor_version"`
	Nonce         uint64   `json:"nonce"`
	Orphan_Status bool     `json:"orphan_status"`
	SyncBlock     bool     `json:"syncblock"`
	SideBlock     bool     `json:"sideblock"`
	TXCount       int64    `json:"txcount"`
	Miners        []string `json:"miners"`

	Reward    uint64   `json:"reward"`
	Tips      []string `json:"tips"`
	Timestamp uint64   `json:"timestamp"`
}

type (
	NameToAddress_Params struct {
		Name       string `json:"name"`
		TopoHeight int64  `json:"topoheight,omitempty"`
	} // no params
	NameToAddress_Result struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Status  string `json:"status"`
	}
)

type (
	GetSC_Params struct {
		SCID       string   `json:"scid"`
		Code       bool     `json:"code,omitempty"`
		Variables  bool     `json:"variables,omitempty"`
		TopoHeight uint64   `json:"topoheight,omitempty"`
		KeysUint64 []uint64 `json:"keysuint64,omitempty"`
		KeysString []string `json:"keysstring,omitempty"`
		KeysBytes  [][]byte `json:"keysbytes,omitempty"`
	}
	GetSC_Result struct {
		ValuesUint64       []string               `json:"valuesuint64,omitempty"`
		ValuesString       []string               `json:"valuesstring,omitempty"`
		ValuesBytes        []string               `json:"valuesbytes,omitempty"`
		VariableStringKeys map[string]interface{} `json:"stringkeys,omitempty"`
		VariableUint64Keys map[uint64]interface{} `json:"uint64keys,omitempty"`
		Balances           map[string]uint64      `json:"balances,omitempty"`
		Balance            uint64                 `json:"balance"`
		Code               string                 `json:"code"`
		Status             string                 `json:"status"`
	}
)
type (
	Query_Key_Params struct {
		Key_type string `json:"key_type"`
	}
	Query_Key_Result struct {
		Key string `json:"key"`
	}
)

type (
	Transfer struct {
		SCID        string    `json:"scid"`
		Destination string    `json:"destination"`
		Amount      uint64    `json:"amount"`
		Burn        uint64    `json:"burn"`
		Payload_RPC Arguments `json:"payload_rpc"`
	}

	Transfer_Params struct {
		Transfers []Transfer `json:"transfers"`
		SC_Code   string     `json:"sc"`
		SC_Value  uint64     `json:"sc_value"`
		SC_ID     string     `json:"scid"`
		SC_RPC    Arguments  `json:"sc_rpc"`
		Ringsize  uint64     `json:"ringsize"`
		Fees      uint64     `json:"fees"`
		Signer    string     `json:"signer"`
	}
	Transfer_Result struct {
		TXID string `json:"txid,omitempty"`
	}
)

type GasEstimate_Params Transfer_Params
type GasEstimate_Result struct {
	GasCompute uint64 `json:"gascompute"`
	GasStorage uint64 `json:"gasstorage"`
	Status     string `json:"status"`
}

type (
	GetRandomAddress_Params struct {
		SCID string `json:"scid"`
	}

	GetRandomAddress_Result struct {
		Address []string `json:"address"`
		Status  string   `json:"status"`
	}
)

var invoke_params = Arguments{
	Argument{
		Name:     "entrypoint",
		DataType: "S",
		Value:    "Store",
	},
}

func RPC_Request(method string, p any) JSONRPCRequest {
	return JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "0",
		Method:  method,
		Params:  p,
	}
}

func SC_AddMessage(msg string) Argument {
	return Argument{
		Name:     "data",
		DataType: "S",
		Value:    msg,
	}
}
