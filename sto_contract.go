package chaincode

// 내부 함수 : transfer mint burn approve spendallowance before,afterTokenTransfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) _validPartition(ctx contractapi.TransactionContextInterface, _partition string, _holder string) bool {
	if partitions[_holder].length < partitionToIndex[_holder][_partition] || partitionToIndex[_holder][_partition] == 0 {
		return false
	} else {
		return true
	}
}

func (s *SmartContract) _validPartitionForReceiver(ctx contractapi.TransactionContextInterface, bytes32 _partition, address _to) bool {
	for i := 0; i < partitions[_to].length; i++ {
		if partitions[_to][i].partition == _partition {
			return true
		}
	}
	return false
}

func (s *SmartContract) _validateParams(ctx contractapi.TransactionContextInterface, _partition string, _value int) bool {
	if _value != uint256(0) {
		// "Zero value not allowed"
		return false
	} else if _partition != bytes32(0) {
		// "Invalid partition"
		return false
	} else {
		return true
	}
}

// / @notice Increases totalSupply and the corresponding amount of the specified owners partition
// / @param _partition The partition to allocate the increase in balance
// / @param _tokenHolder The token holder whose balance should be increased
// / @param _value The amount by which to increase the balance
// / @param _data Additional data attached to the minting of tokens
func (s *SmartContract) issueByPartition(ctx contractapi.TransactionContextInterface, _partition string, _tokenHolder string, _value int, _data string) (int, error) {
	// onlyOwner 고려해야함

	// // Add the function to validate the `_data` parameter
	// _validateParams(_partition, _value)
	// if _tokenHolder != address(0) {
	// 	return false
	// 	//"Invalid token receiver"
	// }

	// Create allowanceKey
	clientPartitionKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionPrefix, []string{_partition, _tokenHolder})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionPrefix, err)
	}

	// Read the allowance amount from the world state
	partitionBytes, err := ctx.GetStub().GetState(clientPartitionKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read allowance for %s from world state: %v", partitionKey, err)
	}

	partitionValue = strconv.Atoi(string(partitionBytes))

	updatedBalance := partitionValue + _value

	err = ctx.GetStub().PutState(clientPartitionKey, updatedBalance)

	// var index int = partitionToIndex[_tokenHolder][_partition]
	// if index == 0 {
	// 	partitions[_tokenHolder].push(Partition(_value, _partition))
	// 	partitionToIndex[_tokenHolder][_partition] = partitions[_tokenHolder].length
	// } else {
	// 	partitions[_tokenHolder][index-1].amount = partitions[_tokenHolder][index-1].amount.add(_value)
	// }

	// _totalSupply = _totalSupply.add(_value)
	// balances[_tokenHolder] = balances[_tokenHolder].add(_value)
	// Mint(30) 요거 들어가야함

	// emit IssuedByPartition(_partition, _tokenHolder, _value, _data);
}

// Burn redeems tokens the minter's account balance
// This function triggers a Transfer event
func (s *SmartContract) Burn(ctx contractapi.TransactionContextInterface, amount int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != "Org1MSP" {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	// address := getAddress([]byte(id))
	// minterKey2, err := ctx.GetStub().CreateCompositeKey(minterKeyPrefix, []string{address})
	// minterBytes, err := ctx.GetStub().GetState(minterKey2)

	minterBytes, err := ctx.GetStub().GetState(minterKeyPrefix)
	if err != nil {
		return fmt.Errorf("failed to read minter account from world state: %v", err)
	}

	minter := string(minterBytes)
	if minter != getAddress([]byte(id)) {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{minter})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
	}
	// Check if minter current balance exists
	if tokenBytes == nil {
		return errors.New("the balance does not exist")
	}

	token := new(Token)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	if clientMSPID != token.MSPID {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	currentBalance := token.Amount

	if currentBalance < amount {
		return fmt.Errorf("minter does not have enough balance for burn")
	}

	updatedBalance := currentBalance - amount

	token.Amount = updatedBalance

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(clientKey, tokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Update the totalSupply
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	// If no tokens have been minted, throw error
	if totalSupplyBytes == nil {
		return errors.New("totalSupply does not exist")
	}

	totalSupply, _ := strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.

	// Subtract the burn amount to the total supply and update the state
	totalSupply -= amount
	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return err
	}

	// Emit the Transfer event
	transferEvent := Event{minter, "0x0", amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("minter account %s balance updated from %s to %s", minter, string(tokenBytes), string(tokenJSON))

	return nil

}

// transfer 여기만 구현 하면 된다.
// approve 등을 연동시키는 과정을 좀 볼딩처리해서 보기 쉽게 하자
func (s *SmartContract) TransferByPartition(ctx contractapi.TransactionContextInterface, _from string, _to string, _value int, _partition string, _data string, _operator string, _operatorData string) {
	if !_validPartition(_partition, _from) {
		return 0, fmt.Errorf("Invalid partition %v", err)
	}
	if partitions[_from][partitionToIndex[_from][_partition]-1].amount >= _value {
		return 0, fmt.Errorf("Insufficient balance %v", err)
	}
	if _to != address(0) {
		return 0, fmt.Errorf("0x address not allowed %v", err)
	}

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	spender := getAddress([]byte(id))

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{_from, spender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Retrieve the allowance of the spender
	currentAllowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve the allowance for %s from world state: %v", allowanceKey, err)
	}

	var currentAllowance int
	currentAllowance, _ = strconv.Atoi(string(currentAllowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.

	// Check if transferred value is less than allowance
	if currentAllowance < value {
		return fmt.Errorf("spender does not have enough allowance for transfer")
	}

	// from, to 계정 파티션 인덱스 구하기
	fromClientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionIndexPrefix, []string{_partition, _from})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionIndexPrefix, err)
	}
	frompartitionIndexBytes, err := ctx.GetStub().GetState(fromClientKey)

	toClientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionIndexPrefix, []string{_partition, _to})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionIndexPrefix, err)
	}
	topartitionIndexBytes, err := ctx.GetStub().GetState(toClientKey)

	fromToken := new(PartitionToken)
	err = json.Unmarshal(frompartitionIndexBytes, fromToken)

	toToken := new(PartitionToken)
	err = json.Unmarshal(topartitionIndexBytes, toToken)

	fromCurrentBalance := fromToken.Amount
	toCurrentBalance := toToken.Amount

	fromUpdatedBalance := fromCurrentBalance - _value
	fromToken.Amount = fromUpdatedBalance
	fromTokenJSON, err := json.Marshal(fromToken)

	toUpdatedBalance := toCurrentBalance + value
	toToken.Amount = toUpdatedBalance
	toTokenJSON, err := json.Marshal(toToken)

	err = ctx.GetStub().PutState(fromClientKey, fromTokenJSON)
	err = ctx.GetStub().PutState(toClientKey, toTokenJSON)

	// Decrease the allowance
	updatedAllowance := currentAllowance - _value
	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(updatedAllowance)))
	if err != nil {
		return err
	}

	// // from, to 계정 파티션 인덱스 구하기
	// fromClientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionIndexPrefix, []string{_partition, _from})
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionIndexPrefix, err)
	// }
	// frompartitionIndexBytes, err := ctx.GetStub().GetState(fromClientKey)

	// toClientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionIndexPrefix, []string{_partition, _to})
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionIndexPrefix, err)
	// }
	// topartitionIndexBytes, err := ctx.GetStub().GetState(toClientKey)

	// pfI, _ := strconv.Atoi(string(frompartitionIndexBytes))
	// ptI, _ := strconv.Atoi(string(topartitionIndexBytes))

	// var partitions = map[string]Partition{}

	// // partitions[_from]
	// // // ex
	// // partitions["address1"] = Partition{Amount: 51, Partition: "bb"}

	// // // 얘는 나중에 고려해도댐
	// // var partitionsToIndex = map[string]map[string]int{}

	// // clientPartitionPrefix + 주소 + 자산 주소 + 인덱스로

	// // var _fromIndex int = partitionToIndex[_from][_partition] - 1
	// var _fromIndex int = pfI

	// if !_validPartitionForReceiver(_partition, _to) {
	// 	partitions[_to].push(Partition(0, _partition))
	// 	partitionToIndex[_to][_partition] = partitions[_to].length
	// }
	// // var _toIndex int = partitionToIndex[_to][_partition] - 1
	// var _toIndex int = ptI

	// // Changing the state values
	// partitions[_from][_fromIndex].amount = partitions[_from][_fromIndex].amount.sub(_value)
	// balances[_from] = balances[_from].sub(_value)
	// partitions[_to][_toIndex].amount = partitions[_to][_toIndex].amount.add(_value)
	// balances[_to] = balances[_to].add(_value)

	// Emit the Transfer event
	// emit TransferByPartition(_partition, _operator, _from, _to, _value, _data, _operatorData);
	transferEvent := Event{"0x0", minter, amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}
}

func (s *SmartContract) _redeemByPartition(ctx contractapi.TransactionContextInterface, _from string, _operator string, _value int, _data string, _operatorData string) {
	// Add the function to validate the `_data` parameter
	_validateParams(_partition, _value)

	if _validPartition(_partition, _from) {
		// "Invalid partition"
	}
	var index int = partitionToIndex[_from][_partition] - 1

	if partitions[_from][index].amount >= _value {
		//  "Insufficient value"
	}

	if partitions[_from][index].amount == _value {
		_deletePartitionForHolder(_from, _partition, index)
	} else {
		partitions[_from][index].amount = partitions[_from][index].amount.sub(_value)
	}
	balances[_from] = balances[_from].sub(_value)
	_totalSupply = _totalSupply.sub(_value)
	// emit RedeemedByPartition(_partition, _operator, _from, _value, _data, _operatorData);
}

func (s *SmartContract) _deletePartitionForHolder(ctx contractapi.TransactionContextInterface, _holder string, _partition string, index int) {
	if index != partitions[_holder].length-1 {
		partitions[_holder][index] = partitions[_holder][partitions[_holder].length-1]
		partitionToIndex[_holder][partitions[_holder][index].partition] = index + 1
	}
	// delete 구현 필
	// delete partitionToIndex[_holder][_partition];
	partitions[_holder].length--
}

// / @notice The standard provides an on-chain function to determine whether a transfer will succeed,
// / and return details indicating the reason if the transfer is not valid.
// / @param _from The address from whom the tokens get transferred.
// / @param _to The address to which to transfer tokens to.
// / @param _partition The partition from which to transfer tokens
// / @param _value The amount of tokens to transfer from `_partition`
// / @param _data Additional data attached to the transfer of tokens
// / @return ESC (Ethereum Status Code) following the EIP-1066 standard
// / @return Application specific reason codes with additional details
// / @return The partition to which the transferred tokens were allocated for the _to address
func (s *SmartContract) canTransferByPartition(ctx contractapi.TransactionContextInterface, _from string, _to string, _partition string, _value int, _data string) (string, string, string) {

	// TODO: Applied the check over the `_data` parameter
	if !_validPartition(_partition, _from) {
		return 0x50, "Partition not exists", bytes32("")
	} else if partitions[_from][partitionToIndex[_from][_partition]].amount < _value {
		return 0x52, "Insufficent balance", bytes32("")
	} else if _to == address(0) {
		return 0x57, "Invalid receiver", bytes32("")
	} else if !KindMath.checkSub(balances[_from], _value) || !KindMath.checkAdd(balances[_to], _value) {
		return 0x50, "Overflow", bytes32("")
	}

	// Call function to get the receiver's partition. For current implementation returning the same as sender's
	return 0x51, "Success", _partition
}

// // Transfer Validity
// func (s *SmartContract) canTransfer(ctx contractapi.TransactionContextInterface) error {
// 	// 호출자 클라이언트, 받는 사람, 수량,
// }
// func (s *SmartContract) canTransferFrom(ctx contractapi.TransactionContextInterface) error {
// // , address _to, uint256 _value, bytes _data) external view returns (bool, byte, bytes32);
// }
