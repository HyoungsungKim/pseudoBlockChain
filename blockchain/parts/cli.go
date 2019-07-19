package parts

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//CLI define *blockchain as struct
type CLI struct {
	Bc *BlockChain
}

func (cli *CLI) createBlockChain(address string) {
	bc := CreateBlockChain(address)
	bc.Db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage : ")
	fmt.Println("	addblock -data BLOCK_DATA -add a block to the blockchain")
	fmt.Println("	printchain -print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain() {
	bci := cli.Bc.Iterator()

	for {
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

//Run CLI
func (cli *CLI) Run() {
	cli.validateArgs()

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)

	createBlockChainAddress := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
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

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
