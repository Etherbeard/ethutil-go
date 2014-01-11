package ethutil

import (
	"fmt"
	"time"
	"math/big"
)

type Block struct {
	// The number of this block
	number uint32
	// Hash to the previous block
	prevHash string
	// Uncles of this block
	uncles   []string
	// The coin base address
	coinbase string
	// Block Trie state
	state      *Trie
	// Difficulty for the current block
	difficulty *big.Int
	// Creation time
	time  int64
	// Block nonce for verification
	nonce *big.Int
	// List of transactions and/or contracts
	transactions []*Transaction
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
		prevHash:     "1234",
		coinbase:     "me",
		difficulty:   big.NewInt(10),
		nonce:        BigInt0,
		time:         time.Now().Unix(),
	}

	return block
}

func CreateBlock(root string,
	num int,
	prevHash string,
	base string,
	difficulty *big.Int,
	nonce *big.Int,
	extra string,
	txes []*Transaction) *Block {

	block := &Block{
		// Slice of transactions to include in this block
		transactions: txes,
		number:       uint32(num),
		prevHash:     prevHash,
		coinbase:     base,
		difficulty:   difficulty,
		nonce:        nonce,
		time:         time.Now().Unix(),
		extra:        extra,
	}
	block.state = NewTrie(Config.Db, root)

	for _, tx := range txes {
		// Create contract if there's no recipient
		if tx.IsContract() {
			addr := tx.Hash()

			contract := NewContract(tx.value, []byte(""))
			block.state.Update(string(addr), string(contract.MarshalRlp()))
			for i, val := range tx.data {
				contract.state.Update(string(NumberToBytes(uint64(i), 32)), val)
			}
			block.UpdateContract(addr, contract)
		}
	}

	return block
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

func (block *Block) PayFee(addr []byte, fee uint64) bool {
	contract := block.GetContract(addr)
	// If we can't pay the fee return
	if contract == nil || contract.amount < fee {
		fmt.Println("Contract has insufficient funds", contract.amount, fee)

		return false
	}

	contract.amount -= fee
	block.state.Update(string(addr), string(contract.MarshalRlp()))

	data := block.state.Get(string(block.coinbase))

	// Get the ether (coinbase) and add the fee (gief fee to miner)
	ether := NewEtherFromData([]byte(data))
	ether.amount += fee

	block.state.Update(string(block.coinbase), string(ether.MarshalRlp()))

	return true
}

// Returns a hash of the block
func (block *Block) Hash() []byte {
	return Sha256Bin(block.MarshalRlp())
}

func (block *Block) MarshalRlp() []byte {
	// Marshal the transactions of this block
	encTx := make([]string, len(block.transactions))
	for i, tx := range block.transactions {
		// Cast it to a string (safe)
		encTx[i] = string(tx.MarshalRlp())
	}
	tsha := Sha256Bin([]byte(Encode(encTx)))

	// Sha of the concatenated uncles
	usha := Sha256Bin(Encode(block.uncles))
	// The block header
	header := block.header(tsha, usha)

	// Encode a slice interface which contains the header and the list of
	// transactions.
	return Encode([]interface{}{header, encTx, block.uncles})
}

func (block *Block) UnmarshalRlp(data []byte) {
	decoder := NewRlpDecoder(data)

	header := decoder.Get(0)
	block.number = uint32(header.Get(0).AsUint())
	block.prevHash = header.Get(1).AsString()
	// sha of uncles is header[2]
	block.coinbase = header.Get(3).AsString()
	block.state = NewTrie(Config.Db, header.Get(4).AsString())
	block.difficulty = header.Get(5).AsBigInt()
	block.time = int64(header.Get(6).AsUint())
	block.nonce = header.Get(7).AsBigInt()
	block.extra = header.Get(8).AsString()

	txes := decoder.Get(1)
	block.transactions = make([]*Transaction, txes.Length())
	for i := 0; i < txes.Length(); i++ {
		tx := &Transaction{}
		tx.UnmarshalRlp(txes.Get(i).AsBytes())
		block.transactions[i] = tx
	}
}


//////////// UNEXPORTED /////////////////
func (block *Block) header(txSha []byte, uncleSha []byte) []interface{} {
	return []interface{}{
		// The block number
		block.number,
		// Sha of the previous block
		block.prevHash,
		// Sha of uncles
		uncleSha,
		// Coinbase address
		block.coinbase,
		// root state
		block.state.Root,
		// Sha of tx
		txSha,
		// Current block difficulty
		block.difficulty,
		// Time the block was found?
		uint64(block.time),
		// Block's nonce for validation
		block.nonce,
		// Extra (unused)
		block.extra,
	}
}
