package chaincode

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"golang.org/x/crypto/sha3"
)

// 현 컨트랙트는 클라이언트만(peer-user1, user2)을 고려하여 작성되어 있음. 관리자(peer-Admin) 에 대한 코드는 추후에 작성 예정.

const (
	// Define key names for options
	totalSupplyKey = "totalSupply"
	tokenName      = "mdl"

	// Define objectType names for prefix
	minterKeyPrefix = "minter"
	allowancePrefix = "allowance"

	// Define client, 자산 분할
	clientPrefix               = "client"
	clientPartitionPrefix      = "clientPartition"
	clientPartitionIndexPrefix = "clientPartitionIndex"

	// AddressLength is the expected length of the address
	addressLength = 20
)

// Errors
var (
	ErrEmptyString   = &decError{"empty hex string"}
	ErrMissingPrefix = &decError{"hex string without 0x prefix"}
	ErrUint64Range   = &decError{"hex number > 64 bits"}
	ErrSyntax        = &decError{"invalid hex string"}
	ErrOddLength     = &decError{"hex string of odd length"}
)

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

// Address represents the 20 byte address of an Ethereum account.
type Address [addressLength]byte

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
type Partition struct {
	Amount int
	// Partition Address
	Partition string
	// 여기 파티션이 주소 일단 스트링으로 적음.
}

// token
type Token struct {
	Name   string `json:"name"`
	MSPID  string `json:"mspid"`
	Amount int    `json:"amount"`
	Locked bool   `json:"locked"`
}

// partition Token
type PartitionToken struct {
	Name      string    `json:"name"`
	MSPID     string    `json:"mspid"`
	Locked    bool      `json:"locked"`
	Partition Partition `json:"partition"`
}

// uint256 _totalSupply;

func getAddress(id []byte) string {
	return Encode(bytesToAddress(Keccak256([]byte(id))[12:]).Bytes())
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func mapError(err error) error {
	if err, ok := err.(*strconv.NumError); ok {
		switch err.Err {
		case strconv.ErrRange:
			return ErrUint64Range
		case strconv.ErrSyntax:
			return ErrSyntax
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return ErrSyntax
	}
	if err == hex.ErrLength {
		return ErrOddLength
	}
	return err
}

// Decode decodes a hex string with 0x prefix.
func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, ErrEmptyString
	}
	if !has0xPrefix(input) {
		return nil, ErrMissingPrefix
	}
	b, err := hex.DecodeString(input[2:])
	if err != nil {
		err = mapError(err)
	}
	return b, err
}

// Encode encodes b as a hex string with 0x prefix.
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}

// KeccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// NewKeccakState creates a new KeccakState
func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func bytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-addressLength:]
	}
	copy(a[addressLength-len(b):], b)
}

// Bytes gets the string representation of the underlying address.
func (a Address) Bytes() []byte { return a[:] }

// TotalSupply returns the total token supply
// 완료
func (s *SmartContract) TotalSupply(ctx contractapi.TransactionContextInterface) (int, error) {

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

// Allowance returns the amount still available for the spender to withdraw from the owner
func (s *SmartContract) Allowance(ctx contractapi.TransactionContextInterface, owner string, spender string) (int, string, error) {

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
	if err != nil {
		return 0, "", fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Read the allowance amount from the world state
	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read allowance for %s from world state: %v", allowanceKey, err)
	}

	var allowance int

	// If no current allowance, set allowance to 0
	if allowanceBytes == nil {
		allowance = 0
	} else {
		allowance, err = strconv.Atoi(string(allowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
		if err != nil {
			return 0, "", fmt.Errorf("failed to convert string to integer: %v", err)
		}
	}

	return allowance, allowanceKey, nil
}

// @notice Counts the sum of all partitions balances assigned to an owner
// @param _tokenHolder An address for whom to query the balance
// @return The number of tokens owned by `_tokenHolder`, possibly zero
func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, _tokenHolder string) (int, error) {
	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{_tokenHolder})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	balanceBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	if balanceBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", _tokenHolder)
	}

	balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

	return balance, nil
	// return balances[_tokenHolder]
}

func (s *SmartContract) increaseAllowance(ctx contractapi.TransactionContextInterface, spender string, addedValue int) error {
	// require(spender != address(0));

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	owner := getAddress([]byte(id))

	allowanceValue, _, err := s.Allowance(ctx, owner, spender)

	err = _approve(ctx, owner, spender, allowanceValue+addedValue)

	// emit Approval(msg.sender, spender, _allowed[msg.sender][spender]);
	return nil
}

func (s *SmartContract) DecreaseAllowance(ctx contractapi.TransactionContextInterface, spender string, subtractedValue int) error {

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	owner := getAddress([]byte(id))

	allowanceValue, _, err := s.Allowance(ctx, owner, spender)

	if allowanceValue < subtractedValue {
		return fmt.Errorf("허용량 보다 뺄셈이 많음 ERC20: decreased allowance below zero", err)
	}

	err = _approve(ctx, owner, spender, allowanceValue-subtractedValue)

	return nil
}

// 운영자..?
// 완료
func _approve(ctx contractapi.TransactionContextInterface, owner string, spender string, value int) error {

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Update the state of the smart contract by adding the allowanceKey and value
	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(value)))
	if err != nil {
		return fmt.Errorf("failed to update state of smart contract for key %s: %v", allowanceKey, err)
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

// Transfer transfers tokens from client account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a Transfer event
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, recipient string, amount int) (bool, error) {

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return false, err
	}

	// owner Address
	owner := getAddress([]byte(id))

	if amount <= 0 {
		return false, fmt.Errorf("mint amount must be a positive integer")
	}

	err = _transfer(ctx, owner, recipient, amount)
	if err != nil {
		return false, fmt.Errorf("failed to transfer: %v", err)
	}

	return true, nil
}

// TransferFrom transfers the value amount from the "from" address to the "to" address
// This function triggers a Transfer event
func (s *SmartContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, value int) error {

	// Get ID of submitting client identity
	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	spender := getAddress([]byte(id))

	allowance, allowanceKey, _ := s.Allowance(ctx, from, spender)

	err = _spendAllowance(ctx, from, spender, value)
	if err != nil {
		return err
	}

	// Initiate the transfer
	err = _transfer(ctx, from, to, value)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	err = _afterTokenTransfer(ctx, allowance, allowanceKey, from, to, value)
	if err != nil {
		return err
	}

	return nil
}

// ClientAccountID returns the id of the requesting client's account
// In this implementation, the client account ID is the clientId itself
// Users can use this function to get their own account id, which they can then give to others as the payment address
func (s *SmartContract) ClientAccountID(ctx contractapi.TransactionContextInterface) (string, error) {

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	clientAccountID := getAddress([]byte(id))

	return clientAccountID, nil
}

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) (int, error) {

	id, err := _msgSender(ctx)
	if err != nil {
		return 0, err
	}

	// owner Address
	owner := getAddress([]byte(id))

	// Create allowanceKey
	ownerKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{owner})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(ownerKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if tokenBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", owner)
	}

	token := new(Token)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	return token.Amount, nil
}

// Helper Functions
// _transfer is a helper function that transfers tokens from the "from" address to the "to" address
// Dependant functions include Transfer and TransferFrom
func _transfer(ctx contractapi.TransactionContextInterface, from string, to string, value int) error {

	if value < 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return fmt.Errorf("transfer amount cannot be negative")
	}

	// Create allowanceKey
	fromKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{from})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	fromTokenBytes, err := ctx.GetStub().GetState(fromKey)
	if err != nil {
		return fmt.Errorf("failed to read client account %s from world state: %v", from, err)
	}
	if fromTokenBytes == nil {
		return fmt.Errorf("from client account %s has no token", from)
	}

	fromToken := new(Token)
	err = json.Unmarshal(fromTokenBytes, fromToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}
	fromCurrentBalance := fromToken.Amount

	if fromCurrentBalance < value {
		return fmt.Errorf("client account %s has insufficient funds", from)
	}

	// Create allowanceKey
	toKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{to})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}
	toTokenBytes, err := ctx.GetStub().GetState(toKey)
	if err != nil {
		return fmt.Errorf("failed to read recipient account %s from world state: %v", to, err)
	}
	if toTokenBytes == nil {
		return fmt.Errorf("to client account %s has no token", from)
	}

	toToken := new(Token)
	err = json.Unmarshal(toTokenBytes, toToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}
	toCurrentBalance := toToken.Amount

	// Math
	fromUpdatedBalance := fromCurrentBalance - value
	fromToken.Amount = fromUpdatedBalance
	fromTokenJSON, err := json.Marshal(fromToken)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(fromKey, fromTokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}
	toUpdatedBalance := toCurrentBalance + value
	toToken.Amount = toUpdatedBalance
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

// client가 어디든 참여하려고 할 때 시작점으로 두자.
func (s *SmartContract) Client(ctx contractapi.TransactionContextInterface) error {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	fmt.Println(clientMSPID)

	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	address := getAddress([]byte(id))

	// value, found, err := ctx.GetClientIdentity().GetAttributeValue("roles")

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{address})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	clientAmountBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account from world state: %v", err)
	}

	if clientAmountBytes != nil {
		return fmt.Errorf("client %s has been already registered: %v", address, err)
	}

	token := Token{tokenName, clientMSPID, 0, false}
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

// func AddMinter()          // event MinterAdded
// func RenounceMinter()     // event MinterRemoved
// func TransferMinterRole() // event MinterAdded & MinterRemoved

// Mint creates new tokens and adds them to minter's account balance
// This function triggers a Transfer event
// 얘도 마찬가지로 매개변수로 빼야함 id
// 완료
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, amount int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return err
	}

	// owner Address
	address := getAddress([]byte(id))

	// minterBytes, err := ctx.GetStub().GetState(minterKey)
	// if err != nil {
	// 	return fmt.Errorf("failed to read minter account from world state: %v", err)
	// }

	// minter := string(minterBytes)
	// if minter != getAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{address})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
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

	updatedBalance := currentBalance + amount

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

func _msgSender(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client id: %v", err)
	}

	return id, nil
}

func (s *SmartContract) Approve(ctx contractapi.TransactionContextInterface, spender string, amount int) (bool, error) {
	id, err := _msgSender(ctx)
	if err != nil {
		return false, err
	}

	// owner Address
	owner := getAddress([]byte(id))

	err = _approve(ctx, owner, spender, amount)
	if err != nil {
		return false, err
	}

	return true, nil
}

func _spendAllowance(ctx contractapi.TransactionContextInterface, owner string, spender string, amount int) error {
	currentAllowance, _, err := Allowance(ctx, owner, spender)
	if err != nil {
		return err
	}

	if currentAllowance < amount {
		return fmt.Errorf("허용량이 더 적다")
	}

	err = _approve(ctx, owner, spender, currentAllowance-amount)
	if err != nil {
		return err
	}

	return nil
}

func _afterTokenTransfer(ctx contractapi.TransactionContextInterface, allowance int, allowanceKey string, from string, to string, amount int) error {

	// 이 부분이 어디에 들어가야할지 무결성, 원자성을 확보하기 위한 고려를 진행하고 이를 문서화 시키자
	// Decrease the allowance
	updatedAllowance := allowance - amount
	err := ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(updatedAllowance)))
	if err != nil {
		return fmt.Errorf("허용량 업데이트 실패")
	}

	// log.Printf("spender %s allowance updated from %d to %d", spender, allowance, updatedAllowance)

	return nil
}
