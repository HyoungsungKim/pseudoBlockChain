package parts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 10

//Transaction includes ID, Transaction input and output
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

//Serialize serialize transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash generate sha256 using tx without ID
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
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

//String set information of tx to sting
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

//NewCoinbaseTx mint coinbase transaction of miner
func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TxInput{
		Txid: []byte{},
		//Vout = -1 means it is coinbase transaction
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}
	txout := NewTxOutput(subsidy, to)
	tx := Transaction{
		ID:   nil,
		Vin:  []TxInput{txin},
		Vout: []TxOutput{*txout},
	}
	tx.ID = tx.Hash()

	return &tx

}

//IsCoinbase check it is coinbase transaction
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//NewUTXOTransaction create new utxo transaction
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	//Read all of wallets
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	//get 1 wallet
	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey(wallet.PublicKey)

	//accumulated, validaoutput
	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR : Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)

		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{
				Txid:      txID,
				Vout:      out,
				Signature: nil,
				PubKey:    wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))
	if acc > amount {
		//Return change to sender
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

//TrimmedCopy trim signature of inputs and public key of inputs and copy others
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TxInput{
			Txid:      vin.Txid,
			Vout:      vin.Vout,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TxOutput{
			Value:      vout.Value,
			PubKeyHash: vout.PubKeyHash,
		})
	}

	txCopy := Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: outputs,
	}

	return txCopy
}

//Sign Set signature of receiver
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	//Coinbase transactions are not signed because there are no real inputs in them.
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		//get previous transaction of vin using txid
		prevTx := prevTxs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		//Hash method serialize the transaction and hashed it with the SHA-256 algorithm
		txCopy.ID = txCopy.Hash()

		txCopy.Vin[inID].PubKey = nil

		//What is sign?
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}

//Verify compare receiver's signature and prevTx's signature
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		//prevTx : Tranaction of vin.Txid
		prevTx := prevTxs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		//serialize txCopy without signature
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		//-----------Same process with sign function-----------

		//signature is a pair of numbers
		//and a public key is a pair of coordinates.
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen)/2])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		//Compare receiver's signature and input parameter
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}
