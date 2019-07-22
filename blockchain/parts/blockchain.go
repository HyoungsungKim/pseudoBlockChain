package parts

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

//BlockChain chain of blocks
//tip is hash of last chain
type BlockChain struct {
	Tip []byte
	Db  *bolt.DB
}

//MineBlock mine a new Block
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.Tip = newBlock.Hash

		return nil
	})
}

//dbExists check there is a .db or not
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//NewBlockChain creates new blockchain
func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		//bucket -> block -> tip
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{
		Tip: tip,
		Db:  db,
	}
	return &bc
}

//CreateBlockChain literally create new blockchain
func CreateBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("BlockChain already exists")
		os.Exit(1)
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Println("Cannot open .db")
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		cbTx := NewCoinbaseTx(address, genesisCoinbaseData)
		//Put cbTx to make a genesis block
		genesis := NewGenesisBlock(cbTx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{
		Tip: tip,
		Db:  db,
	}
	return &bc
}

//FindUnspentTransactions find transaction containing UTXO
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction
	//key : string, value : []int
	spentTxos := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			//Umm... Can't understand well
			for outIdx, out := range tx.Vout {
				if spentTxos[txID] != nil {
					for _, spentOutIdx := range spentTxos[txID] {
						//For what...?
						if spentOutIdx == outIdx {
							//If number of input is same with output,
							//It means all of utxos are spent
							//so go back
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			//Check in a block to find all of UTXO
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					//Compare address and script and check it is true
					if in.UseKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTxos[inTxID] = append(spentTxos[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTxs
}

//FindUTXO find UTXO
func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

//FindSpendableOutputs retrieve spendable output
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

//FindTransaction find transaction using ID
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			//Find only one transaction?
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transation is not found")
}

//SignTransaction set tx's signature using privkey
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	//If There are transations using Txid of vin
	for _, vin := range tx.Vin {
		prevTx, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		//Assignment! not compare
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	//Set signature of tx using privKey and prevTxs
	tx.Sign(privKey, prevTxs)
}

//VerifyTransaction find input transaction in blockchain(receiver) and verify it
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		//bc.FindTransaction(vin.Txid) : find only 1 transaction
		prevTx, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
}
