package test

import (
	"encoding/json"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ledgermanager"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/wallet"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	chaincode.SmartContract
}

func (s *SmartContract) Map(ctx contractapi.TransactionContextInterface) (string, error) {

	x := map[string][]token.PartitionToken{}
	b := make([]token.PartitionToken, 1)
	b[0].Amount = 10
	x["QQ"] = b

	// a := wallet.TokenWallet{TokenWalletId: "Check", PartitionTokens: x}
	a := wallet.TW{Name: "hihi"}

	ttt := "AA"

	key, err := ledgermanager.PutState(wallet.DocType_TokenWallet, ttt, a, ctx)
	if err != nil {
		return "", err
	}

	// // key 어떻게 할건지
	// fromBytes, err := ledgermanager.GetState(wallet.DocType_TokenWallet, ttt, ctx)
	// if err != nil {
	// 	return "", err
	// }

	return key, nil
}

func (s *SmartContract) Map2(ctx contractapi.TransactionContextInterface) (*ccutils.Response, error) {
	ttt := "AA"

	// key 어떻게 할건지
	fromBytes, err := ledgermanager.GetState(wallet.DocType_TokenWallet, ttt, ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	trade := wallet.TW{}
	if err := json.Unmarshal(fromBytes, &trade); err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	// qq := wallet.TokenWallet{}
	// err = json.Unmarshal(abcd, qq)

	// abcdd := qq.PartitionTokens["QQ"][0]

	// qwe, err := json.Marshal(qq.PartitionTokens)

	// ww := map[string]token.PartitionToken{}
	// ww := make(map[string]interface{})
	// err = json.Unmarshal(qwe, &ww)

	// fmt.Println(ww["QQ"].Amount)

	// fmt.Println(qq.PartitionTokens["hi"].Amount)

	// return ww["QQ"], nil

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], trade)

}
