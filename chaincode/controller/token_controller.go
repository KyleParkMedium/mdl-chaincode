package controller

import (
	"fmt"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/test"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	test.SmartContract
	// chaincode.SmartContract
}

// ERC20 Strandard Code
/**
 * @dev Total number of tokens in existence
 */
func (s *SmartContract) TotalSupply(ctx contractapi.TransactionContextInterface) (*ccutils.Response, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	// log.Printf("TotalSupply: %d tokens", totalSupply)

	totalSupply, err := token.TotalSupply(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	retData, err := ccutils.StructToMap(totalSupply)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)
}

// ERC20 Strandard Code
/**
 * @dev Total number of tokens in existence
 */
func (s *SmartContract) TotalSupplyByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	requireParameterFields := []string{token.FieldPartition}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	partition := args[token.FieldPartition].(string)

	totalSupplyByPartition, err := token.TotalSupplyByPartition(ctx, partition)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	retData, err := ccutils.StructToMap(totalSupplyByPartition)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)
}

func (s *SmartContract) BalanceOfByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldTokenHolder, token.FieldPartition}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldTokenHolder, token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	_tokenHoler := args[token.FieldTokenHolder].(string)
	_partition := args[token.FieldPartition].(string)

	balanceOfByPartition, err := token.BalanceOfByPartition(ctx, _tokenHoler, _partition)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	// int 값은 리턴을 어떻게 해줄건지 고민을 해보자능
	// retData, err := ccutils.StructToMap(balanceOfByPartition)
	// if err != nil {
	// 	return ccutils.GenerateErrorResponse(err)
	// }

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], balanceOfByPartition)
}

func (s *SmartContract) AllowanceByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}, skip bool) (*ccutils.Response, error) {
	if !skip {
		// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
		err := ccutils.GetMSPID(ctx)
		if err != nil {
			return nil, err
		}

		requireParameterFields := []string{token.FieldOwner, token.FieldSpender, token.FieldPartition}
		err = ccutils.CheckRequireParameter(requireParameterFields, args)
		if err != nil {
			return ccutils.GenerateErrorResponse(err)
		}

		stringParameterFields := []string{token.FieldOwner, token.FieldSpender, token.FieldPartition}
		err = ccutils.CheckRequireTypeString(stringParameterFields, args)
		if err != nil {
			return ccutils.GenerateErrorResponse(err)
		}
	}
	owner := args[token.FieldOwner].(string)
	spender := args[token.FieldSpender].(string)
	partition := args[token.FieldPartition].(string)

	allowanceByPartition, err := token.AllowanceByPartition(ctx, owner, spender, partition)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	// retData, err := ccutils.StructToMap(allowanceByPartition)
	// if err != nil {
	// 	return ccutils.GenerateErrorResponse(err)
	// }

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], allowanceByPartition.Amount)

	// if skip {
	// }
	// return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)
}

func (s *SmartContract) ApproveByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	requireParameterFields := []string{token.FieldSpender, token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldSpender, token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	int64ParameterFields := []string{token.FieldAmount}
	err = ccutils.CheckRequireTypeInt64(int64ParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	owner := ccutils.GetAddress([]byte(id))
	spender := args[token.FieldSpender].(string)
	partition := args[token.FieldPartition].(string)
	amount := int64(args[token.FieldAmount].(float64))

	err = _approveByPartition(ctx, owner, spender, partition, amount)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
}

func (s *SmartContract) IncreaseAllowanceByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	requireParameterFields := []string{token.FieldSpender, token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldSpender, token.FieldPartition}
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
	fmt.Println(owner)
	spender := args[token.FieldSpender].(string)
	partition := args[token.FieldPartition].(string)
	addedValue := int64(args[token.FieldAmount].(float64))

	if addedValue <= 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return nil, fmt.Errorf("addValue cannot be negative")
	}

	skip := true
	allowance, err := s.AllowanceByPartition(ctx, args, skip)

	// allowanceValue := allowance.Data.(int64)

	// err = _approveByPartition(ctx, owner, spender, partition, allowanceValue+addedValue)
	// if err != nil {
	// 	return nil, err
	// }

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
}

func (s *SmartContract) DecreaseAllowanceByPartition(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	requireParameterFields := []string{token.FieldSpender, token.FieldPartition, token.FieldAmount}
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldSpender, token.FieldPartition}
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
	spender := args[token.FieldSpender].(string)
	partition := args[token.FieldPartition].(string)
	subtractedValue := int64(args[token.FieldAmount].(float64))

	if subtractedValue <= 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return nil, fmt.Errorf("subtractedValue cannot be negative")
	}

	skip := true
	allowance, err := s.AllowanceByPartition(ctx, args, skip)

	allowanceValue := allowance.Data.(int64)

	if allowanceValue < subtractedValue {
		return ccutils.GenerateErrorResponse(err)
		// return nil, fmt.Errorf("The subtraction is greater than the allowable amount. ERC20: decreased allowance below zero : %v", err)
	}

	err = _approveByPartition(ctx, owner, spender, partition, allowanceValue-subtractedValue)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
		// return nil, err

	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], nil)
}

func _approveByPartition(ctx contractapi.TransactionContextInterface, owner string, spender string, partition string, value int64) error {

	allowanceByPartition := token.AllowanceByPartitionStruct{Owner: owner, Spender: spender, Partition: partition, Amount: value}

	err := token.ApproveByPartition(ctx, allowanceByPartition)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) IssuanceAsset(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (*ccutils.Response, error) {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := ccutils.GetMSPID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := ccutils.GetID(ctx)
	if err != nil {
		return nil, err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// Asset.Name, Asset.Partition
	requireParameterFields := []string{token.FieldPartition}

	// codename.FieldCode
	err = ccutils.CheckRequireParameter(requireParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	stringParameterFields := []string{token.FieldPartition}
	err = ccutils.CheckRequireTypeString(stringParameterFields, args)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	partition := args[token.FieldPartition].(string)

	newToken := token.PartitionToken{Publisher: address, TokenID: partition}

	asset, err := token.IssuanceAsset(ctx, newToken)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	retData, err := ccutils.StructToMap(asset)
	if err != nil {
		return ccutils.GenerateErrorResponse(err)
	}

	return ccutils.GenerateSuccessResponse(ctx.GetStub().GetTxID(), ccutils.ChaincodeSuccess, ccutils.CodeMessage[ccutils.ChaincodeSuccess], retData)

}
