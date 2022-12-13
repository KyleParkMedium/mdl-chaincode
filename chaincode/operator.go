package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Operator Information

// @notice Determines whether `_operator` is an operator for a specified partition of `_tokenHolder`
func IsOperatorForPartition(ctx contractapi.TransactionContextInterface, _partition string, _operator string, _tokenHolder string) (bool, error) {

	operatorForPartitionKey, err := ctx.GetStub().CreateCompositeKey(operatorForPartitionPrefix, []string{_partition, _operator, _tokenHolder})
	if err != nil {
		return false, fmt.Errorf("failed to create the composite key for prefix %s: %v", operatorForPartitionPrefix, err)
	}

	boolBytes, err := ctx.GetStub().GetState(operatorForPartitionKey)
	if err != nil {
		return false, fmt.Errorf("failed to read operator %s for partition %s of tokenHolder %s: %v", _operator, _partition, _tokenHolder, err)
	}
	if boolBytes != nil {
		return false, fmt.Errorf("the operator %s for partition %s of tokenHolder %s does not exist.", _operator, _partition, _tokenHolder)
	}

	// bytes -> bool 변환 찾아 봐야함
	// bool2, err := strconv.Atob(string(boolBytes))
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to convert string to integer: %v", err)
	// }

	return true, nil
}

// func isOperator(ctx contractapi.TransactionContextInterface, _operator string, _tokenHolder string) bool {}

// // Operator Management
// func AuthorizeOperatorByPartition(ctx contractapi.TransactionContextInterface, _partition string, _operator string) {

// 	operatorForPartitionKey, err := ctx.GetStub().CreateCompositeKey(operatorForPartitionPrefix, []string{_partition, _operator, _tokenHolder})
// 	if err != nil {
// 		return false, fmt.Errorf("failed to create the composite key for prefix %s: %v", mintingByPartitionPrefix, err)
// 	}

// 	boolBytes, err := ctx.GetStub().GetState(operatorForPartitionKey)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to read operator %s for partition %s of tokenHolder %s: %v", _operator, _partition, _tokenHolder, err)
// 	}
// 	if boolBytes != nil {
// 		return false, fmt.Errorf("the operator %s for partition %s of tokenHolder %s does not exist.", _operator, _partition, _tokenHolder)
// 	}

// }

// func authorizeOperator(_operator string) {}

func revokeOperator(_operator string) {}

func revokeOperatorByPartition(_partition string, _operator string) {}
