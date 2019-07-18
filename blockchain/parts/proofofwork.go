package parts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const (
	targetBits = 16
	maxNonce   = math.MaxInt64
)

//ProofOfWork define basic struct for PoW
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

//NewProofOfWork initialize ProofOfWork struct
func NewProofOfWork(b *Block) *ProofOfWork {
	//bit move uint(256-targetBits) step to right
	//000...00100000...00
	//24 zeros + 1 + 231 zeros -> 6 leading zero in hex
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

//prepareData is concatnate block information to []byte type data
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.Data,
			IntToHex(pow.Block.TimeStamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

//Run Proof of Work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("mining the block containing \"%s\"n", pow.Block.Data)
	//To avoid overflow of nonce
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		//%x : base 16, with lower-case letters for a-f
		//\r : Carriage return
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])

		// x.Cmp(y)
		//   -1 if x <  y
		//    0 if x == y (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
		//   +1 if x >  y
		if hashInt.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

//Validate proof of work
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid
}
