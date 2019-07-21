package parts

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//CLI define *blockchain as struct
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage : ")
	fmt.Println("	createblockchain =address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("	getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("	printchain - Print all the blocks of the blockchain")
	fmt.Println("	send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) createBlockChain(address string) {
	bc := CreateBlockChain(address)
	bc.Db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.Db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s' : %d\n", address, balance)
}

func (cli *CLI) printChain() {
	bc := NewBlockChain("")
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		//.db is opened in Next() function
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\t\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\t\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\t\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	//After finishing transaction, process consensus too
	//As a result, In this blockchain, a block can has only 1 transaction
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

//Run CLI
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	//String(name, value, usage)
	//name : when it is called
	//value : default value
	//usage : output of help
	createBlockChainAddress := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			createBlockChainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createBlockChainAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
