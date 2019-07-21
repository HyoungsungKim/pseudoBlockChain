package parts

import "bytes"

//TxInput includes TXid, output, scriptsig
//Input reference previous output. That is why TXInput has Vout
type TxInput struct {
	Txid      []byte
	Vout      int //stores an index of an output in the transaction.
	Signature []byte
	PubKey    []byte
}

//UseKey compare input parameter with receiver's public key hash
//If it is same, then return true else false
func (in *TxInput) UseKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

//CanUnlockOutputWith compare TxInput's ScriptSig with unlockingData
func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
