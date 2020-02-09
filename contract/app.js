var Web3 = require('web3');
var web3 = new Web3(new Web3.providers.IpcProvider("chain/geth.ipc",net));
var eth=web3.eth;

var MyTokenABI = [{"constant":false,"inputs":[{"name":"a","type":"uint256"}],"name":"mul","outputs":[{"name":"b","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]

var tokenContract = new web3.eth.Contract(MyTokenABI, null, {
    from: '0x419db46a3dbddabf0c4b9364891db226735c35a1' 
});

tokenContract.deploy({
    data: MyTokenBin,
    arguments: []
}).send({
    from: '0x419db46a3dbddabf0c4b9364891db226735c35a1',
    gas: 1500000,
    gasPrice: '30000000000000'
}, function(error, transactionHash){
    console.log("deploy tx hash:"+transactionHash)
})
.on('error', function(error){ console.error(error) })
.on('transactionHash', function(transactionHash){ console.log("hash:",transactionHash)})
.on('receipt', function(receipt){
   console.log(receipt.contractAddress) // contains the new contract address
})
.on('confirmation', function(confirmationNumber, receipt){console.log("receipt,",receipt)})
.then(function(newContractInstance){
    console.log(newContractInstance.options.address) // instance with the new contract address
});