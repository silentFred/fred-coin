package cmd

import (
	"context"
	"day-3/lottery"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const TestContractAddress = "0xEf4B9cf94fC0139880c4aE697fa09Fbf71600c05"

func deployAndTestLotteryContract() *cobra.Command {
	return &cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {

			balance, _ := GetAccountBalance()
			log.Println("current account balance is: ", balance)

			log.Println("deploying contract...")
			_, address, _ := DeployContract()

			log.Println("contract deployed to address: ", address)

			log.Println("entering the lottery...")
			_, _ = EnterLottery()

			balanceAfterEntry, _ := GetAccountBalance()
			log.Println("current account balance after lottery entry is: ", balanceAfterEntry)

			players, _ := GetLotteryPlayers()
			log.Println("current lottery players are: ", players)

			winner, _ := PickLotteryWinner()
			log.Println("lottery winner described in transaction hash: ", winner.Hash())
			log.Println("done.")
		},
	}
}

func GetAccountBalance() (*big.Int, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	address := common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS"))
	balance, err := client.BalanceAt(context.Background(), address, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return balance, nil
}

func EnterLottery() (*types.Transaction, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	lotteryContract, err := lottery.NewLottery(common.HexToAddress(TestContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Storage contract: %w", err)
	}

	transactionOptions, err := GetTransactionOptions(client)
	if err != nil {
		return nil, err
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS")))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending nonce: %w", err)
	}

	transactionOptions.Nonce = big.NewInt(int64(nonce))
	transactionOptions.Value = big.NewInt(12000000000000000)
	transactionOptions.GasLimit = uint64(300000)

	transaction, err := lotteryContract.Enter(transactionOptions)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch lottery players: %w", err)
	}

	log.Println("Lottery entered: ", transaction.Hash())
	return transaction, nil
}

func GetLotteryPlayers() ([]common.Address, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	lotteryContract, err := lottery.NewLottery(common.HexToAddress(TestContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind lottery contract: %w", err)
	}

	players, err := lotteryContract.GetPlayers(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lottery players: %w", err)
	}

	return players, nil
}

func PickLotteryWinner() (*types.Transaction, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	lotteryContract, err := lottery.NewLottery(common.HexToAddress(TestContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind lottery contract: %w", err)
	}

	auth, err := GetTransactionOptions(client)
	if err != nil {
		return nil, err
	}

	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS")))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(12000000000000000)
	auth.GasLimit = uint64(300000)

	winner, err := lotteryContract.PickWinner(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to pick lottery winner: %w", err)
	}

	return winner, nil
}

func DeployContract() (*lottery.Lottery, common.Address, error) {

	client, err := GetClient()
	if err != nil {
		return nil, common.Address{}, err
	}

	transactionOptions, err := GetTransactionOptions(client)
	if err != nil {
		return nil, common.Address{}, err
	}

	contractAddress, transaction, contract, err := lottery.DeployLottery(transactionOptions, client)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to deploy contract: %w", err)
	}
	_ = contractAddress
	_ = contract

	// Wait until the contract is deployed
	contractAddress, err = bind.WaitDeployed(context.Background(), client, transaction)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("error occured while waiting for contract to deploy: %w", err)
	}

	fmt.Printf("Contract deployed at %v\n", contractAddress)

	return contract, contractAddress, nil
}

func GetTransactionOptions(cl *ethclient.Client) (*bind.TransactOpts, error) {
	var (
		sk = crypto.ToECDSAUnsafe(common.FromHex(os.Getenv("ACCOUNT_PRIVATE_KEY")))
	)
	// Retrieve the chainid (needed for signer)
	chainid, err := cl.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chain id: %w", err)
	}

	// Create the transactOpts (signer)
	transactOpts, err := bind.NewKeyedTransactorWithChainID(sk, chainid)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyed transaction: %w", err)
	}
	return transactOpts, nil
}

func GetClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(os.Getenv("NODE_ENDPOINT"))
	if err != nil {
		return nil, fmt.Errorf("failed to create eth client: %w", err)
	}
	return client, nil
}
