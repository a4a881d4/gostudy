package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"os"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

//go run ethereum/account/Somekey.go ../ethereum/chain/keystore 4000000000000000000000000000000000000000000000000000000000000000 931
func main() {
	dir := os.Args[1]
	var privkey = new(big.Int)
	privkey.SetString(os.Args[2],16)
	passphrase := os.Args[3]
	fmt.Println("Import private key: ",privkey.Text(16)," to key store: ",dir," passphrase: ",passphrase)

	for i:=0; i < 16; i++ {
		if pk,err := crypto.ToECDSA(privkey.Bytes()); err!=nil {
			fmt.Println(err)
		} else {
			ks := keystore.NewKeyStore(dir, 262144, 1)
			if a,err := ks.ImportECDSA(pk,passphrase); err!=nil {
				fmt.Println(err)
			} else {
				fmt.Println(a.Address.Hex(),privkey.Text(16))
			}
		}
		privkey = privkey.Add(privkey,big.NewInt(1))
	}
}
