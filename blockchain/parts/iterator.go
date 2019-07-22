package parts

import (
	"log"

	"github.com/boltdb/bolt"
)

//BlockChainIterator define struct for blockchain iteration
type BlockChainIterator struct {
	CurrentHash []byte
	Db          *bolt.DB
}

//Iterator iterate blockchain
func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{
		CurrentHash: bc.Tip,
		Db:          bc.Db,
	}

	return bci
}

//Next move iterator to the next
//.db is opened in this function
func (i *BlockChainIterator) Next() *Block {
	var block *Block

	err := i.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//Get serialized block
		encodedBlock := b.Get(i.CurrentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	//current block's prevBlochHash is current block hash of last block
	i.CurrentHash = block.PrevBlockHash

	return block
}
