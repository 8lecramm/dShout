package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi/mnemonics"
)

var new_topoheight_callback = func(value any) {

	var height int64 = int64(value.(float64))
	log.Println(height)
}

func SC_Request(height uint64) error {

	req := RPC_Request(DAEMON_GET_SC, SC_Build_GetSC_Request(height))
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if !xswd.xswd_send(data) {
		return fmt.Errorf("error sending request")
	}

	var r GetSC_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return err
	}

	if err = Parse_SC(r); err != nil {
		return err
	}

	return nil
}

func GetTimestamp(height uint64) (ts string, err error) {

	req := RPC_Request(DAEMON_BLOCK, GetBlock_Params{Height: height})
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if !xswd.xswd_send(data) {
		return "", fmt.Errorf("error sending request")
	}

	var r GetBlock_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return "", err
	}

	timestamp := time.UnixMilli(int64(r.Block_Header.Timestamp)).Format(time.DateTime)

	return timestamp, nil
}

func SC_SendMessage(msg string, ringsize string) (txid string, err error) {

	var t Transfer_Params
	p := invoke_params

	p = append(p,
		SC_AddMessage(msg),
		Argument{
			Name:     "SC_ACTION",
			DataType: "U",
			Value:    0,
		},
		Argument{
			Name:     "SC_ID",
			DataType: DataHash,
			Value:    SC_Config.SCID,
		})

	t.SC_RPC = p
	t.Ringsize, err = strconv.ParseUint(ringsize, 10, 64)
	if err != nil {
		return "", err
	}

	t.Transfers = append(t.Transfers, BuildTransfer())
	if t.Transfers == nil {
		return "", fmt.Errorf("empty transfer")
	}

	req := RPC_Request(DAEMON_GAS_ESTIMATE, t)
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	if !xswd.xswd_send(data) {
		return "", fmt.Errorf("error sending request")
	}
	log_xswd.Println(">", DAEMON_GAS_ESTIMATE)

	var r GasEstimate_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return "", err
	}

	t.Fees = r.GasStorage + tx_fees[t.Ringsize]

	req = RPC_Request(WALLET_TRANSFER, t)
	data, err = json.Marshal(req)
	if err != nil {
		return "", err
	}

	if !xswd.xswd_send(data) {
		return "", fmt.Errorf("error sending request")
	}
	log_xswd.Println(">", WALLET_TRANSFER)

	var result Transfer_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return "", err
	}

	return result.TXID, nil
}

func SC_SyncLoop() (int, error) {

	if err := SC_Request(0); err != nil {
		log_xswd.Println(err)
		return 0, err
	}

	current_height := SC_Data.Height

	if SC_Data.Height == SC_Data.LastUpdate {
		return 0, nil
	}

	var msg_count int
	for {
		plain, err := hex.DecodeString(SC_Data.Msg)
		if err != nil {
			continue
		}
		contents := DecryptMessages(string(plain))

		if len(contents) > 0 {

			ts, err := GetTimestamp(SC_Data.Height)
			if err != nil {
				ts = "#no timestamp"
			}
			for _, m := range contents {
				if m == "" {
					continue
				}
				decrypted_messages = append(decrypted_messages, MsgDecryped{
					Message: m,
					Block:   SC_Data.Height,
					Time:    ts,
				})
				msg_count++
			}
		}

		time.Sleep(50 * time.Millisecond)
		if err := SC_Request(SC_Data.Prev); err != nil {
			return msg_count, err
		}

		if SC_Data.Height == SC_Data.Prev || SC_Data.Height == SC_Data.LastUpdate {
			SC_Data.LastUpdate = current_height
			break
		}
	}

	return msg_count, nil
}

func GetWalletKey() (key *big.Int, err error) {

	req := RPC_Request(WALLET_QUERY_KEY, Query_Key_Params{
		Key_type: "mnemonic",
	})
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if !xswd.xswd_send(data) {
		return nil, fmt.Errorf("error sending request")
	}

	var r Query_Key_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return nil, err
	}

	_, key, err = mnemonics.Words_To_Key(r.Key)

	return key, err
}

func RPC_GetRandomAddress() string {

	req := RPC_Request(DAEMON_GET_RANDOM_ADDRESS, GetRandomAddress_Params{
		SCID: ZEROHASH,
	})
	data, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	if !xswd.xswd_send(data) {
		return ""
	}

	var r GetRandomAddress_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return ""
	}

	if len(r.Address) > 0 {
		return r.Address[0]
	}

	return ""
}

func RPC_NameToAddress(name string) string {

	req := RPC_Request(DAEMON_NAME_TO_ADDRESS, NameToAddress_Params{
		Name:       name,
		TopoHeight: -1,
	})
	data, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	if !xswd.xswd_send(data) {
		return ""
	}

	var r NameToAddress_Result
	if err = xswd_response(<-xswd.response, &r); err != nil {
		return ""
	}

	return r.Address
}

func ValidateReceivers(r []string) (result []string) {

	for _, a := range r {
		_, err := globals.ParseValidateAddress(a)
		if err != nil {
			if w := RPC_NameToAddress(a); w != "" {
				result = append(result, w)
			}
		} else {
			result = append(result, a)
		}
	}

	return
}

func BuildTransfer() (t Transfer) {

	if t.Destination = RPC_GetRandomAddress(); t.Destination == "" {
		return t
	}

	t.Amount = 0
	t.SCID = ZEROHASH

	return t
}
