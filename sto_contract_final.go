package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) balanceOfByPartition(ctx contractapi.TransactionContextInterface, _partition string, _tokenHolder string) (int, error) {

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionPrefix, []string{_partition, _tokenHolder})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionPrefix, err)
	}

	balanceByPartitionBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	if balanceByPartitionBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", _tokenHolder)
	}

	// // 이런 조건절 함수를 패브릭에서는 어디에서 처리를 해주어야 할까?
	// if _validPartition(result._partition, result.owner) {
	// 	return partitions[owner][partitionToIndex[owner][_partition]-1].amount
	// } else {
	// 	return 0
	// }

	balance, _ := strconv.Atoi(string(balanceByPartitionBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

	return balance, nil
	// return balances[_tokenHolder]
}

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) ClientAccountBalanceByPartition(ctx contractapi.TransactionContextInterface, _partition string) (int, error) {

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return 0, fmt.Errorf("failed to get client id: %v", err)
	}

	owner := getAddress([]byte(id))

	// Create allowanceKey
	clientPartitionKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionPrefix, []string{_partition, owner})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientPartitionKey)

	result, err := json.Marshal(tokenBytes)

	// if _validPartition(result._partition, result.owner) {
	// 	return partitions[owner][partitionToIndex[owner][_partition]-1].amount
	// } else {
	// 	return 0
	// }

	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if tokenBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", owner)
	}

	token := new(PartitionToken)
	err = json.Unmarshal(tokenBytes, token)

	if err != nil {
		return 0, fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	return PartitionToken.Amount, nil
}
