/*
 SPDX-License-Identifier: Apache-2.0
*/

// =========================== CHAINCODE EXECUTION SAMPLES (CLI) =================================
/* 
key1 : did인증 제공서비스의 identifier
       향후 복수개의 인증서비스 수용을 고려
key2 : hash(회원의 주민번호)
       회원번호와는 다름
1)save : did인증을 통한 신규회원가입 발생시
peer chaincode invoke -C didauth -n didauth -c '{"Args":["save", "ssw", "userAuthNum", "userDid", "userDidDoc", "pairwiseDid", "pairwiseDidDoc"]}'

2) update : did credential 변경 및 유효기간 만료 등으로 재발급 발생시
peer chaincode invoke -C didauth -n didauth -c '{"Args":["update", "ssw", "userAuthNum", "userDidU", "userDidDocU", "pairwiseDidU", "pairwiseDidDocU"]}'

3) delete : 회원 탈퇴에 따른 삭제- 삭제로 상태 변경만 하며 실제 delstate를 수행하지 않음
peer chaincode invoke -C didauth -n didauth -c '{"Args":["delete", "ssw", "userAuthNum"]}'

4) get : 회원의hash(주민번호) key값에 따른 did 인증정보를 리턴
peer chaincode invoke -C didauth -n didauth -c '{"Args":["get", "ssw", "userAuthNum"]}'

5) gethistory : 특정 did인증정보의 모든 이력을 조회
peer chaincode invoke -C didauth -n didauth -c '{"Args":["gethistory", "ssw", "userAuthNum"]}'
====================================================================================================*/

 
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	_"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// DidAuthChaincode example simple Chaincode implementation
type DidAuthChaincode struct {
}

type auth struct {
	AuthSvcId		string `json:"authSvcId"`     	// auth service name - can be multiple  in the future
	UserAuthNum		string `json:"userAuthNum"`    	// sha256(identification number - juminbunho)
	UserDid			string `json:"userDid"` 		// holders did 
	UserDidDoc		string `json:"userDidDoc"`		// holders did document (includes pubkey)
	PairwiseDid 	string `json:"pairwiseDid"`		
	PairwiseDidDoc	string `json:"pairwiseDidDoc"`
	Status			string `json:"status"`
}

const (
	ITEMNUM	= 6
	KEYNUM	= 2
	NEW		= "new"
	RENEW	= "renew"
	DISUSE	= "disuse"
	INDEX 	= "baseIndex"
)


// ============================================================
// save
// ============================================================
func (t *DidAuthChaincode) save(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	if len(args) != ITEMNUM {
		return shim.Error("Incorrect number of arguments.")
	}

	// ============= Input sanitation ===============
	fmt.Println("- start save")
	
	authSvcId := args[0]
	userAuthNum := args[1]
	userDid := args[2]
	userDidDoc := args[3]
	pairwiseDid := args[4]
	pairwiseDidDoc := args[5]
	status := NEW
	
	// ===== create composite key ===================
	authKey, err := stub.CreateCompositeKey(INDEX, []string{authSvcId, userAuthNum})
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("authKey" + authKey)
	
	// ==== Check if user auth info already exists ====
	userAuthAsBytes, err := stub.GetState(authKey)
	if err != nil {
		return shim.Error("Failed to get userAuthInfo: " + err.Error())
	} else if userAuthAsBytes != nil { //--> update
		fmt.Println("This userAuthInfo already exists: " + authKey)
		return shim.Error("This userAuthinfo already exists: " + authKey)
	}

	
	// ==== Create auth object and marshal to JSON ====
	authObj := &auth{authSvcId, userAuthNum, userDid, userDidDoc, pairwiseDid, pairwiseDidDoc, status}
	authJSONasBytes, err := json.Marshal(authObj)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("authJSONasBytes" + string(authJSONasBytes))
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)
		
	
	// === Save userAuthInfo to state ===
	err = stub.PutState(authKey, authJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== create authKeyWithStat  index to enable status-based range queries, e.g. return all active auth ====
	/*authKeyByStat, err := stub.CreateCompositeKey("statIndex", []string{status, authSvcId, userAuthNum })
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the authObj
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(authKeyByStat, value)
    */
	// ==== userAuthInfo saved and indexed. Return success ====
	fmt.Println("- end save userAuthInfo")
	return shim.Success(nil)

}


// ============================================================
// update
// ============================================================
func (t *DidAuthChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != ITEMNUM {
		return shim.Error("Incorrect number of arguments.")
	}

	// ============= Input sanitation ==============
	fmt.Println("- start update")
	
	authSvcId := args[0]
	userAuthNum := args[1]
	userDid := args[2]
	userDidDoc := args[3]
	pairwiseDid := args[4]
	pairwiseDidDoc := args[5]
	status := RENEW
	
	// ===== create composite key ===================
	authKey, err := stub.CreateCompositeKey(INDEX, []string{authSvcId, userAuthNum})
	if err != nil {
		return shim.Error(err.Error())
	}
	
	// ==== Check if user auth info exists ====
	userAuthAsBytes, err := stub.GetState(authKey)
	if err != nil {
		return shim.Error("Failed to get userAuthInfo: " + err.Error())
	} else if userAuthAsBytes == nil { //--> save
		fmt.Println("This userAuthInfo no exists: " + authKey)
		return shim.Error("This userAuthinfo no exists: " + authKey)
	}

	
	// ==== Create auth object and marshal to JSON ====
	authObj := &auth{authSvcId, userAuthNum, userDid, userDidDoc, pairwiseDid, pairwiseDidDoc, status }
	authJSONasBytes, err := json.Marshal(authObj)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)
		
	
	// === Save userAuthInfo to state ===
	err = stub.PutState(authKey, authJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== create authKeyWithStat  index to enable status-based range queries, e.g. return all active auth ====
	/*authKeyByStat, err := stub.CreateCompositeKey("statIndex", []string{authSvcId, status})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the authObj
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(authKeyByStat, value)
	*/
	
	// ==== userAuthInfo saved and indexed. Return success ====
	fmt.Println("- end update userAuthInfo")
	return shim.Success(nil)

}

// ============================================================
// delete
// ============================================================
func (t *DidAuthChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != KEYNUM {
		return shim.Error("Incorrect number of arguments.")
	}

	// ============== Input sanitation ===============
	fmt.Println("- start delete")
	
	authSvcId := args[0]
	userAuthNum := args[1]
	status := DISUSE
	
	// ===== create composite key ===================
	authKey, err := stub.CreateCompositeKey(INDEX, []string{authSvcId, userAuthNum})
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Check if user auth info exists ====
	userAuthAsBytes, err := stub.GetState(authKey)
	if err != nil {
		return shim.Error("Failed to get userAuthInfo: " + err.Error())
	} else if userAuthAsBytes == nil { //--> save
		fmt.Println("userAuthInfo to be deleted no exists: " + authKey)
		return shim.Error("userAuthinfo to be deleted no exists: " + authKey)
	}
	
	// ==== unmarshal ====
	authToDisuse := auth{}
	err = json.Unmarshal(userAuthAsBytes, &authToDisuse) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	authToDisuse.Status = status //change status
	
	// ==== marshal ====
	authJSONasBytes, _ := json.Marshal(authToDisuse)
	err = stub.PutState(authKey, authJSONasBytes) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end delete (success)")
	return shim.Success(nil)
}


// ============================================================
// get
// ============================================================
func (t *DidAuthChaincode) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	if len(args) != KEYNUM {
		return shim.Error("Incorrect number of arguments.")
	}

	// =========== Input sanitation =================
	fmt.Println("- start get")
	
	authSvcId := args[0]
	userAuthNum := args[1]

	
	// ===== create composite key ===================
	authKey, err := stub.CreateCompositeKey(INDEX, []string{authSvcId, userAuthNum})
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Check if user auth info exists ====
	userAuthAsBytes, err := stub.GetState(authKey)
	if err != nil {
		return shim.Error("Failed to get userAuthInfo: " + err.Error())
	} else if userAuthAsBytes == nil { //--> save
		fmt.Println("userAuthInfo no exists: " + authKey)
		return shim.Error("userAuthinfo no exists: " + authKey)
	}
	
	fmt.Printf("- get returning:\n%s\n", string(userAuthAsBytes))
	return shim.Success(userAuthAsBytes)
	

}
// ============================================================
// getHistory
// ============================================================
func (t *DidAuthChaincode) getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != KEYNUM {
		return shim.Error("Incorrect number of arguments.")
	}

	authSvcId := args[0]
	userAuthNum := args[1]
	
	fmt.Printf("- start getHistory: %s\n", authSvcId+userAuthNum)

	// ===== create composite key ===================
	authKey, err := stub.CreateCompositeKey(INDEX, []string{authSvcId, userAuthNum})
	if err != nil {
		return shim.Error(err.Error())
	}
	
	
	resultsIterator, err := stub.GetHistoryForKey(authKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the authInfo
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}


// Invoke - Our entry point for Invocations
// ========================================
func (t *DidAuthChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "save" { 
		return t.save(stub, args)
	} else if function == "update" { 
		return t.update(stub, args)
	} else if function == "delete" { 
		return t.delete(stub, args)
	} else if function == "get" { 
		return t.get(stub, args)
	} else if function == "getHistory" { 
		return t.getHistory(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}


// Init initializes chaincode
// ===========================
func (t *DidAuthChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(DidAuthChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}


