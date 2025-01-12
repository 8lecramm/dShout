package main

import (
	"os"
)

var err error

func main() {

	if err := ReadConfig(); err != nil {
		os.Exit(1)
	}
	//Init_JSON_Clients()

	xswd = XSWD_Init()
	xswd.AppInfo = &AppicationInfo{
		Name:        "dShout",
		Description: "Send messages to one or more users",
		Url:         "https://github.com/8lecramm/dShout",
	}

	if err := xswd.XSWD_Connect(); err != nil {
		log_xswd.Println(err)
		os.Exit(1)
	}

	defer xswd.XSWD_Exit()

	// ask for permission
	if privateKey, err = GetWalletKey(); err != nil {
		log_xswd.Println("No permission for QueryKey")
		os.Exit(1)
	}

	CreateWindow().ShowAndRun()
}
