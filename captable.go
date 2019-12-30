package main

import (
        "fmt"
        "github.com/hyperledger/fabric/core/chaincode/shim"
        "github.com/hyperledger/fabric/protos/peer"
)

type Captable struct {

}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Captable) Init(stub shim.ChaincodeStubInterface) peer.Response {
        return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *Captable) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
        // Extract the function and args from the transaction proposal
        fn, args := stub.GetFunctionAndParameters()

        var result string
        var err error
        if fn == "save" {
                result, err = save(stub, args)
        } else if fn =="get" {
                result, err = get(stub, args)
        } else if fn == "delete" {
                err = delete(stub,args)
        } else{
                err = nil
        }

        if err != nil {
                return shim.Error(err.Error())
        }


        return shim.Success([]byte(result))
}


func save(stub shim.ChaincodeStubInterface, args []string) (string, error) {
        if len(args) != 4 {
                return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
        }
        key, er := stub.CreateCompositeKey("captable",[]string{args[0],args[1],args[2]})
        if er!=nil{
                return "", fmt.Errorf("Failed to Create Composite Key")
        }
        err := stub.PutState(key, []byte(args[3]))
        if err != nil {
                return "", fmt.Errorf("Failed to set asset: %s", key)
        }
        return key, nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
        if len(args) != 3 {
                return "", fmt.Errorf("Incorrect arguments. Expecting a key")
        }
        key, er := stub.CreateCompositeKey("captable",[]string{args[0],args[1],args[2]})
        if er!=nil{
                return "", fmt.Errorf("Failed to Create Composite Key")
        }
        value, err := stub.GetState(key)
        if err != nil {
                return "", fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
        }
        if value == nil {
                return "", fmt.Errorf("Asset not found: %s", key)
        }
        return string(value), nil
}
// Get returns the value of the specified asset key
func delete(stub shim.ChaincodeStubInterface, args []string) (error) {
        if len(args) != 3 {
                return fmt.Errorf("Incorrect arguments. Expecting a key")
        }
        key, er := stub.CreateCompositeKey("captable",[]string{args[0],args[1],args[2]})
        if er!=nil{
                return fmt.Errorf("Failed to Create Composite Key")
        }
        err := stub.DelState(key)
        if err != nil {
                return fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
        }

        return nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
        if err := shim.Start(new(Captable)); err != nil {
                fmt.Printf("Error starting Captable chaincode: %s", err)
        }
}
