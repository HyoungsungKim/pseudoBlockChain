package parts

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

//Block Define basic block struct
type Block struct {
	TimeStamp     int64
	Transaction   []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//NewBlock constructor Block
func NewBlock(transactions []*Transaction, PrevBlockHash []byte) *Block {
	block := &Block{
		TimeStamp:     time.Now().Unix(),
		Transaction:   transactions,
		PrevBlockHash: PrevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

//NewGenesisBlock generate genesis block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//Serialize *Block to []byte
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	//Transmit b to encoder
	err := encoder.Encode(b)

	if err != nil {
		fmt.Println("Serialize error. Check receiver block")
		log.Panic(err)
	}

	return result.Bytes()
}

//DeserializeBlock []byte to *Block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	//read decoder and store it to block
	err := decoder.Decode(&block)

	if err != nil {
		fmt.Println("Serialize error. Check input parameter")
		log.Panic(err)
	}

	return &block
}

//HashTransactions concatnate ID of transactions and put it into sha256
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	//https://www.dotnetperls.com/2d-go
	//Good example to understand 2d slice append
	for _, tx := range b.Transaction {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
