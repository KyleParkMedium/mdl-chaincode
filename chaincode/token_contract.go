package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mdl-chaincode/chaincode/ccutils"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Changelog
const (
	Author           = "Kyle"
	DateCreated      = "2022/10/28"
	ChaincodeName    = "STO TOKEN Standard"
	ChaincodeVersion = "0.0.1"
)

const (
	// Define key names for options
	totalSupplyKey = "totalSupply"
	tokenName      = "mdl"

	// Define objectType names by partition for prefix
	totalSupplyByPartitionPrefix = "totalSupplyByPartition"

	// _allowances
	allowanceByPartitionPrefix = "allowanceByPartition"

	// issue or mint partition token
	mintingByPartitionPrefix = "mintingByPartition"
	clientByPartitionPrefix  = "clientByPartition"

	clientWalletPrefix = "clientWallet"

	operatorForPartitionPrefix = "operatorForPartition"
)

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// event provides an organized struct for emitting events
type Event struct {
	From  string
	To    string
	Value int
}

// Represents a fungible set of tokens.
type TotalSupplyByPartition struct {
	TotalSupply int
	// Partition Address
	Partition string
}

// Represents a fungible set of tokens.

// token
type Token struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Amount int    `json:"amount"`
	Locked bool   `json:"locked"`
}

// address or wallet
type Wallet struct {
	DocType string `json:"docType"`
	Name    string `json:"name"`

	AuthWalletId string `json:"authWalletId"`
	CreatedDate  string `json:"createdDate"`
	UpdatedDate  string `json:"updatedDate"`
	ExpiredDate  string `json:"expiredDate"`

	AA map[string]struct{}

	BB []PartitionToken

	CC interface{}

	// 배열로..?
	PartitionToken Partition
}

type Partition struct {
	Amount int
	// Partition Address
	Partition string
}

// partition Token
type PartitionToken struct {
	DocType string `json:"docType"`

	Name   string `json:"name"`
	ID     string `json:"id"`
	Locked bool   `json:"locked"`

	Publisher string `json:"publisher"`

	ExpiredDate string `json:"expiredDate"`
	CreatedDate string `json:"createdDate"`
	UpdatedDate string `json:"updatedDate"`

	TxId string `json:"txId"`

	// 요기가 fix 될 예정
	Partition Partition `json:"partition"`
}

// ERC20 Strandard Code
/**
 * @dev Total number of tokens in existence
 */
func (s *SmartContract) TotalSupply(ctx contractapi.TransactionContextInterface) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	// Retrieve total supply of tokens from state of smart contract
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	var totalSupply int

	// If no tokens have been minted, return 0
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	log.Printf("TotalSupply: %d tokens", totalSupply)

	return totalSupply, nil
}

// ERC20 Strandard Code
/**
 * @dev Total number of tokens in existence
 */
func (s *SmartContract) TotalSupplyByPartition(ctx contractapi.TransactionContextInterface, partition string) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{partition})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	// Retrieve total supply of tokens from state of smart contract
	totalSupplyByPartitionBytes, err := ctx.GetStub().GetState(totalSupplyByPartitionKey)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	// If no tokens have been minted, return 0
	if totalSupplyByPartitionBytes == nil {
		return 0, nil
	}

	supplyToken := new(TotalSupplyByPartition)
	err = json.Unmarshal(totalSupplyByPartitionBytes, supplyToken)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	return supplyToken.TotalSupply, nil
}

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) BalanceOfByPartition(ctx contractapi.TransactionContextInterface, _tokenHolder string, _partition string) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{_tokenHolder, _partition})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	balanceByPartitionBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	if balanceByPartitionBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", _tokenHolder)
	}

	// // 이런 이더리움의 조건절 함수를 패브릭에서는 어디에서 처리를 해주어야 할까?
	// if _validPartition(result._partition, result.owner) {
	// 	return partitions[owner][partitionToIndex[owner][_partition]-1].amount
	// } else {
	// 	return 0
	// }

	token := new(PartitionToken)
	err = json.Unmarshal(balanceByPartitionBytes, token)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	return token.Partition.Amount, nil
}

/**
 * @dev Function to check the amount of tokens that an owner allowed to a spender.
 * @param owner address The address which owns the funds.
 * @param spender address The address which will spend the funds.
 * @return A uint256 specifying the amount of tokens still available for the spender.
 */
func (s *SmartContract) AllowanceByPartition(ctx contractapi.TransactionContextInterface, owner string, spender string, partition string) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	// Create allowanceKey
	allowancePartitionKey, err := ctx.GetStub().CreateCompositeKey(allowanceByPartitionPrefix, []string{owner, spender, partition})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", allowanceByPartitionPrefix, err)
	}

	// Read the allowance amount from the world state
	allowanceBytes, err := ctx.GetStub().GetState(allowancePartitionKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read allowance for %s from world state: %v", allowancePartitionKey, err)
	}

	var allowance int

	// If no current allowance, set allowance to 0
	if allowanceBytes == nil {
		allowance = 0
	} else {
		allowance, err = strconv.Atoi(string(allowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
		if err != nil {
			return 0, fmt.Errorf("failed to convert string to integer: %v", err)
		}
	}

	return allowance, nil
}

/**
 * @dev Approve the passed address to spend the specified amount of tokens on behalf of msg.sender.
 * Beware that changing an allowance with this method brings the risk that someone may use both the old
 * and the new allowance by unfortunate transaction ordering. One possible solution to mitigate this
 * race condition is to first reduce the spender's allowance to 0 and set the desired value afterwards:
 * https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
 * @param spender The address which will spend the funds.
 * @param value The amount of tokens to be spent.
 */
func (s *SmartContract) ApproveByPartition(ctx contractapi.TransactionContextInterface, spender string, partition string, amount int) (bool, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return false, err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return false, err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	err = _approveByPartition(ctx, owner, spender, partition, amount)
	if err != nil {
		return false, err
	}

	return true, nil
}

/**
 * @dev Increase the amount of tokens that an owner allowed to a spender.
 * approve should be called when allowed_[_spender] == 0. To increment
 * allowed value is better to use this function to avoid 2 calls (and wait until
 * the first transaction is mined)
 * From MonolithDAO Token.sol
 * @param spender The address which will spend the funds.
 * @param addedValue The amount of tokens to increase the allowance by.
 */
func (s *SmartContract) IncreaseAllowanceByPartition(ctx contractapi.TransactionContextInterface, spender string, partition string, addedValue int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	if addedValue <= 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return fmt.Errorf("addValue cannot be negative")
	}

	allowanceValue, err := s.AllowanceByPartition(ctx, owner, spender, partition)

	err = _approveByPartition(ctx, owner, spender, partition, allowanceValue+addedValue)

	// emit Approval(msg.sender, spender, _allowed[msg.sender][spender]);
	return nil
}

/**
 * @dev Decrease the amount of tokens that an owner allowed to a spender.
 * approve should be called when allowed_[_spender] == 0. To decrement
 * allowed value is better to use this function to avoid 2 calls (and wait until
 * the first transaction is mined)
 * From MonolithDAO Token.sol
 * @param spender The address which will spend the funds.
 * @param subtractedValue The amount of tokens to decrease the allowance by.
 */
func (s *SmartContract) DecreaseAllowanceByPartition(ctx contractapi.TransactionContextInterface, spender string, partition string, subtractedValue int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	if subtractedValue <= 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return fmt.Errorf("subtractedValue cannot be negative")
	}

	allowanceValue, err := s.AllowanceByPartition(ctx, owner, spender, partition)

	if allowanceValue < subtractedValue {
		return fmt.Errorf("The subtraction is greater than the allowable amount. ERC20: decreased allowance below zero : %v", err)
	}

	err = _approveByPartition(ctx, owner, spender, partition, allowanceValue-subtractedValue)

	return nil
}

/**
 * @dev Sets `amount` as the allowance of `spender` over the `owner` s tokens.
 *
 * This internal function is equivalent to `approve`, and can be used to
 * e.g. set automatic allowances for certain subsystems, etc.
 *
 * Emits an {Approval} event.
 *
 * Requirements:
 *
 * - `owner` cannot be the zero address.
 * - `spender` cannot be the zero address.
 */
func _approveByPartition(ctx contractapi.TransactionContextInterface, owner string, spender string, partition string, value int) error {

	// Create allowanceKey
	allowancePartitionKey, err := ctx.GetStub().CreateCompositeKey(allowanceByPartitionPrefix, []string{owner, spender, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowanceByPartitionPrefix, err)
	}

	// Update the state of the smart contract by adding the allowanceKey and value
	err = ctx.GetStub().PutState(allowancePartitionKey, []byte(strconv.Itoa(value)))
	if err != nil {
		return fmt.Errorf("failed to update state of smart contract for key %s: %v", allowancePartitionKey, err)
	}

	// Emit the Approval event
	approvalEvent := Event{owner, spender, value}
	approvalEventJSON, err := json.Marshal(approvalEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Approval", approvalEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("client %s approved a withdrawal allowance of %d for spender %s", owner, value, spender)

	return nil
}

/**
 * @dev See {IERC20-transfer}.
 *
 * Requirements:
 *
 * - `recipient` cannot be the zero address.
 * - the caller must have a balance of at least `amount`.
 */
func (s *SmartContract) TransferByPartition(ctx contractapi.TransactionContextInterface, recipient string, partition string, amount int) (bool, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return false, err
	}

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return false, err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	if amount <= 0 {
		return false, fmt.Errorf("mint amount must be a positive integer")
	}

	err = _transferByPartition(ctx, owner, recipient, partition, amount)
	if err != nil {
		return false, fmt.Errorf("failed to transfer: %v", err)
	}

	return true, nil
}

/**
 * @dev Transfer tokens from one address to another
 * @param from address The address which you want to send tokens from
 * @param to address The address which you want to transfer to
 * @param value uint256 the amount of tokens to be transferred
 */
func (s *SmartContract) TransferFromByPartition(ctx contractapi.TransactionContextInterface, from string, to string, partition string, value int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	spender := ccutils.GetAddress([]byte(id))

	// allowance, allowanceKey, _ := s.Allowance(ctx, from, spender)
	allowance, err := s.AllowanceByPartition(ctx, from, spender, partition)
	if err != nil {
		return err
	}

	if allowance < value {
		return fmt.Errorf("Allowance is less than value")
	}

	if value <= 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return fmt.Errorf("transfer amount cannot be negative")
	}

	// Decrease the allowance
	updatedAllowance := allowance - value
	// Create allowanceKey
	allowancePartitionKey, err := ctx.GetStub().CreateCompositeKey(allowanceByPartitionPrefix, []string{from, spender, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowanceByPartitionPrefix, err)
	}
	err = ctx.GetStub().PutState(allowancePartitionKey, []byte(strconv.Itoa(updatedAllowance)))
	if err != nil {
		return err
	}

	// Initiate the transfer
	err = _transferByPartition(ctx, from, to, partition, value)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	return nil
}

/**
 * @dev Transfer token for a specified addresses
 * @param from The address to transfer from.
 * @param to The address to transfer to.
 * @param value The amount to be transferred.
 */
func _transferByPartition(ctx contractapi.TransactionContextInterface, from string, to string, partition string, value int) error {

	// Create allowanceKey
	fromKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{from, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	fromTokenBytes, err := ctx.GetStub().GetState(fromKey)
	if err != nil {
		return fmt.Errorf("failed to read client account %s from world state: %v", from, err)
	}
	if fromTokenBytes == nil {
		return fmt.Errorf("from client account %s has no token", from)
	}

	fromToken := new(PartitionToken)
	err = json.Unmarshal(fromTokenBytes, fromToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}
	fromCurrentBalance := fromToken.Partition.Amount

	if fromCurrentBalance < value {
		return fmt.Errorf("client account %s has insufficient funds", from)
	}

	// Create allowanceKey
	toKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{to, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}
	toTokenBytes, err := ctx.GetStub().GetState(toKey)
	if err != nil {
		return fmt.Errorf("failed to read recipient account %s from world state: %v", to, err)
	}
	if toTokenBytes == nil {
		return fmt.Errorf("to client account %s has no token", from)
	}

	toToken := new(PartitionToken)
	err = json.Unmarshal(toTokenBytes, toToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}
	toCurrentBalance := toToken.Partition.Amount

	// Math
	fromUpdatedBalance := fromCurrentBalance - value
	fromToken.Partition.Amount = fromUpdatedBalance
	fromTokenJSON, err := json.Marshal(fromToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(fromKey, fromTokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}
	toUpdatedBalance := toCurrentBalance + value
	toToken.Partition.Amount = toUpdatedBalance
	toTokenJSON, err := json.Marshal(toToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(toKey, toTokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Emit the Transfer event
	transferEvent := Event{from, to, value}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("client %s balance updated from %s to %s", from, string(fromTokenBytes), string(fromTokenJSON))
	log.Printf("recipient %s balance updated from %s to %s", to, string(toTokenBytes), string(toTokenJSON))

	return nil
}

/** 체인코드 init 위해 임시로 코드 작성
 */
func (s *SmartContract) IsInit(ctx contractapi.TransactionContextInterface) error {

	log.Printf("Initial Isinit run")

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	// Initial Isinit run
	err = ctx.GetStub().PutState("Isinit", []byte("Isinit"))
	if err != nil {
		return err
	}

	return nil
}

/** org1, org2, 피어(관리자, 클라이언트) 노드 주소 생성
 */
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	return owner, nil
}

/** 호출자 Address(shim / ctx.GetClientIdentity().GetID() 모듈화)
 */
func _msgSender(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client id: %v", err)
	}

	return id, nil
}

/** shim / ctx.GetClientIdentity().GeMSPID() 모듈화
 */
func _getMSPID(ctx contractapi.TransactionContextInterface) error {

	// Get ID of submitting client identity
	_, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	return nil
}

func (s *SmartContract) IssuanceAsset(ctx contractapi.TransactionContextInterface, partition string) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// Generate random bytes of assets
	partitionAddress := ccutils.GetAddress([]byte(partition))

	// minter := string(minterBytes)
	// if minter != ccutils.GetAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	// Create Asset
	mintingByPartitionKey, err := ctx.GetStub().CreateCompositeKey(mintingByPartitionPrefix, []string{address, partitionAddress})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", mintingByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(mintingByPartitionKey)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
	}

	if tokenBytes != nil {
		return "", fmt.Errorf("The asset is already registered : %s, %v", partitionAddress, err)
	}

	example := Partition{Amount: 0, Partition: partitionAddress}
	token := PartitionToken{Name: "tokenName", ID: address, Locked: false, Partition: example}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(mintingByPartitionKey, tokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("The Asset %s is registered", partitionAddress)

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{partitionAddress})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	supplyToken := TotalSupplyByPartition{TotalSupply: 0, Partition: partitionAddress}
	supplyTokenJSON, err := json.Marshal(supplyToken)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(totalSupplyByPartitionKey, supplyTokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	return partitionAddress, nil
}

// Client Account Initialize
func (s *SmartContract) ClientByPartition(ctx contractapi.TransactionContextInterface, partition string) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// value, found, err := ctx.GetClientIdentity().GetAttributeValue("roles")

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{address, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	clientAmountBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account from world state: %v", err)
	}
	if clientAmountBytes != nil {
		return fmt.Errorf("client %s has been already registered: %v", address, err)
	}

	// example := Partition{Amount: 0, Partition: partition}
	token := PartitionToken{Name: "token"}
	// token := PartitionToken{"token", "tokenName", address, false, example}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(clientKey, tokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("client account %s registered", address)

	return nil
}

/** @dev Creates `amount` tokens and assigns them to `account`, increasing
 * the total supply.
 *
 * Emits a {Transfer} event with `from` set to the zero address.
 *
 * Requirements:
 *
 * - `account` cannot be the zero address.
 */
func (s *SmartContract) MintByPartition(ctx contractapi.TransactionContextInterface, partition string, amount int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// minterBytes, err := ctx.GetStub().GetState(minterKey)
	// if err != nil {
	// 	return fmt.Errorf("failed to read minter account from world state: %v", err)
	// }

	// minter := string(minterBytes)
	// if minter != ccutils.GetAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{address, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
	}
	if tokenBytes == nil {
		return fmt.Errorf("The information of the calling client does not exist.")
	}

	token := new(PartitionToken)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	// if clientMSPID != token.MSPID {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	currentBalance := token.Partition.Amount

	updatedBalance := currentBalance + amount

	token.Partition.Amount = updatedBalance

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(clientKey, tokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Update the totalSupply, totalSupplyByPartition
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	var totalSupply int

	// If no tokens have been minted, initialize the totalSupply
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	// Add the mint amount to the total supply and update the state
	totalSupply += amount
	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return err
	}

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	totalSupplyByPartitionBytes, err := ctx.GetStub().GetState(totalSupplyByPartitionKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	supplyToken := new(TotalSupplyByPartition)
	err = json.Unmarshal(totalSupplyByPartitionBytes, supplyToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	totalSupplyByPartition := supplyToken.TotalSupply
	updateTotalSupply := totalSupplyByPartition + amount
	supplyToken.TotalSupply = updateTotalSupply

	supplyTokenJSON, err := json.Marshal(supplyToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(totalSupplyByPartitionKey, supplyTokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Emit the Transfer event
	transferEvent := Event{"0x0", address, amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("minter account %s balance updated from %s to %s", address, string(tokenBytes), string(tokenJSON))

	return nil
}

// Burn redeems tokens the minter's account balance
// This function triggers a Transfer event
func (s *SmartContract) BurnByPartition(ctx contractapi.TransactionContextInterface, partition string, amount int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{address, partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
	}
	// Check if minter current balance exists
	if tokenBytes == nil {
		return errors.New("the balance does not exist")
	}

	token := new(PartitionToken)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	// if clientMSPID != token.MSPID {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	currentBalance := token.Partition.Amount

	if currentBalance < amount {
		return fmt.Errorf("minter does not have enough balance for burn")
	}

	updatedBalance := currentBalance - amount

	token.Partition.Amount = updatedBalance

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

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	totalSupplyByPartitionBytes, err := ctx.GetStub().GetState(totalSupplyByPartitionKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	supplyToken := new(TotalSupplyByPartition)
	err = json.Unmarshal(totalSupplyByPartitionBytes, supplyToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	totalSupplyByPartition := supplyToken.TotalSupply
	updateTotalSupply := totalSupplyByPartition - amount
	supplyToken.TotalSupply = updateTotalSupply

	supplyTokenJSON, err := json.Marshal(supplyToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(totalSupplyByPartitionKey, supplyTokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Emit the Transfer event
	transferEvent := Event{address, "0x0", amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("minter account %s balance updated from %s to %s", address, string(tokenBytes), string(tokenJSON))

	return nil
}

// ClientAccountID returns the id of the requesting client's account
// In this implementation, the client account ID is the clientId itself
// Users can use this function to get their own account id, which they can then give to others as the payment address
func (s *SmartContract) ClientAccountID(ctx contractapi.TransactionContextInterface) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	clientAccountID := ccutils.GetAddress([]byte(id))

	return clientAccountID, nil
}

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) ClientAccountBalanceByPartition(ctx contractapi.TransactionContextInterface, _partition string) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return 0, err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	// Create allowanceKey
	clientPartitionKey, err := ctx.GetStub().CreateCompositeKey(clientByPartitionPrefix, []string{owner, _partition})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientPartitionKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if tokenBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", owner)
	}

	// if _validPartition(result._partition, result.owner) {
	// 	return partitions[owner][partitionToIndex[owner][_partition]-1].amount
	// } else {
	// 	return 0
	// }

	token := new(PartitionToken)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	return token.Partition.Amount, nil
}
