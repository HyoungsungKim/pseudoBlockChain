package main

import (
	"practice/pseudoBlockChain/blockchain/parts"
)

func main() {
	bc := parts.NewBlockChain()
	defer bc.Db.Close()

	cli := parts.CLI{Bc: bc}
	cli.Run()
}
