package controller

import (
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/airdrop"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/wallet"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) AirDrop(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {

	// dataSet := airdop.AirDrop{}

	// wallet.PartitionTokens[mintByPartition.Partition][0] = token.PartitionToken{Amount: mintByPartition.Amount}

	check := make(map[string][]token.PartitionToken)
	check["AA"] = append(check["AA"], token.PartitionToken{Amount: 100})
	// map_slice := map[int][]int{}

	testST := wallet.TokenWallet{}
	// testST.TokenWalletId = tokenWalletId
	testST.PartitionTokens = check

	newWallet, err := airdrop.AirDrop(ctx, testST)

	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	retData := make(map[string]interface{})
	retData[wallet.FieldWalletId] = newWallet.TokenWalletId

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)
}
