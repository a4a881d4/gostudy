package main
import (
	"fmt"
	_ "net"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"

	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"
)

func main() {
	// client, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: `\\.\pipe\geth.ipc`, Net: "unix"})
	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println(client)
	// }
	var connection = web3.NewWeb3(providers.NewIPCProvider("../ethereum/chain/geth.ipc"))
	ext := web3ext.NewWeb3Ext(connection.Provider)
	if ver, err := ext.DBGetString("DatabaseVersion"); err == nil {
		fmt.Println("Version:",ver)
	} else {
		fmt.Println(err)
	}
}