package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"os"
	"math/big"
)

//go run ethereum/account/readkeyfile.go ../../ethereum/chain/keystore 3A703C956f29Da6666d681fd143170f9a84D20db 931
func main() {
	dir := os.Args[1]
	var address = new(big.Int)
	address.SetString(os.Args[2],16)
	passphrase := os.Args[3]

	fmt.Println("Form key store: ",dir,"read address: ",address.Text(16),"private key")

	ks := keystore.NewKeyStore(dir, 262144, 1)
	accounts := ks.Accounts()
	for i,a := range accounts {
		if address.Sub(a.Address.Big(), address).Sign() == 0 {
			if keyjson, err := ks.Export(accounts[i], passphrase, passphrase); err != nil {
				fmt.Println("Exort",err)
			} else {
				if key,err := keystore.DecryptKey(keyjson, passphrase); err != nil {
					fmt.Println("Exort",err)
				} else {
					fmt.Println(key.PrivateKey.D.Text(16))
				}
			}
		}
	}
}
