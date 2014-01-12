package ethutil

import (
	"fmt"
	"github.com/obscuren/secp256k1-go"
)

type Transaction struct {
	nonce     string
	sender    string
	recipient string
	value     uint64
	fee       uint32
	data      []string
	memory    []int
	lastTx    string
	v         uint32
	r, s      []byte
}

func NewTransaction(to string, value uint64, data []string) *Transaction {
	tx := Transaction{sender: "1234567890", recipient: to, value: value}
	tx.nonce = "0"
	tx.fee = 0 //uint32((ContractFee + MemoryFee * float32(len(tx.data))) * 1e8)
	tx.lastTx = "0"

	// Serialize the data
	tx.data = make([]string, len(data))
	for i, val := range data {
		instr, err := CompileInstr(val)
		if err != nil {
			//fmt.Printf("compile error:%d %v\n", i+1, err)
		}

		tx.data[i] = instr
	}

	tx.Sign([]byte("privkey"))
	tx.Sender()

	return &tx
}

func (tx *Transaction) Hash() []byte {
	preEnc := []interface{}{
		tx.nonce,
		tx.recipient,
		tx.value,
		tx.fee,
		tx.data,
	}

	return Sha256Bin(Encode(preEnc))
}

func (tx *Transaction) IsContract() bool {
	return tx.recipient == ""
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

func (tx *Transaction) MarshalRlp() []byte {
	// Prepare the transaction for serialization
	preEnc := []interface{}{
		tx.nonce,
		tx.recipient,
		tx.value,
		tx.fee,
		tx.data,
		tx.v,
		tx.r,
		tx.s,
	}

	return Encode(preEnc)
}

func (tx *Transaction) UnmarshalRlp(data []byte) {
	decoder := NewRlpDecoder(data)

	tx.nonce = decoder.Get(0).AsString()
	tx.recipient = decoder.Get(0).AsString()
	tx.value = decoder.Get(2).AsUint()
	tx.fee = uint32(decoder.Get(3).AsUint())

	d := decoder.Get(4)
	tx.data = make([]string, d.Length())
	fmt.Println(d.Get(0))
	for i := 0; i < d.Length(); i++ {
		tx.data[i] = d.Get(i).AsString()
	}

	tx.v = uint32(decoder.Get(5).AsUint())
	tx.r = decoder.Get(6).AsBytes()
	tx.s = decoder.Get(7).AsBytes()
}
