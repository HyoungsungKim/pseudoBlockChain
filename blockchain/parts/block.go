package parts

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

//Block Define basic block struct
type Block struct {
	TimeStamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//SetHash Set Hash of Block struct
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
	//func Join(s [][]byte, sep []byte) []byte
	//Concatenate 2D-byte using 1D-byte seperator
	//{"Bob", "Alice"} + {", "}	= {"Bob, Alice"}
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

//NewBlock constructor Block
func NewBlock(data string, PrevBlockHash []byte) *Block {
	block := &Block{
		TimeStamp:     time.Now().Unix(),
		Data:          []byte(data),
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
