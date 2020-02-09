package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"fmt"
	"os"
)

//go run ethereum/account/readkeyfile.go ../ethereum/chain/keystore 9c6898f5f38a015d74601dd297852791ac42f40a 931
func main() {
	dir := os.Args[1]
	var address = common.HexToAddress(os.Args[2])
	// address.SetString(os.Args[2],16)
	passphrase := os.Args[3]

	fmt.Println("Form key store: ",dir,"read address: ",address.Hex(),"private key")

	ks := keystore.NewKeyStore(dir, 262144, 1)
	accounts := ks.Accounts()
	for i,a := range accounts {
		if address.Hex() == a.Address.Hex() {
			if keyjson, err := ks.Export(accounts[i], passphrase, passphrase); err != nil {
				fmt.Println("Exort",err)
			} else {
				if key,err := keystore.DecryptKey(keyjson, passphrase); err != nil {
					fmt.Println("Exort",err)
				} else {
					fmt.Println(key.PrivateKey.D.Text(16))
				}
			}
		} else {
			// fmt.Println(address.Sub(a.Address.Big(), address))
			fmt.Println(i)
		}

	}
}
