package ethutil

import (
	"fmt"
	"github.com/obscuren/secp256k1-go"
	"math/big"
)

type Transaction struct {
	Nonce     string
	Recipient string
	Value     *big.Int
	Data      []string
	Memory    []int
	v         uint32
	r, s      []byte
}

func NewTransaction(to string, value *big.Int, data []string) *Transaction {
	tx := Transaction{Recipient: to, Value: value}
	tx.Nonce = "0"

	// Serialize the data
	tx.Data = make([]string, len(data))
	for i, val := range data {
		instr, err := CompileInstr(val)
		if err != nil {
			//fmt.Printf("compile error:%d %v\n", i+1, err)
		}

		tx.Data[i] = instr
	}

	tx.Sign([]byte("privkey"))
	tx.Sender()

	return &tx
}

func NewTransactionFromData(data []byte) *Transaction {
	tx := &Transaction{}
	tx.RlpDecode(data)

	return tx
}

func (tx *Transaction) Hash() []byte {
	preEnc := []interface{}{
		tx.Nonce,
		tx.Recipient,
		tx.Value,
		tx.Data,
	}

	return Sha256Bin(Encode(preEnc))
}

func (tx *Transaction) IsContract() bool {
	return tx.Recipient == ""
}

func (tx *Transaction) Signature(key []byte) []byte {
	hash := tx.Hash()
	sec := Sha256Bin(key)

	sig, _ := secp256k1.Sign(hash, sec)

	return sig
}

func (tx *Transaction) PublicKey() []byte {
	hash := Sha256Bin(tx.Hash())
	sig := append(tx.r, tx.s...)

	pubkey, _ := secp256k1.RecoverPubkey(hash, sig)

	return pubkey
}

func (tx *Transaction) Sender() []byte {
	pubkey := tx.PublicKey()

	// Validate the returned key.
	// Return nil if public key isn't in full format (04 = full, 03 = compact)
	if pubkey[0] != 4 {
		return nil
	}

	return Sha256Bin(pubkey[1:65])[12:]
}

func (tx *Transaction) Sign(privk []byte) {
	sig := tx.Signature(privk)

	// Add 27 so we get either 27 or 28 (for positive and negative)
	tx.v = uint32(sig[64]) + 27
	tx.r = sig[:32]
	tx.s = sig[32:65]
}

func (tx *Transaction) RlpEncode() []byte {
	// Prepare the transaction for serialization
	preEnc := []interface{}{
		tx.Nonce,
		tx.Recipient,
		tx.Value,
		tx.Data,
		tx.v,
		tx.r,
		tx.s,
	}

	return Encode(preEnc)
}

func (tx *Transaction) RlpDecode(data []byte) {
	decoder := NewRlpDecoder(data)

	tx.Nonce = decoder.Get(0).AsString()
	tx.Recipient = decoder.Get(0).AsString()
	tx.Value = decoder.Get(2).AsBigInt()

	d := decoder.Get(3)
	tx.Data = make([]string, d.Length())
	fmt.Println(d.Get(0))
	for i := 0; i < d.Length(); i++ {
		tx.Data[i] = d.Get(i).AsString()
	}

	tx.v = uint32(decoder.Get(4).AsUint())
	tx.r = decoder.Get(5).AsBytes()
	tx.s = decoder.Get(6).AsBytes()
}
