package ethutil

import (
	"math/big"
)

type Contract struct {
	Amount *big.Int
	Nonce  uint64
	state  *Trie
}

func NewContract(Amount *big.Int, root []byte) *Contract {
	contract := &Contract{Amount: Amount}
	contract.state = NewTrie(Config.Db, string(root))

	return contract
}

func (c *Contract) RlpEncode() []byte {
	return Encode([]interface{}{c.Amount, c.Nonce, c.state.Root})
}

func (c *Contract) RlpDecode(data []byte) {
	decoder := NewRlpDecoder(data)

	c.Amount = decoder.Get(0).AsBigInt()
	c.Nonce = decoder.Get(1).AsUint()
	c.state = NewTrie(Config.Db, decoder.Get(2).AsString())
}

func (c *Contract) State() *Trie {
	return c.state
}

type Ether struct {
	Amount *big.Int
	Nonce  uint64
}

func NewEther(amount *big.Int) *Ether {
	return &Ether{Amount: amount, Nonce: 0}
}

func NewEtherFromData(data []byte) *Ether {
	ether := &Ether{}
	ether.RlpDecode(data)

	return ether
}

func (e *Ether) AddFee(fee *big.Int) {
	e.Amount = e.Amount.Add(e.Amount, fee)
}

func (e *Ether) RlpEncode() []byte {
	return Encode([]interface{}{e.Amount, e.Nonce, ""})
}

func (e *Ether) RlpDecode(data []byte) {
	decoder := NewRlpDecoder(data)

	e.Amount = decoder.Get(1).AsBigInt()
	e.Nonce = decoder.Get(2).AsUint()
}
