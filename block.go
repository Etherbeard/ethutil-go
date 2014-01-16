package ethutil

import (
	"fmt"
	"math/big"
	"time"
)

type Block struct {
	// The number of this block
	number uint32
	// Hash to the previous block
	PrevHash string
	// Uncles of this block
	Uncles   []*Block
	UncleSha []byte
	// The coin base address
	Coinbase string
	// Block Trie state
	state *Trie
	// Difficulty for the current block
	Difficulty *big.Int
	// Creation time
	time int64
	// Block Nonce for verification
	Nonce *big.Int
	// List of transactions and/or contracts
	transactions []*Transaction
	TxSha        []byte
	// Extra (unused)
	extra string
}

// New block takes a raw encoded string
func NewBlock(raw []byte) *Block {
	block := &Block{}
	block.UnmarshalRlp(raw)

	return block
}

// Creates a new block. This is currently for testing
func CreateTestBlock( /* TODO use raw data */ transactions []*Transaction) *Block {
	block := &Block{
		// Slice of transactions to include in this block
		transactions: transactions,
		number:       1,
		PrevHash:     "1234",
		Coinbase:     "me",
		Difficulty:   big.NewInt(10),
		Nonce:        BigInt0,
		time:         time.Now().Unix(),
	}

	return block
}

func CreateBlock(root string,
	num int,
	PrevHash string,
	base string,
	Difficulty *big.Int,
	Nonce *big.Int,
	extra string,
	txes []*Transaction) *Block {

	block := &Block{
		// Slice of transactions to include in this block
		transactions: txes,
		number:       uint32(num),
		PrevHash:     PrevHash,
		Coinbase:     base,
		Difficulty:   Difficulty,
		Nonce:        Nonce,
		time:         time.Now().Unix(),
		extra:        extra,
	}
	block.state = NewTrie(Config.Db, root)

	for _, tx := range txes {
		// Create contract if there's no recipient
		if tx.IsContract() {
			addr := tx.Hash()

			value := tx.Value
			contract := NewContract(value, []byte(""))
			block.state.Update(string(addr), string(contract.MarshalRlp()))
			for i, val := range tx.Data {
				contract.state.Update(string(NumberToBytes(uint64(i), 32)), val)
			}
			block.UpdateContract(addr, contract)
		}
	}

	return block
}

func (block *Block) State() *Trie {
	return block.state
}

func (block *Block) Transactions() []*Transaction {
	return block.transactions
}

func (block *Block) GetContract(addr []byte) *Contract {
	data := block.state.Get(string(addr))
	if data == "" {
		return nil
	}

	contract := &Contract{}
	contract.UnmarshalRlp([]byte(data))

	return contract
}

func (block *Block) UpdateContract(addr []byte, contract *Contract) {
	block.state.Update(string(addr), string(contract.MarshalRlp()))
}

func (block *Block) PayFee(addr []byte, fee *big.Int) bool {
	contract := block.GetContract(addr)
	// If we can't pay the fee return
	if contract == nil || contract.Amount.Cmp(fee) < 0 /* amount < fee */ {
		fmt.Println("Contract has insufficient funds", contract.Amount, fee)

		return false
	}

	base := new(big.Int)
	contract.Amount = base.Sub(contract.Amount, fee)
	block.state.Update(string(addr), string(contract.MarshalRlp()))

	data := block.state.Get(string(block.Coinbase))

	// Get the ether (Coinbase) and add the fee (gief fee to miner)
	ether := NewEtherFromData([]byte(data))

	base = new(big.Int)
	ether.Amount = base.Add(ether.Amount, fee)

	block.state.Update(string(block.Coinbase), string(ether.MarshalRlp()))

	return true
}

// Returns a hash of the block
func (block *Block) Hash() []byte {
	return Sha256Bin(Encode(block.header(block.TxSha, block.UncleSha)))
}

func (block *Block) MarshalRlp() []byte {
	// Marshal the transactions of this block
	encTx := make([]string, len(block.transactions))
	for i, tx := range block.transactions {
		// Cast it to a string (safe)
		encTx[i] = string(tx.MarshalRlp())
	}
	tsha := Sha256Bin([]byte(Encode(encTx)))

	uncles := make([]interface{}, len(block.Uncles))
	for i, uncle := range block.Uncles {
		uncles[i] = uncle.uncleHeader()
	}

	// Sha of the concatenated uncles
	usha := Sha256Bin(Encode(uncles))
	// The block header
	header := block.header(tsha, usha)

	// Encode a slice interface which contains the header and the list of
	// transactions.
	return Encode([]interface{}{header, encTx, uncles})
}

func (block *Block) UnmarshalRlp(data []byte) {
	decoder := NewRlpDecoder(data)

	header := decoder.Get(0)

	block.PrevHash = header.Get(0).AsString()
	block.UncleSha = header.Get(1).AsBytes()
	block.Coinbase = header.Get(2).AsString()
	block.state = NewTrie(Config.Db, header.Get(3).AsString())
	block.TxSha = header.Get(4).AsBytes()
	block.Difficulty = header.Get(5).AsBigInt()
	block.time = int64(header.Get(6).AsUint())
	block.Nonce = header.Get(7).AsBigInt()

	// Tx list might be empty if this is an uncle. Uncles only have their
	// header set.
	if decoder.Get(1).IsNil() == false { // Yes explicitness
		txes := decoder.Get(1)
		block.transactions = make([]*Transaction, txes.Length())
		for i := 0; i < txes.Length(); i++ {
			tx := &Transaction{}
			tx.UnmarshalRlp(txes.Get(i).AsBytes())
			block.transactions[i] = tx
		}
	}

	if decoder.Get(2).IsNil() == false { // Yes explicitness
		uncles := decoder.Get(2)
		block.Uncles = make([]*Block, uncles.Length())
		for i := 0; i < uncles.Length(); i++ {
			block := &Block{}
			// This is terrible but it's the way it has to be since
			// I'm going by now means doing it by hand (the data is in it's raw format in interface form)
			block.UnmarshalRlp(Encode(uncles.Get(i).AsRaw()))
			block.Uncles[i] = block
		}
	}
}

//////////// UNEXPORTED /////////////////
func (block *Block) header(txSha []byte, uncleSha []byte) []interface{} {
	return []interface{}{
		// Sha of the previous block
		block.PrevHash,
		// Sha of uncles
		uncleSha,
		// Coinbase address
		block.Coinbase,
		// root state
		block.state.Root,
		// Sha of tx
		txSha,
		// Current block Difficulty
		block.Difficulty,
		// Time the block was found?
		uint64(block.time),
		// Block's Nonce for validation
		block.Nonce,
	}
}

func (block *Block) uncleHeader() []interface{} {
	return []interface{}{
		// Sha of the previous block
		block.PrevHash,
		// Sha of uncles
		block.UncleSha,
		// Coinbase address
		block.Coinbase,
		// root state
		block.state.Root,
		// Sha of tx
		block.TxSha,
		// Current block Difficulty
		block.Difficulty,
		// Time the block was found?
		uint64(block.time),
		// Block's Nonce for validation
		block.Nonce,
	}
}
