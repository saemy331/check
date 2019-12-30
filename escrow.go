/* 상태 코드갱신에 대한유효성 검사요건
[before] |	[after]	
		 |	300
300	     |  311	개시 상태에서 입금거부 기록 가능
311	     |	311 입금거부 상태에서 재거부 기록 가능
300	  	 |	320	개시 상태에서 입금완료 기록 가능
311	     |	320	입금거부상태에서 입금완료 기록 가능
320	     |	330	입금완료 상태에서 매도자계좌이체지시 가능 (*was 업무단은 24시간 경과와 기업승인 체크)
330	     |	331	매도자계좌이체 지시 상태에서 매도자 계좌이체 실행 결과(실패, 성공) 기록 가능
331(fail)|	331	매도자계좌이체실패시 재이체 실행 결과(실패, 성공) 기록 가능
320	     |	340	입금완료시 환불요청 가능(하나은행 중앙계좌에 대금 존재 상태)
330	     |	340	매도자계좌이체지시 하였으나, 아직 이체 실행전 환불요청 가능 (* 시간차로 인한 이체 실행에 대해선 방법 없음)
331(fail)|	340 매도자계좌이체 실패시 환불요청 가능
340	  	 |	341	환불요청이 있을 경우에만 환불처리 가능
*/

package main


import( 
    "bytes"
    "encoding/json"
    "fmt"
    "strconv"
    "time"

    "github.com/hyperledger/fabric/core/chaincode/shim"
    pb "github.com/hyperledger/fabric/protos/peer"
)

type EscrowChaincode struct {
}

type escrowTx struct {
	VaNo           string `json:"VaNo"`            //가상계좌번호
	CtrctNo        string `json:"CtrctNo"`         //계약번호
	StatusCode     string `json:"StatusCode"`      //상태코드
    Crcy           string `json:"Crcy"`            //통화
    Depositor      string `json:"Depositor"`       //입금자2019.08.30 추가  
	DepositAmt     string `json:"DepositAmt"`      //입금금액
    DepositOpenDttm  string `json:"DepositOpenDttm"`  //입금시작일자
    DepositCloseDttm string `json:"DepositCloseDttm"` //입금마감일자
    DepositDttm    string `json:"DepositDttm"`     //입금시각
    TrsfBnkCode    string `json:"TrsfBnkCode"`     //지급은행코드
    TrsfBnkAcnt    string `json:"TrsfBnkAcnt"`     //지급계좌번호
    TrsfAmt        string `json:"TrsfAmt"`         //지급금액
    TrsfReqDttm    string `json:"TrsfReqDttm"`     //지급요청시각
    TrsfDttm       string `json:"TrsfDttm"`        //지급시각
    RfundBnkCode   string `json:"RfundBnkCode"`    //환불은행코드
    RfundBnkAcnt   string `json:"RfundBnkAcnt"`    //환불계좌번호
    RfundAmt       string `json:"RfundAmt"`        //환불금액
    RfundReqDttm   string `json:"RfundReqDttm"`    //환불요청시각
    RfundDttm      string `json:"RfundDttm"`       //환불시각
    TrdFeeAmt      string `json:"TrdFeeAmt"`       //매매수수료금액
    EscrFeeAmt     string `json:"EscrFeeAmt"`      //은행수수료금액
    AdjBaseDt      string `json:"AdjBaseDt"`       //정산기준일
    ReadYn         string `json:"ReadYn"`          //Read여부(처리여부)
    ErrCode        string `json:"ErrCode"`         //에러코드 
	ErrRsn		   string `json:"ErrRsn"`		   //에러사유
    ChgDttm        string `json:"ChgDttm"`         //변경시각
}

const (
    TRANSFERCOMPLETESUCCESS = "NCOM00000"
)


// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(EscrowChaincode))
	if err != nil {
		fmt.Printf("Error starting Escrow chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *EscrowChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *EscrowChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "startNewEscrow" {              //에스크로 거래 신규 생성 : 300 (koscom)
		return t.startNewEscrow(stub, args)
	} else if function == "rejectDeposit" {       //가상계좌 입금 거부 :311 (bank)
		return t.rejectDeposit(stub, args)
	} else if function == "confirmDeposit" {       //가상계좌 입금 확인 :320 (bank)
		return t.confirmDeposit(stub, args)
	} else if function == "orderFundTransfer" {   //계좌 이체지시 :330 (koscom)
		return t.orderFundTransfer(stub, args)
	} else if function == "transferComplete" {     //계좌 이체처리 완료 : 331 (bank)
		return t.transferComplete(stub, args)
	} else if function == "orderRefund" {          //환불 지시 : 340 (koscom)
		return t.orderRefund(stub, args)
	} else if function == "refundComplete" {       //환불처리 완료 : 341 (bank)
		return t.refundComplete(stub, args)
    } else if function == "getEscrowHistory" {     //에스크로 내역 조회
    	return t.getEscrowHistory(stub, args)
    } else if function == "readEscrowStatus" {     //에스크로 미확인 건 조회
    	return t.readEscrowStatus(stub, args)
    } else if function == "readOnlyEscrowStatus" {     //에스크로 미확인 건 조회Only
    	return t.readOnlyEscrowStatus(stub, args)
    } else if function == "readEscrowStatusByVaNo" {     //계좌처리상태 조회
    	return t.readEscrowStatusByVaNo(stub, args)
    }
 

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// 에스크로 거래 신규 생성 : 300 (koscom)
// ============================================================
func (t *EscrowChaincode) startNewEscrow(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 13 {
		return shim.Error("Incorrect number of arguments. Expecting 13")
	}

	// ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
    if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
    if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
    if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
    if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
    if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}
    if len(args[9]) <= 0 {
		return shim.Error("10th argument must be a non-empty string")
	}
    if len(args[10]) <= 0 {
		return shim.Error("11th argument must be a non-empty string")
	}
	if len(args[11]) <= 0 {
	    return shim.Error("11th argument must be a non-empty string")
	}
	// ssi add
	if len(args[12]) <= 0 {
		return shim.Error("12th argument must be a non-empty string")
	}


    //time := time.Now()

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "300"
    Crcy := args[2]
	Depositor := args[3] // 2019.08.30 추가
    DepositAmt := args[4]
    DepositOpenDttm := args[5]
    DepositCloseDttm := args[6]
    TrsfBnkCode := args[7]
    TrsfBnkAcnt := args[8]
    TrsfAmt := args[9]
    TrdFeeAmt := args[10]
    EscrFeeAmt := args[11]
    ReadYn := "N"
    // YYYYMMDDHHmmss
    // ChgDttm := fmt.Sprintf("%d%02d%02d%02d%02d%02d\n", time.Year(), time.Month(), time.Day(),time.Hour(), time.Minute(), time.Second())
    /*ts, tserr := stub.GetTxTimestamp()
	if tserr != nil {
		fmt.Printf("Error getting transaction timestamp: %s", tserr)
	    return shim.Error(fmt.Sprintf("Error getting transaction timestamp: %s", tserr))
	}
	fmt.Printf("Transaction Time : %v\n", ts)
	*/
	ChgDttm := args[12]

	DepositDttm := ""
    TrsfReqDttm := ""
    TrsfDttm := ""
    RfundBnkCode := ""
    RfundBnkAcnt := ""
    RfundAmt := ""
    RfundReqDttm := ""
    RfundDttm := ""
    AdjBaseDt := ""
    ErrCode := ""
	ErrRsn := ""

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 존재여부 확인 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes != nil {
		fmt.Println("This escrow tx already exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx already exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

	// ==== 에스크로 tx 신규건 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt,RfundReqDttm,RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ReadYn, ErrCode, ErrRsn, ChgDttm}
	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === 에스크로 Tx 저장 ===
	err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)

	fmt.Println("- end startNewEscrow()")
	return shim.Success(nil)
}

// ============================================================
// 가상계좌 입금 거부 :311 (bank)
// ============================================================
func (t *EscrowChaincode) rejectDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments.") 
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
    if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "311"
    ErrCode := args[2]
	ErrRsn := args[3]
    ChgDttm := args[4]
    ReadYn := "N"

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    if escrowTxJSON.StatusCode != "300" && escrowTxJSON.StatusCode != "311"{
        fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
        jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
        return shim.Error(jsonResp)
    }
   
   // ssi add del index = 기존인덱스 삭제
   EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})  
   if err != nil {
   	 	return shim.Error(err.Error())
   }

   err = stub.DelState(EscrowCurrIndexKey)
   if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
   }
   
    // ==== 에스크로 tx 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ReadYn, ErrCode, ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)

    fmt.Println("- end rejectDeposit()")
    return shim.Success(nil)

}


// ============================================================
// 가상계좌 입금 확인 :320 (bank)
// ============================================================
func (t *EscrowChaincode) confirmDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments.") 
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "320"
    DepositDttm := args[2]
    ChgDttm := args[3]
    ReadYn := "N"
    ErrCode := ""
	ErrRsn := ""

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    if escrowTxJSON.StatusCode != "300" && escrowTxJSON.StatusCode != "311" {
		fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
	    jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
        return shim.Error(jsonResp)
    }

   // ssi add del index = 기존인덱스 삭제
   EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})
   if err != nil {
        return shim.Error(err.Error())
   }

   err = stub.DelState(EscrowCurrIndexKey)
   if err != nil {
        return shim.Error("Failed to delete state:" + err.Error())
   }

    // ==== 에스크로 tx 생성 ====
    // ==== 오류사유(ErrCode)을 escrowTxJSON의 값을쓰지 않고 공백을 넣는 이유는,
    //      311->320 으로 넘어오는 경우 ErrCode이 채워져서 들어갈 것인데, 320 상태코드는 정상을 뜻하므로 삭제를 해주어야 하기 떄문
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ReadYn, ErrCode, ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy,  escrowTxJSON.Depositor,escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)


    fmt.Println("- end confirmDeposit()")
    return shim.Success(nil)
}

// ============================================================
// 계좌 이체지시 :330 (koscom)
// ============================================================
func (t *EscrowChaincode) orderFundTransfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments.")
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	//ssi add
    if len(args[3]) <= 0 {
	    return shim.Error("4th argument must be a non-empty string")
    }

    //time := time.Now()

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "330"
    TrsfReqDttm := args[2]
    ReadYn := "N"
    // YYYYMMDDHHmmss
    //ChgDttm := fmt.Sprintf("%d%02d%02d%02d%02d%02d\n",
    //           time.Year(), time.Month(), time.Day(),time.Hour(), time.Minute(), time.Second())
    ChgDttm:= args[3]

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    if escrowTxJSON.StatusCode != "320" {
    	fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
	    jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
		return shim.Error(jsonResp)
    }

    // ssi add del index = 기존인덱스 삭제
    EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})
	if err != nil {
	   return shim.Error(err.Error())
	}

	err = stub.DelState(EscrowCurrIndexKey)
	if err != nil {
	   return shim.Error("Failed to delete state:" + err.Error())
	}




    // ==== 에스크로 tx 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy,  escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ReadYn, escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy,  escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)


    fmt.Println("- end orderFundTransfer()")
    return shim.Success(nil)
}

// ============================================================
// 계좌 이체처리 완료 : 331 (bank)
// ============================================================
func (t *EscrowChaincode) transferComplete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments.") 
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4nd argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5rd argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
	    return shim.Error("5rd argument must be a non-empty string")
	}
    if len(args[6]) <= 0 {
	    return shim.Error("5rd argument must be a non-empty string")
	}

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "331"
    TrsfDttm := args[2]
    AdjBaseDt := args[3]
	ErrCode := args[4] //2019.08.30
	ErrRsn := args[5] //2019.08.30
    ChgDttm := args[6]
    ReadYn := "N"

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    if escrowTxJSON.StatusCode != "330" && !(escrowTxJSON.StatusCode == "331" && escrowTxJSON.ErrCode != TRANSFERCOMPLETESUCCESS) {
	    fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
	    jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
        return shim.Error(jsonResp)
    }

    // ssi add del index = 기존인덱스 삭제
    EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})
	if err != nil {
	   return shim.Error(err.Error())
	}

	err = stub.DelState(EscrowCurrIndexKey)
	if err != nil {
	   return shim.Error("Failed to delete state:" + err.Error())
	}






    // ==== 에스크로 tx 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, AdjBaseDt, ReadYn, ErrCode, ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)


    fmt.Println("- end transferComplete()")
    return shim.Success(nil)
}

// ============================================================
// 환불 지시 : 340 (koscom)
// ============================================================
func (t *EscrowChaincode) orderRefund(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments.") 
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
    if len(args[3]) <= 0 {
		return shim.Error("4st argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5nd argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6rd argument must be a non-empty string")
	}
	// ssi add
	if len(args[6]) <= 0 {
	    return shim.Error("7th argument must be a non-empty string")
	}


    //time := time.Now()

    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "340"
    RfundBnkCode := args[2]
    RfundBnkAcnt := args[3]
    RfundAmt := args[4]
    RfundReqDttm := args[5]
    ReadYn := "N"
    // YYYYMMDDHHmmss
    //ChgDttm := fmt.Sprintf("%d%02d%02d%02d%02d%02d\n",
    //           time.Year(), time.Month(), time.Day(),time.Hour(), time.Minute(), time.Second())
    ChgDttm := args[6]

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    // 환불이 가능한 상태코드는 어디까지 보아야 하는가??
    if escrowTxJSON.StatusCode != "320" && escrowTxJSON.StatusCode != "330" && !(escrowTxJSON.StatusCode == "331" && escrowTxJSON.ErrCode != TRANSFERCOMPLETESUCCESS){
		fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
	    jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
       
        return shim.Error(jsonResp)
    }

     // ssi add del index = 기존인덱스 삭제
	 EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})
	 if err != nil {
	    return shim.Error(err.Error())
	 }
     err = stub.DelState(EscrowCurrIndexKey)
	 if err != nil {
	    return shim.Error("Failed to delete state:" + err.Error())
	 }



    // ==== 에스크로 tx 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ReadYn, escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)


    fmt.Println("- end orderRefund()")
    return shim.Success(nil)
}

// ============================================================
// 환불처리 완료 : 341 (bank)
// ============================================================
func (t *EscrowChaincode) refundComplete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
	var jsonResp string
    var escrowTxJSON escrowTx

	if len(args) != 6 { // 331과 동일하게 환불에러코드 및 사유 항목 추가 1112
		return shim.Error("Incorrect number of arguments. ")
	}

    // ==== 입력값 검증 ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4nd argument must be a non-empty string")
	}
    if len(args[4]) <= 0 {
	    return shim.Error("4nd argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
	    return shim.Error("4nd argument must be a non-empty string")
    }
 
    VaNo := args[0]
    CtrctNo := args[1]
    StatusCode := "341"
    RfundDttm := args[2]
	ErrCode := args[3] //2019.08.30
	ErrRsn := args[4] //2019.08.30
    ChgDttm := args[5]
    ReadYn := "N"

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	// ==== 에스크로 tx 조회 ====
	escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
	if err != nil {
		return shim.Error("Failed to get escrow Tx: " + err.Error())
	} else if escrowAsBytes == nil {
		fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
	}

    // Unmashal
    err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
    	if err != nil {
    		jsonResp = "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
    		return shim.Error(jsonResp)
    	}

    // check consistency
    if escrowTxJSON.StatusCode != "340" && !(escrowTxJSON.StatusCode == "341" && escrowTxJSON.ErrCode != TRANSFERCOMPLETESUCCESS) {
    	fmt.Println("It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "])")
	    jsonResp = "{\"Error\":\" : It cannot be processed because existing escrowTx's statusCode is "+escrowTxJSON.StatusCode +".(VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
        return shim.Error(jsonResp)
    }

	 // ssi add del index = 기존인덱스 삭제
       EscrowCurrIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{escrowTxJSON.ReadYn, escrowTxJSON.StatusCode, escrowTxJSON.VaNo, escrowTxJSON.CtrctNo})
		   if err != nil {
		       return shim.Error(err.Error())
       }

       err = stub.DelState(EscrowCurrIndexKey)
		    if err != nil {
  		        return shim.Error("Failed to delete state:" + err.Error())
	   }





    // ==== 에스크로 tx 생성 ====
    escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ReadYn, ErrCode, ErrRsn, ChgDttm}

	escrowTxJSONasBytes, err := json.Marshal(escrowTx)
	if err != nil {
		return shim.Error(err.Error())
	}

    // === 에스크로 Tx 저장 ===
    err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // make index (for readEscrowStatus())
    //EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ ReadYn, StatusCode, VaNo, CtrctNo, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, ErrCode, ErrRsn,  ChgDttm})
    EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{ReadYn, StatusCode, VaNo, CtrctNo}) //index key chg ssi
	if err != nil {
        return shim.Error(err.Error())
    }

    // === save index ===
    value := []byte{0x00}
    stub.PutState(EscrowIndexKey, value)


    fmt.Println("- end refundComplete()")
    return shim.Success(nil)
}

// ============================================================
// 에스크로 내역 조회
// ============================================================
func (t *EscrowChaincode) getEscrowHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

    if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

    VaNo := args[0]
    CtrctNo := args[1]
	fmt.Println("- start getEscrowHistory() ==> VaNo : "+ VaNo+" CtrctNo :" +CtrctNo)

	fmt.Printf("- start getEscrowHistory\n")

    // make compositeKey
    EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
    if err != nil {
        return shim.Error(err.Error())
    }

	resultsIterator, err := stub.GetHistoryForKey(EscrowCompositeKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values 
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
		//as-is (as the Value itself a JSON)
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

	fmt.Printf("- getEscrowHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ============================================================
// 에스크로 미확인 건 조회 : 인자로 입력된 상태코드와 처리여부(ReadYn) 에 부합되는 KVS 데이터를 출력하기 위함
// ============================================================
func (t *EscrowChaincode) readEscrowStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
    var buffer bytes.Buffer
	var escrowTxJSON escrowTx //ssi change
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

    StatusCodeIn := args[0]
    ReadYnIn := args[1]
	fmt.Println("- start readEscrowStatus() ==> StatusCode : "+ StatusCodeIn+" ReadYn :" +ReadYnIn)

    //EscrowIndexKey로 생성해놓았던 데이터를 조회하기 위한 key 생성
	StatuCodeReadYnResultIterator, err := stub.GetStateByPartialCompositeKey("EscrowIndexKey", []string{ReadYnIn, StatusCodeIn})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer StatuCodeReadYnResultIterator.Close()

	// 반복수행하며 해당되는 데이터 set 조회
	var i int
    buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false	
	for i = 0; StatuCodeReadYnResultIterator.HasNext(); i++ {
		responseRange, err := StatuCodeReadYnResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// data Set 하나를 가져와서 각 항목에 저장
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		ReadYn := compositeKeyParts[0]
		StatusCode := compositeKeyParts[1]
        VaNo := compositeKeyParts[2]
        CtrctNo := compositeKeyParts[3]
        fmt.Println(" -- compositeKeyParts : " + ReadYn + " "  + StatusCode + " " + VaNo + " " + CtrctNo )
		/*
		Crcy := compositeKeyParts[4]
		Depositor := compositeKeyParts[5]
        DepositAmt := compositeKeyParts[6]
        DepositOpenDttm := compositeKeyParts[7]
        DepositCloseDttm := compositeKeyParts[8]
        DepositDttm := compositeKeyParts[9]
        TrsfBnkCode := compositeKeyParts[10]
        TrsfBnkAcnt := compositeKeyParts[11]
        TrsfAmt := compositeKeyParts[12]
        TrsfReqDttm := compositeKeyParts[13]
        TrsfDttm := compositeKeyParts[14]
        RfundBnkCode := compositeKeyParts[15]
        RfundBnkAcnt := compositeKeyParts[16]
        RfundAmt := compositeKeyParts[17]
        RfundReqDttm := compositeKeyParts[18]
        RfundDttm := compositeKeyParts[19]
        TrdFeeAmt := compositeKeyParts[20]
        EscrFeeAmt := compositeKeyParts[21]
        AdjBaseDt := compositeKeyParts[22]
        ErrCode := compositeKeyParts[23]
		ErrRsn := compositeKeyParts[24]
        // YYYYMMDDHHmmss
        ChgDttm :=compositeKeyParts[25]
		*/

        // ssi change 

        EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
		if err != nil {
			return shim.Error(err.Error())
		}
		
		escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
		if err != nil {
			return shim.Error("Failed to get escrow Tx: " + err.Error())
        } else if escrowAsBytes == nil {
		    fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
			return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		}
		
		err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
		if err != nil {
		    jsonResp := "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
		    return shim.Error(jsonResp)
		}

        
        // Add a comma before array members, suppress it for the first array member
   		if bArrayMemberAlreadyWritten == true {
   			buffer.WriteString(",")
   		}

   		buffer.WriteString("{\"VaNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.VaNo)
		buffer.WriteString("\"")
        
		buffer.WriteString(", \"CtrctNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.CtrctNo)
		buffer.WriteString("\"")

        buffer.WriteString(", \"StatusCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.StatusCode)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Crcy\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.Crcy)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Depositor\":")
		buffer.WriteString("\"")
		buffer.WriteString(escrowTxJSON.Depositor)
		buffer.WriteString("\"")

        buffer.WriteString(", \"DepositAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositOpenDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositOpenDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositCloseDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositCloseDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrdFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrdFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"EscrFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.EscrFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"AdjBaseDt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.AdjBaseDt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ReadYn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ReadYn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrRsn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrRsn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ChgDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ChgDttm)
        buffer.WriteString("\"")
		
        buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true

        // 읽기여부(처리여부) == "N" 인 경우 읽기여부를 "Y" 로 변경 후 업뎃처리
        if ReadYn == "N" {
			fmt.Println("change readYn value from N to Y : "+VaNo+","+CtrctNo+","+StatusCode)
            // ==== 에스크로 tx 생성 ====
            //escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, "Y", ErrCode, ErrRsn, ChgDttm}
		  	escrowTx := &escrowTx{escrowTxJSON.VaNo, escrowTxJSON.CtrctNo, escrowTxJSON.StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, "Y", escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, escrowTxJSON.ChgDttm}

			escrowTxJSONasBytes, err := json.Marshal(escrowTx)
			if err != nil {
				return shim.Error(err.Error())
			}

			// make compositeKey
			EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
			if err != nil {
				return shim.Error(err.Error())
			}

			// === 에스크로 Tx 저장 ===
			err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		   // make index (for remove)
           // EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
            EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo}) //index key chg ssi
			if err != nil {
	            return shim.Error(err.Error())
	        }

			// === delete index ===
			err = stub.DelState(EscrowIndexKey)
			if err != nil {
				return shim.Error("Failed to delete state:" + err.Error())
			}
	
			// make index (for new index)
			//EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
		    EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo}) //index key chg ssi
		    if err != nil {
				return shim.Error(err.Error())
			}
			
			// === save index ===
			value := []byte{0x00}
		    stub.PutState(EscrowIndexKey, value)
		}
	}
    buffer.WriteString("]")

    fmt.Println("readEscrowStatus queryResult : \n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ============================================================
// 에스크로 미확인 건 조회Only : 인자로 입력된 상태코드와 처리여부(ReadYn) 에 부합되는 KVS 데이터를 출력하기 위함
// ============================================================
func (t *EscrowChaincode) readOnlyEscrowStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
    var buffer bytes.Buffer
    var escrowTxJSON escrowTx

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

    StatusCodeIn := args[0]
    ReadYnIn := args[1]
	fmt.Println("- start readOnlyEscrowStatus() ==> StatusCode : "+ StatusCodeIn+" ReadYn :" +ReadYnIn)

    //EscrowIndexKey로 생성해놓았던 데이터를 조회하기 위한 key 생성
	StatuCodeReadYnResultIterator, err := stub.GetStateByPartialCompositeKey("EscrowIndexKey", []string{ReadYnIn, StatusCodeIn})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer StatuCodeReadYnResultIterator.Close()

	// 반복수행하며 해당되는 데이터 set 조회
	var i int
    buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false	
	for i = 0; StatuCodeReadYnResultIterator.HasNext(); i++ {
		responseRange, err := StatuCodeReadYnResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// data Set 하나를 가져와서 각 항목에 저장
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		ReadYn := compositeKeyParts[0]
		StatusCode := compositeKeyParts[1]
        VaNo := compositeKeyParts[2]
        CtrctNo := compositeKeyParts[3]
        /*
		Crcy := compositeKeyParts[4]
		Depositor := compositeKeyParts[5]
        DepositAmt := compositeKeyParts[6]
        DepositOpenDttm := compositeKeyParts[7]
        DepositCloseDttm := compositeKeyParts[8]
        DepositDttm := compositeKeyParts[9]
        TrsfBnkCode := compositeKeyParts[10]
        TrsfBnkAcnt := compositeKeyParts[11]
        TrsfAmt := compositeKeyParts[12]
        TrsfReqDttm := compositeKeyParts[13]
        TrsfDttm := compositeKeyParts[14]
        RfundBnkCode := compositeKeyParts[15]
        RfundBnkAcnt := compositeKeyParts[16]
        RfundAmt := compositeKeyParts[17]
        RfundReqDttm := compositeKeyParts[18]
        RfundDttm := compositeKeyParts[19]
        TrdFeeAmt := compositeKeyParts[20]
        EscrFeeAmt := compositeKeyParts[21]
        AdjBaseDt := compositeKeyParts[22]
        ErrCode := compositeKeyParts[23]
		ErrRsn := compositeKeyParts[24]
        // YYYYMMDDHHmmss
        ChgDttm :=compositeKeyParts[25]
        */

		fmt.Println("-ssi " + ReadYn + " " + StatusCode + " " + VaNo + " " + CtrctNo)
        // ssi change
        EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
        if err != nil {
		    return shim.Error(err.Error())
		}

		escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
		if err != nil {
			return shim.Error("Failed to get escrow Tx: " + err.Error())
		} else if escrowAsBytes == nil {
			fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
			return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		}

        err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
        if err != nil {
	    	jsonResp := "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
			return shim.Error(jsonResp)																									   }


        // Add a comma before array members, suppress it for the first array member
   		if bArrayMemberAlreadyWritten == true {
   			buffer.WriteString(",")
   		}

   		buffer.WriteString("{\"VaNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.VaNo)
   		buffer.WriteString("\"")

		buffer.WriteString(", \"CtrctNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.CtrctNo)
   		buffer.WriteString("\"")

        buffer.WriteString(", \"StatusCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.StatusCode)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Crcy\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.Crcy)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Depositor\":")
		buffer.WriteString("\"")
		buffer.WriteString(escrowTxJSON.Depositor)
		buffer.WriteString("\"")

        buffer.WriteString(", \"DepositAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositOpenDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositOpenDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositCloseDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositCloseDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrdFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrdFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"EscrFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.EscrFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"AdjBaseDt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.AdjBaseDt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ReadYn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ReadYn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrRsn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrRsn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ChgDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ChgDttm)
        buffer.WriteString("\"")
        buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true

        
	}
    buffer.WriteString("]")

    fmt.Println("readOnlyEscrowStatus queryResult : \n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ============================================================
// 계좌 처리상태 조회 : 인자로 입력된 가상계좌번호, 상태코드와 처리여부(ReadYn) 에 부합되는 KVS 데이터를 출력하기 위함(하나은행 가상계좌로 입금이 되면 하나은행에서는 해당 가상계좌번호+상태코드(300)+readYn(N) 의 조합으로 데이터를 찾아 311 혹은 320의 데이터를 생성 
// ============================================================
func (t *EscrowChaincode) readEscrowStatusByVaNo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	  var err error
    var buffer bytes.Buffer
    var escrowTxJSON escrowTx
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	VaNo := args[0]
    StatusCodeIn := args[1]
    ReadYnIn := args[2]
	fmt.Println("- start readEscrowStatus() ==> StatusCode : "+ StatusCodeIn+" ReadYn :" +ReadYnIn+" VaNo :"+VaNo)

    //EscrowIndexKey로 생성해놓았던 데이터를 조회하기 위한 key 생성
	StatuCodeReadYnResultIterator, err := stub.GetStateByPartialCompositeKey("EscrowIndexKey", []string{ReadYnIn, StatusCodeIn, VaNo})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer StatuCodeReadYnResultIterator.Close()

	// 반복수행하며 해당되는 데이터 set 조회
	var i int
    buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false	
	for i = 0; StatuCodeReadYnResultIterator.HasNext(); i++ {
		responseRange, err := StatuCodeReadYnResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// data Set 하나를 가져와서 각 항목에 저장
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		ReadYn := compositeKeyParts[0]
		StatusCode := compositeKeyParts[1]
        VaNo := compositeKeyParts[2]
        CtrctNo := compositeKeyParts[3]
        /*Crcy := compositeKeyParts[4]
		Depositor := compositeKeyParts[5]
        DepositAmt := compositeKeyParts[6]
        DepositOpenDttm := compositeKeyParts[7]
        DepositCloseDttm := compositeKeyParts[8]
        DepositDttm := compositeKeyParts[9]
        TrsfBnkCode := compositeKeyParts[10]
        TrsfBnkAcnt := compositeKeyParts[11]
        TrsfAmt := compositeKeyParts[12]
        TrsfReqDttm := compositeKeyParts[13]
        TrsfDttm := compositeKeyParts[14]
        RfundBnkCode := compositeKeyParts[15]
        RfundBnkAcnt := compositeKeyParts[16]
        RfundAmt := compositeKeyParts[17]
        RfundReqDttm := compositeKeyParts[18]
        RfundDttm := compositeKeyParts[19]
        TrdFeeAmt := compositeKeyParts[20]
        EscrFeeAmt := compositeKeyParts[21]
        AdjBaseDt := compositeKeyParts[22]
        ErrCode := compositeKeyParts[23]
		ErrRsn := compositeKeyParts[24]
        // YYYYMMDDHHmmss
        ChgDttm :=compositeKeyParts[25]
        */
      // ssi change
        EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
        if err != nil {
		    return shim.Error(err.Error())
		}

		escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
		if err != nil {
			return shim.Error("Failed to get escrow Tx: " + err.Error())
		} else if escrowAsBytes == nil {
			fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
			return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		}

        err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
        if err != nil {
	    	jsonResp := "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
			return shim.Error(jsonResp)																									   }


        // Add a comma before array members, suppress it for the first array member
   		if bArrayMemberAlreadyWritten == true {
   			buffer.WriteString(",")
   		}

   		buffer.WriteString("{\"VaNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.VaNo)
   		buffer.WriteString("\"")

		buffer.WriteString(", \"CtrctNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.CtrctNo)
   		buffer.WriteString("\"")

        buffer.WriteString(", \"StatusCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.StatusCode)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Crcy\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.Crcy)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Depositor\":")
		buffer.WriteString("\"")
		buffer.WriteString(escrowTxJSON.Depositor)
		buffer.WriteString("\"")

        buffer.WriteString(", \"DepositAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositOpenDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositOpenDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositCloseDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositCloseDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrdFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrdFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"EscrFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.EscrFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"AdjBaseDt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.AdjBaseDt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ReadYn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ReadYn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrRsn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrRsn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ChgDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ChgDttm)
        buffer.WriteString("\"")
        buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true

        
	
        // 읽기여부(처리여부) == "N" 인 경우 읽기여부를 "Y" 로 변경 후 업뎃처리
        if ReadYn == "N" {
			fmt.Println("change readYn value from N to Y : "+VaNo+","+CtrctNo+","+StatusCode)
            // ==== 에스크로 tx 생성 ====
            //escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, "Y", ErrCode, ErrRsn, ChgDttm}
		  	escrowTx := &escrowTx{escrowTxJSON.VaNo, escrowTxJSON.CtrctNo, escrowTxJSON.StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, "Y", escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, escrowTxJSON.ChgDttm}

			escrowTxJSONasBytes, err := json.Marshal(escrowTx)
			if err != nil {
				return shim.Error(err.Error())
			}

			// make compositeKey
			EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
			if err != nil {
				return shim.Error(err.Error())
			}

			// === 에스크로 Tx 저장 ===
			err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		   // make index (for remove)
           // EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
            EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo}) //index key chg ssi
			if err != nil {
	            return shim.Error(err.Error())
	        }

			// === delete index ===
			err = stub.DelState(EscrowIndexKey)
			if err != nil {
				return shim.Error("Failed to delete state:" + err.Error())
			}
	
			// make index (for new index)
			//EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
			EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo}) //index key chg ssi
			if err != nil {
				return shim.Error(err.Error())
			}
			
			// === save index ===
			value := []byte{0x00}
		    stub.PutState(EscrowIndexKey, value)
		}
	}
    buffer.WriteString("]")

    fmt.Println("readEscrowStatusByVaNo queryResult : \n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())

}

// ============================================================
// 미처리된 내역을parameter 없이 출력해주는 함수 (koscom만 사용)
// ============================================================
func (t *EscrowChaincode) readEscrowStatusByAll(stub shim.ChaincodeStubInterface) pb.Response {
	var err error
    var buffer bytes.Buffer
    var escrowTxJSON escrowTx
    //EscrowIndexKey로 생성해놓았던 데이터를 조회하기 위한 key 생성
	StatuCodeReadYnResultIterator, err := stub.GetStateByPartialCompositeKey("EscrowIndexKey", []string{"N"})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer StatuCodeReadYnResultIterator.Close()

	// 반복수행하며 해당되는 데이터 set 조회
	var i int
    buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false	
	for i = 0; StatuCodeReadYnResultIterator.HasNext(); i++ {
		responseRange, err := StatuCodeReadYnResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// data Set 하나를 가져와서 각 항목에 저장
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		ReadYn := compositeKeyParts[0]
		StatusCode := compositeKeyParts[1]
        VaNo := compositeKeyParts[2]
        CtrctNo := compositeKeyParts[3]
        /*
        Crcy := compositeKeyParts[4]
        Depositor := compositeKeyParts[5]
		DepositAmt := compositeKeyParts[6]
        DepositOpenDttm := compositeKeyParts[7]
        DepositCloseDttm := compositeKeyParts[8]
        DepositDttm := compositeKeyParts[9]
        TrsfBnkCode := compositeKeyParts[10]
        TrsfBnkAcnt := compositeKeyParts[11]
        TrsfAmt := compositeKeyParts[12]
        TrsfReqDttm := compositeKeyParts[13]
        TrsfDttm := compositeKeyParts[14]
        RfundBnkCode := compositeKeyParts[15]
        RfundBnkAcnt := compositeKeyParts[16]
        RfundAmt := compositeKeyParts[17]
        RfundReqDttm := compositeKeyParts[18]
        RfundDttm := compositeKeyParts[19]
        TrdFeeAmt := compositeKeyParts[20]
        EscrFeeAmt := compositeKeyParts[21]
        AdjBaseDt := compositeKeyParts[22]
        ErrCode := compositeKeyParts[23]
		ErrRsn := compositeKeyParts[24]
        // YYYYMMDDHHmmss
        ChgDttm :=compositeKeyParts[25]
        */

  // ssi change
        EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
        if err != nil {
		    return shim.Error(err.Error())
		}

		escrowAsBytes, err := stub.GetState(EscrowCompositeKey)
		if err != nil {
			return shim.Error("Failed to get escrow Tx: " + err.Error())
		} else if escrowAsBytes == nil {
			fmt.Println("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
			return shim.Error("This escrow tx doesn't exists: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]")
		}

        err = json.Unmarshal([]byte(escrowAsBytes), &escrowTxJSON)
        if err != nil {
	    	jsonResp := "{\"Error\":\"Failed to decode JSON of: VaNo [" + VaNo +"] CtrctNo [" + CtrctNo + "]\"}"
			return shim.Error(jsonResp)																									   }


        // Add a comma before array members, suppress it for the first array member
   		if bArrayMemberAlreadyWritten == true {
   			buffer.WriteString(",")
   		}

   		buffer.WriteString("{\"VaNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.VaNo)
   		buffer.WriteString("\"")

		buffer.WriteString(", \"CtrctNo\":")
   		buffer.WriteString("\"")
   		buffer.WriteString(escrowTxJSON.CtrctNo)
   		buffer.WriteString("\"")

        buffer.WriteString(", \"StatusCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.StatusCode)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Crcy\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.Crcy)
        buffer.WriteString("\"")

		buffer.WriteString(", \"Depositor\":")
		buffer.WriteString("\"")
		buffer.WriteString(escrowTxJSON.Depositor)
		buffer.WriteString("\"")

        buffer.WriteString(", \"DepositAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositOpenDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositOpenDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositCloseDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositCloseDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"DepositDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.DepositDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrsfDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrsfDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundBnkAcnt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundBnkAcnt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundReqDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundReqDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"RfundDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.RfundDttm)
        buffer.WriteString("\"")

        buffer.WriteString(", \"TrdFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.TrdFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"EscrFeeAmt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.EscrFeeAmt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"AdjBaseDt\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.AdjBaseDt)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ReadYn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ReadYn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrCode\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrCode)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ErrRsn\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ErrRsn)
        buffer.WriteString("\"")

        buffer.WriteString(", \"ChgDttm\":")
        buffer.WriteString("\"")
        buffer.WriteString(escrowTxJSON.ChgDttm)
        buffer.WriteString("\"")
        buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true
       
        // 읽기여부(처리여부) == "N" 인 경우 읽기여부를 "Y" 로 변경 후 업뎃처리
        if ReadYn == "N" {
			fmt.Println("change readYn value from N to Y : "+VaNo+","+CtrctNo+","+StatusCode)
            // ==== 에스크로 tx 생성 ====
            //escrowTx := &escrowTx{VaNo, CtrctNo, StatusCode, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, "Y", ErrCode, ErrRsn, ChgDttm}
		  	    escrowTx := &escrowTx{escrowTxJSON.VaNo, escrowTxJSON.CtrctNo, escrowTxJSON.StatusCode, escrowTxJSON.Crcy, escrowTxJSON.Depositor, escrowTxJSON.DepositAmt, escrowTxJSON.DepositOpenDttm, escrowTxJSON.DepositCloseDttm, escrowTxJSON.DepositDttm, escrowTxJSON.TrsfBnkCode, escrowTxJSON.TrsfBnkAcnt, escrowTxJSON.TrsfAmt, escrowTxJSON.TrsfReqDttm, escrowTxJSON.TrsfDttm, escrowTxJSON.RfundBnkCode, escrowTxJSON.RfundBnkAcnt, escrowTxJSON.RfundAmt, escrowTxJSON.RfundReqDttm, escrowTxJSON.RfundDttm, escrowTxJSON.TrdFeeAmt, escrowTxJSON.EscrFeeAmt, escrowTxJSON.AdjBaseDt, "Y", escrowTxJSON.ErrCode, escrowTxJSON.ErrRsn, escrowTxJSON.ChgDttm}

			escrowTxJSONasBytes, err := json.Marshal(escrowTx)
			if err != nil {
				return shim.Error(err.Error())
			}

			// make compositeKey
			EscrowCompositeKey, err := stub.CreateCompositeKey("EscrowKey", []string{VaNo, CtrctNo})
			if err != nil {
				return shim.Error(err.Error())
			}

			// === 에스크로 Tx 저장 ===
			err = stub.PutState(EscrowCompositeKey, escrowTxJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		   // make index (for remove)
           // EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
            EscrowIndexKey, err := stub.CreateCompositeKey("EscrowIndexKey", []string{"N", StatusCode, VaNo, CtrctNo}) //index key chg ssi
			if err != nil {
	            return shim.Error(err.Error())
	        }

			// === delete index ===
			err = stub.DelState(EscrowIndexKey)
			if err != nil {
				return shim.Error("Failed to delete state:" + err.Error())
			}
	
			// make index (for new index)
			//EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo, Crcy, Depositor, DepositAmt, DepositOpenDttm, DepositCloseDttm, DepositDttm, TrsfBnkCode, TrsfBnkAcnt, TrsfAmt, TrsfReqDttm, TrsfDttm, RfundBnkCode, RfundBnkAcnt, RfundAmt, RfundReqDttm, RfundDttm, TrdFeeAmt, EscrFeeAmt, AdjBaseDt, ErrCode, ErrRsn, ChgDttm})
			EscrowIndexKey, err = stub.CreateCompositeKey("EscrowIndexKey", []string{"Y", StatusCode, VaNo, CtrctNo}) //index key chg ssi
			if err != nil {
				return shim.Error(err.Error())
			}
			
			// === save index ===
			value := []byte{0x00}
		    stub.PutState(EscrowIndexKey, value)
		}
	}
    buffer.WriteString("]")

    fmt.Println("readEscrowStatusByAll queryResult : \n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())


}
