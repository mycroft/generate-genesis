package main

import (
	"bytes"
	"encoding/binary"
)

type Block struct {
	VersionNumber uint32
	PreviousHash  []byte
	Hash          []byte
	MerkleRoot    []byte
	Timestamp     uint32
	Bits          uint32
	Nonce         uint32
	TxCount       int
	Txs           []*Transaction
	Size          int64
}

func (blk *Block) Serialize() []byte {
	var out bytes.Buffer

	uint32buffer := make([]byte, 4)

	binary.LittleEndian.PutUint32(uint32buffer, blk.VersionNumber)
	out.Write(uint32buffer)

	out.Write(blk.PreviousHash)
	out.Write(blk.MerkleRoot)

	binary.LittleEndian.PutUint32(uint32buffer, blk.Timestamp)
	out.Write(uint32buffer)

	binary.LittleEndian.PutUint32(uint32buffer, blk.Bits)
	out.Write(uint32buffer)

	binary.LittleEndian.PutUint32(uint32buffer, blk.Nonce)
	out.Write(uint32buffer)

	return out.Bytes()
}

func CreateBlock(params *GenesisParams) *Block {
	blk := new(Block)

	tx := CreateTransaction(
		params.Psz,
		params.Coins,
		params.Pubkey,
	)

	tx_hash := ComputeSha256(ComputeSha256(tx.Serialize()))
	blk.MerkleRoot = tx_hash
	tx.Hash = tx_hash

	blk.VersionNumber = 1
	blk.PreviousHash = make([]byte, 32)
	blk.MerkleRoot = tx_hash

	blk.Timestamp = params.Timestamp
	blk.Nonce = params.Nonce
	blk.Bits = params.Bits

	blk.Txs = append(blk.Txs, tx)

	blk.Hash = ComputeSha256(ComputeSha256(blk.Serialize()))

	return blk
}

// https://en.bitcoin.it/wiki/Block_hashing_algorithm
func ComputeBlockHash(blk *Block) []byte {
	return ComputeSha256(ComputeSha256(blk.Serialize()))
}
