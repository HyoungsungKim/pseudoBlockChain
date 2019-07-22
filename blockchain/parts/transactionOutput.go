package parts

import "bytes"

//TxOutput includes value and PubKeyHash
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

//Lock function set receiver's public key to hash of address
//Hash function is base58 decoder
func (out *TxOutput) Lock(address []byte) {
	//Why decoder? not encoder?
	pubKeyHash := Base58Decode(address)
	//checksum use only 4 bytes of result of sha256
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//IsLockedWithKey compare receiver's public key and input public key
// If it is same, then return true
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

//NewTxOutput set *TxOutput
func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{
		Value:      value,
		PubKeyHash: nil,
	}
	txo.Lock([]byte(address))
	return txo
}
