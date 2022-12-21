package token

const (
	DocType_Token                  = "DOCTYPE_TOKEN"
	DocType_TotalSupply            = "DOCTYPE_TOTALSUPPLY"
	DocType_TotalSupplyByPartition = "DOCTYPE_TOTALSUPPLYBYPARTITION"
	DocType_Allowance              = "DOCTYPE_ALLOWANCE"
	DocType_Test                   = "DOCTYPE_TEST"

	BalancePrefix = "balancePrefix"

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

	clientBalancePrefix = "clientWallet"

	operatorForPartitionPrefix = "operatorForPartition"
)

// totalSupply
type TotalSupplyStruct struct {
	DocType string `json:"docType"`

	TotalSupply int64 `json:"totalSupply"`
}

type TotalSupplyByPartitionStruct struct {
	DocType string `json:"docType"`

	TotalSupply int64 `json:"totalSupply"`
	// Partition Address
	Partition string `json:"partition"`
}

type AllowanceByPartitionStruct struct {
	DocType string `json:"docType"`

	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Partition string `json:"partition"`
	Amount    int64  `json:"amount"`
}

type TransferByPartitionStruct struct {
	DocType string `json:"docType"`

	From      string `json:"from"`
	To        string `json:"to"`
	Partition string `json:"partition"`
	Amount    int64  `json:"amount"`
}

type MintByPartitionStruct struct {
	DocType string `json:"docType"`

	Minter    string `json:"minter"`
	Partition string `json:"partition"`
	Amount    int64  `json:"amount"`
}

// partition Token
type PartitionToken struct {
	DocType string `json:"docType"`

	TokenName string `json:"name"`
	TokenID   string `json:"id"`
	IsLocked  bool   `json:"islocked"`
	TxId      string `json:"txId"`

	Publisher string `json:"publisher"`

	CreatedDate string `json:"createdDate"`
	UpdatedDate string `json:"updatedDate"`
	ExpiredDate string `json:"expiredDate"`

	Amount int64 `json:"amount"`

	// 요기가 fix 될 예정
	// 여기는 이제 사라짐
	// Partition Partition `json:"partition"`
}
