package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"os"
)

func main() {
	dir := os.argv[1]
	address := os.argv[2]
	passphase := os.argv[3]

	fmt.Println("Form key store: ",dir,"read address: ",address,"private key")

	ks := keystore.NewKeyStore(dir, 262144, 1)
	accounts := ks.Accounts()
	if keyjson, err := ks.Export(accounts[0], "931", "931"); err != nil {
		fmt.Println("Exort",err)
	} else {
		if key,err := keystore.DecryptKey(keyjson, "931"); err != nil {
			fmt.Println("Exort",err)
		} else {
			fmt.Println(key.PrivateKey.D.Text(16))
		}
	}
}
