package ethutil

import (
	"math/big"
)

type Contract struct {
	t      uint32 // contract is always 1
	Amount *big.Int
	state  *Trie
}

func NewContract(Amount *big.Int, root []byte) *Contract {
	contract := &Contract{t: 1, Amount: Amount}
	contract.state = NewTrie(Config.Db, string(root))

	return contract
}

func (c *Contract) MarshalRlp() []byte {
	return Encode([]interface{}{c.t, c.Amount, c.state.Root})
}

func (c *Contract) UnmarshalRlp(data []byte) {
	decoder := NewRlpDecoder(data)

	c.t = uint32(decoder.Get(0).AsUint())
	c.Amount = decoder.Get(1).AsBigInt()
	c.state = NewTrie(Config.Db, decoder.Get(2).AsString())
}

func (c *Contract) State() *Trie {
	return c.state
}

type Ether struct {
	t      uint32
	Amount *big.Int
	Nonce  string
}

func NewEtherFromData(data []byte) *Ether {
	ether := &Ether{}
	ether.UnmarshalRlp(data)

	return ether
}

func (e *Ether) AddFee(fee *big.Int) {
	e.Amount = e.Amount.Add(e.Amount, fee)
}

func (e *Ether) MarshalRlp() []byte {
	return Encode([]interface{}{e.t, e.Amount, e.Nonce})
}

func (e *Ether) UnmarshalRlp(data []byte) {
	decoder := NewRlpDecoder(data)

	e.t = uint32(decoder.Get(0).AsUint())
	e.Amount = decoder.Get(1).AsBigInt()
	e.Nonce = decoder.Get(2).AsString()
}
