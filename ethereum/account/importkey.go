package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"os"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

//go run ethereum/account/importkey.go ../../ethereum/chain/keystore 40dad29726f7e1b56359d2f1cc5a5365cb105b410e1108b3da65c1d97bfe6f8e 931
func main() {
	dir := os.Args[1]
	var privkey = new(big.Int)
	privkey.SetString(os.Args[2],16)
	passphrase := os.Args[3]
	fmt.Println("Import private key: ",privkey.Text(16)," to key store: ",dir," passphrase: ",passphrase)
	
	if pk,err := crypto.ToECDSA(privkey.Bytes()); err!=nil {
		fmt.Println(err)
	} else {
		ks := keystore.NewKeyStore(dir, 262144, 1)
		if a,err := ks.ImportECDSA(pk,passphrase); err!=nil {
			fmt.Println(err)
		} else {
			fmt.Println(a.Address.Hex())
		}
	}
}
