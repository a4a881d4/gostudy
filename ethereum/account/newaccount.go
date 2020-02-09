package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"os"
)

//go run ethereum/account/newaccount.go ../ethereum/chain/keystore 931
func main() {
	dir := os.Args[1]
	passphrase := os.Args[2]

	fmt.Println("Add new account to key store: ",dir," passphrase: ",passphrase)

	ks := keystore.NewKeyStore(dir, 262144, 1)
	if a,err := ks.NewAccount(passphrase); err!=nil {
		fmt.Println(err)
	} else {
		fmt.Println(a.Address.Hex())
	}
}
