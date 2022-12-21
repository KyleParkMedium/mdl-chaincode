package airdrop

import (
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ledgermanager"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/wallet"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func AirDrop(ctx contractapi.TransactionContextInterface, tokenWallet wallet.TokenWallet) (*wallet.TokenWallet, error) {

	_, err := ledgermanager.PutState(wallet.DocType_TokenWallet, tokenWallet.TokenWalletId, tokenWallet, ctx)
	if err != nil {
		return nil, err
	}
	return &tokenWallet, nil
}
