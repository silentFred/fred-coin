package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
)

const GoerliChainId = int64(5)

func buildAndBindContractCommand() *cobra.Command {
	return &cobra.Command{
		Use: "build",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("building api...")
			buildAbi()
			log.Println("building binary...")
			buildBinary()
			log.Println("generating go client code...")
			buildGoContractClient()
			log.Println("testing contract binding with goerli chain...")
			manager, _ := GetContractManagerAddress()
			log.Println("contract owner address: ", manager)
			log.Println("done.")
		},
	}
}

func GetContractManagerAddress() (interface{}, error) {
	contract, err := getContract()

	manager, err := contract.Call("manager")
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return manager, nil
}

func getContract() (*eth.Contract, error) {
	var rpcProviderURL = os.Getenv("NODE_ENDPOINT")
	web3Client, err := web3.NewWeb3(rpcProviderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to provision web3 client: %w", err)
	}

	err = web3Client.Eth.SetAccount(os.Getenv("ACCOUNT_PRIVATE_KEY"))
	if err != nil {
		return nil, fmt.Errorf("failed to set ETH account private key: %w", err)
	}

	web3Client.Eth.SetChainId(GoerliChainId)

	abiFileContentString := getContractAbi()

	contract, err := web3Client.Eth.NewContract(abiFileContentString, os.Getenv("LOTTERY_CONTRACT_ADDRESS"))
	if err != nil {
		return nil, fmt.Errorf("failed to bind smart contract: %w", err)
	}
	log.Println("contract binding successful at address: ", contract.Address())
	return contract, nil
}

func getContractAbi() string {
	abi, err := os.ReadFile("./build/Lottery.abi")
	if err != nil {
		log.Fatal(err)
	}
	abiFileContentString := string(abi)
	return abiFileContentString
}

func buildGoContractClient() {
	command := exec.Command(
		"abigen",
		"--abi=./build/Lottery.abi",
		"--bin=./build/Lottery.bin",
		"--pkg=lottery",
		"--out=./lottery/Lottery.go")

	_, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}
}

func buildBinary() {
	command := exec.Command(
		"solc",
		"--optimize",
		"--bin",
		"--overwrite",
		"./contracts/Lottery.sol",
		"-o",
		"build")

	_, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}
}

func buildAbi() {
	command := exec.Command(
		"solc",
		"--optimize",
		"--abi",
		"--overwrite",
		"./contracts/lottery.sol",
		"-o",
		"build")

	_, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}
}
