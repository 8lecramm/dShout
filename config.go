package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	SCID      string `json:"scid"`
	RateLimit uint64 `json:"limiter"`
}
type SCData struct {
	Height     uint64
	Prev       uint64
	Msg        string
	LastUpdate uint64
}
type MsgDecryped struct {
	Message string
	Block   uint64
	Time    string
}
type Limiter struct {
	Init  time.Time
	Count uint64
}

var log_xswd = log.New(os.Stdout, "dShout > ", log.Ldate|log.Ltime)

var SC_Config Config
var SC_Data SCData
var lastCheck uint64

var decrypted_messages []MsgDecryped
var tx_fees = map[uint64]uint64{
	2:   40,
	4:   60,
	8:   60,
	16:  80,
	32:  100,
	64:  120,
	128: 180,
}

var rateLimit Limiter

func ReadConfig() error {

	data, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &SC_Config); err != nil {
		return err
	}

	return nil
}

func Parse_SC(r GetSC_Result) error {

	if !SC_SanityCheck(r) {
		return fmt.Errorf("SC sanity check failed")
	}

	SC_UpdateData(r)

	return nil
}

func SC_SanityCheck(r GetSC_Result) bool {

	if len(r.ValuesString) != 3 {
		return false
	}

	_, err := strconv.ParseUint(r.ValuesString[0], 10, 64)
	if err != nil {
		return false
	}
	_, err = strconv.ParseUint(r.ValuesString[1], 10, 64)
	if err != nil {
		return false
	}
	if r.ValuesString[2] == "" {
		return false
	}

	return true
}

func SC_UpdateData(r GetSC_Result) {
	SC_Data.Height, _ = strconv.ParseUint(r.ValuesString[0], 10, 64)
	SC_Data.Prev, _ = strconv.ParseUint(r.ValuesString[1], 10, 64)
	SC_Data.Msg = r.ValuesString[2]
}

func SC_Build_GetSC_Request(height uint64) GetSC_Params {
	return GetSC_Params{
		SCID:       SC_Config.SCID,
		TopoHeight: height,
		KeysString: []string{"height", "prev", "msg"},
	}
}

func (l *Limiter) Check() bool {
	if time.Since(l.Init) > time.Second {
		l.Init = time.Now()
		l.Count = 1

		return true
	}
	l.Count++
	if l.Count >= SC_Config.RateLimit {
		return false
	}

	return true
}
