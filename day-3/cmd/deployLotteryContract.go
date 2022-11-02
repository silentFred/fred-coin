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
	"github.com/ethereum/go-ethereum/params"
)

const TestContractAddress = "0xEf4B9cf94fC0139880c4aE697fa09Fbf71600c05"

func deployAndTestLotteryContract() *cobra.Command {
	return &cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {

			balance, _ := GetAccountBalance()
			log.Println("current account balance is: ", balance)

			log.Println("deploying contract...")
			_, address := DeployContract()

			log.Println("contract deployed to address: ", address)

			log.Println("entering the lottery...")
			EnterLottery()

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
	client := GetClient()

	ctx := context.Background()
	// Get Balance of an account (nil means at newest block)
	addr := common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS"))
	balance, err := client.BalanceAt(ctx, addr, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return balance, nil
}

func EnterLottery() *types.Transaction {
	client := GetClient()

	lotteryContract, err := lottery.NewLottery(common.HexToAddress(TestContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate Storage contract: %v", err)
	}

	auth := GetTransactor(client)

	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS")))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(12000000000000000)
	auth.GasLimit = uint64(300000)

	enter, err := lotteryContract.Enter(auth)

	if err != nil {
		log.Fatalf("Failed to fetch lottery players: %v", err)
	}
	log.Println("Lottery entered: ", enter.Hash())
	return enter
}

func GetLotteryPlayers() ([]common.Address, error) {
	client := GetClient()
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
	client := GetClient()
	lotteryContract, err := lottery.NewLottery(common.HexToAddress(TestContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind lottery contract: %w", err)
	}

	auth := GetTransactor(client)

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

func DeployContract() (*lottery.Lottery, common.Address) {

	cl := GetClient()

	transactOpts := GetTransactor(cl)

	// Deploy a CoolContract
	addr, tx, contract, err := lottery.DeployLottery(transactOpts, cl)
	if err != nil {
		log.Fatal(err)
	}
	_ = addr
	_ = contract

	// Wait until the contract is deployed
	addr, err = bind.WaitDeployed(context.Background(), cl, tx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Contract deployed at %v\n", addr)

	return contract, addr
}

func GetTransactor(cl *ethclient.Client) *bind.TransactOpts {
	var (
		sk = crypto.ToECDSAUnsafe(common.FromHex(os.Getenv("ACCOUNT_PRIVATE_KEY")))
	)
	// Retrieve the chainid (needed for signer)
	chainid, err := cl.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create the transactOpts (signer)
	transactOpts, err := bind.NewKeyedTransactorWithChainID(sk, chainid)
	if err != nil {
		log.Fatal(err)
	}
	return transactOpts
}

func SendTransaction(cl *ethclient.Client) error {
	var (
		sk       = crypto.ToECDSAUnsafe(common.FromHex(os.Getenv("ACCOUNT_PRIVATE_KEY")))
		to       = common.HexToAddress("0xa8467374a4288582CA894EF0127f48B35D0F35d0")
		value    = new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether))
		sender   = common.HexToAddress(os.Getenv("ACCOUNT_ADDRESS"))
		gasLimit = uint64(21000)
	)

	// Retrieve the chainid (needed for signer)
	chainid, err := cl.ChainID(context.Background())
	if err != nil {
		return err
	}

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), sender)
	if err != nil {
		return err
	}

	// Get suggested gas price
	tipCap, _ := cl.SuggestGasTipCap(context.Background())
	feeCap, _ := cl.SuggestGasPrice(context.Background())
	// Create a new transaction
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainid,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			Gas:       gasLimit,
			To:        &to,
			Value:     value,
			Data:      nil,
		})

	// Sign the transaction using our keys
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)

	// Send the transaction to our node
	return cl.SendTransaction(context.Background(), signedTx)
}

func GetClient() *ethclient.Client {
	client, err := ethclient.Dial(os.Getenv("NODE_ENDPOINT"))
	if err != nil {
		log.Fatalln(err)
	}
	return client
}
