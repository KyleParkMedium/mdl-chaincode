package controller

import (
	"fmt"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/wallet"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) CreateWallet(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	// 우선 msp id 로 발급
	address := ccutils.GetAddress([]byte(id))

	// requireParameterFields := []string{wallet.FieldWalletId}
	// err = ccutils.CheckRequireParameter(requireParameterFields, args)
	// if err != nil {
	// 	return ccutils.GenerateErrorResponse(err)
	// }

	// stringParameterFields := []string{wallet.FieldWalletId}
	// err = ccutils.CheckTypeString(stringParameterFields, args)
	// if err != nil {
	// 	return ccutils.GenerateErrorResponse(err)
	// }
	// tokenWalletId := args[wallet.FieldWalletId].(string)

	tokenWalletId := address

	// wallet.PartitionTokens[mintByPartition.Partition][0] = token.PartitionToken{Amount: mintByPartition.Amount}

	check := make(map[string][]token.PartitionToken)
	// check["AA"] = append(check["AA"], token.PartitionToken{Amount: 100})
	// map_slice := map[int][]int{}

	testST := wallet.TokenWallet{}
	testST.TokenWalletId = tokenWalletId
	testST.PartitionTokens = check

	// wallet.TokenWallet{
	// 	TokenWalletId: tokenWalletId,
	// 	// PartitionTokens: check,
	// }

	newWallet, err := wallet.CreateWallet(ctx, testST)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	retData := make(map[string]interface{})
	retData[wallet.FieldWalletId] = newWallet.TokenWalletId

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)
}

func (s *SmartContract) TransferByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldRecipient, token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldRecipient, token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	int64ParameterFields := []string{token.FieldAmount}
	err = ccutils.CheckRequireTypeInt64(int64ParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))
	recipient := args[token.FieldRecipient].(string)
	partition := args[token.FieldPartition].(string)
	amount := int64(args[token.FieldAmount].(float64))

	if amount <= 0 {
		return ccutils.GenerateErrorResponse(err)
		// return false, fmt.Errorf("mint amount must be a positive integer")
	}

	err = _transferByPartition(ctx, owner, recipient, partition, amount)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
		// return false, fmt.Errorf("failed to transfer: %v", err)
	}

	// // retData, err := ccutils.StructToMap(asset)
	// // if err != nil {
	// // 	return ccutils.GenerateErrorResponse(err)
	// // }

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
	// // return true, nil
}

func (s *SmartContract) TransferFromByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldFrom, token.FieldTo, token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldFrom, token.FieldTo, token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	int64ParameterFields := []string{token.FieldAmount}
	err = ccutils.CheckRequireTypeInt64(int64ParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	// owner Address
	spender := ccutils.GetAddress([]byte(id))
	fmt.Println(spender)
	skip := true

	from := args[token.FieldFrom].(string)
	to := args[token.FieldTo].(string)
	partition := args[token.FieldPartition].(string)
	amount := int64(args[token.FieldAmount].(float64))

	args[token.FieldOwner] = from
	args[token.FieldSpender] = spender

	// allowance, allowanceKey, _ := s.Allowance(ctx, from, spender)
	allowance, err := s.AllowanceByPartition(ctx, args, skip)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	allowanceValue := int64(allowance.Data.(float64))

	if allowanceValue < amount {
		return ccutils.GenerateErrorResponse(err)
		// return fmt.Errorf("Allowance is less than value")
	}

	if amount <= 0 {
		return ccutils.GenerateErrorResponse(err)
		// return false, fmt.Errorf("mint amount must be a positive integer")
	}

	// Decrease the allowance
	updatedAllowance := allowanceValue - amount

	err = _approveByPartition(ctx, from, spender, partition, updatedAllowance)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	err = _transferByPartition(ctx, from, to, partition, amount)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
		// return false, fmt.Errorf("failed to transfer: %v", err)
	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
	// // return true, nil
}

func _transferByPartition(ctx contractapi.TransactionContextInterface, from string, to string, partition string, value int64) error {

	transferByPartition := token.TransferByPartitionStruct{}
	transferByPartition.From = from
	transferByPartition.To = to
	transferByPartition.Partition = partition
	transferByPartition.Amount = value

	err := wallet.TransferByPartition(ctx, transferByPartition)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) MintByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	int64ParameterFields := []string{token.FieldAmount}
	err = ccutils.CheckRequireTypeInt64(int64ParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	minter := ccutils.GetAddress([]byte(id))
	partition := args[token.FieldPartition].(string)
	amount := int64(args[token.FieldAmount].(float64))

	if amount <= 0 {
		// return "", err
		// return ccutils.GenerateErrorResponse(err)
		return nil, fmt.Errorf("mint amount must be a positive integer")
	}

	mintByPartition := token.MintByPartitionStruct{Minter: minter, Partition: partition, Amount: amount}

	// fmt.Println(mintByPartition)
	err = wallet.MintByPartition(ctx, mintByPartition)
	if err != nil {
		// return err
		return ccutils.GenerateErrorResponse(err)

	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
	// return "ok", nil
}

func (s *SmartContract) BurnByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	int64ParameterFields := []string{token.FieldAmount}
	err = ccutils.CheckRequireTypeInt64(int64ParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	minter := ccutils.GetAddress([]byte(id))
	partition := args[token.FieldPartition].(string)
	amount := int64(args[token.FieldAmount].(float64))

	if amount <= 0 {
		// return "", err
		// return ccutils.GenerateErrorResponse(err)
		return nil, fmt.Errorf("mint amount must be a positive integer")
	}

	burnByPartition := token.MintByPartitionStruct{Minter: minter, Partition: partition, Amount: amount}

	// fmt.Println(mintByPartition)
	err = wallet.BurnByPartition(ctx, burnByPartition)
	if err != nil {
		// return err
		return ccutils.GenerateErrorResponse(err)

	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
	// return "ok", nil
}
