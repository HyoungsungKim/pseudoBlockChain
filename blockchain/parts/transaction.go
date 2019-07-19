package parts

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

const subsidy = 10

//Transaction includes ID, Transaction input and output
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

//TxInput includes TXid, output, scriptsig
//Input reference previous output. That is why TXInput has Vout
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

//TxOutput includes value and scriptPubKey
type TxOutput struct {
	Value int
	//If ScriptPubKey is correct, then output can be unlocked
	ScriptPubKey string
}

//SetID literally set transaction ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	//receiver of Encode function is pointer
	//Therefore when wnc is changed, encoded will be changed too
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

//NewCoinbaseTx mint coinbase transaction of miner
func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TxInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}
	txout := TxOutput{
		Value:        subsidy,
		ScriptPubKey: to,
	}

	tx := Transaction{
		ID:   nil,
		Vin:  []TxInput{txin},
		Vout: []TxOutput{txout},
	}
	tx.SetID()

	return &tx

}

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
