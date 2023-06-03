package main

import (
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
)

func main() {
	w, err := wallet.GenerateNewWallet("a")
	if err != nil {
		log.Printf("Can not create wallet. Error %v", err)
	}
	err = w.Store()
	if err != nil {
		log.Println(err)
		return
	}
}
