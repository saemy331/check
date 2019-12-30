package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Contract struct {

}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Contract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}


func (t *Contract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "save" {
		result, err = save(stub, args)
	} else if fn =="get" { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	} else{
		 err = nil
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func save(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}
	key := args[0]
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return key, nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	key := args[0]
	value, err := stub.GetState(key)
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", key)
	}
	return string(value), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(Contract)); err != nil {
		fmt.Printf("Error starting Contract chaincode: %s", err)
	}
}
