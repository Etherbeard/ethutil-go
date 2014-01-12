package ethutil

// TODO I think this should be moved elsewhere

import (
)

/*
 * This is the special genesis block.
 */

var GenisisHeader = []interface{}{
	// Previous hash (none)
	"",
	// Sha of uncles
	string(Sha256Bin(Encode([]interface{}{}))),
	// Coinbase
	"",
	// Root state
	"",
	// Sha of transactions
	string(Sha256Bin(Encode([]interface{}{}))),
	// Difficulty
	BigPow(2, 26),
	// Time
	uint64(1),
	// Nonce
	Big("0"),
	// Extra
	"",
}

var Genesis = []interface{}{GenisisHeader, []interface{}{}, []interface{}{}}
