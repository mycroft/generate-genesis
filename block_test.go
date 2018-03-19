package main

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
)

type GenesisTest struct {
	Params     GenesisParams
	MerkleRoot string
	BlockHash  string
}

func CheckHash(t *testing.T, expected string, current []byte) {
	decoded_len := hex.DecodedLen(len(expected))
	decoded := make([]byte, decoded_len)
	_, err := hex.Decode(decoded, []byte(expected))
	if err != nil {
		panic(err)
	}

	decoded = Reverse(decoded)

	if !bytes.Equal(decoded, current) {
		t.Errorf("Invalid hash: expected(0x%x)\n\t\t differs current(0x%x)", decoded, current)
	}
}

func TestGeneration(t *testing.T) {
	tests := []GenesisTest{
		GenesisTest{ // this is official bitcoin
			Params: GenesisParams{
				Algo:      "sha256",
				Psz:       "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks",
				Coins:     50 * 100000000,
				Pubkey:    "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f",
				Timestamp: 1231006505,
				Nonce:     2083236893,
				Bits:      0x1d00ffff,
			},
			MerkleRoot: "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			BlockHash:  "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		},
		GenesisTest{ // this is official litecoin
			Params: GenesisParams{
				Algo:      "scrypt",
				Psz:       "NY Times 05/Oct/2011 Steve Jobs, Apple’s Visionary, Dies at 56",
				Coins:     50 * 100000000,
				Pubkey:    "040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9",
				Timestamp: 1317972665,
				Nonce:     2084524493,
				Bits:      0x1e0ffff0,
			},
			MerkleRoot: "97ddfbbae6be97fd6cdf3e7ca13232a3afff2353e29badfab7f73011edd4ced9",
			BlockHash:  "12a765e31ffd4059bada1e25190f6e98c99d9714d334efa41a195a7e7e04bfe2",
		},
		GenesisTest{ // this is bitcoin testnet
			Params: GenesisParams{
				Algo:      "sha256",
				Psz:       "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks",
				Coins:     50 * 100000000,
				Pubkey:    "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f",
				Timestamp: 1296688602,
				Nonce:     414098458,
				Bits:      0x1d00ffff,
			},
			MerkleRoot: "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			BlockHash:  "000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943",
		},
		GenesisTest{ // dash
			Params: GenesisParams{
				Algo:      "x11",
				Psz:       "Wired 09/Jan/2014 The Grand Experiment Goes Live: Overstock.com Is Now Accepting Bitcoins",
				Coins:     50 * 100000000,
				Pubkey:    "040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9",
				Timestamp: 1390095618,
				Nonce:     28917698,
				Bits:      0x1e0ffff0,
			},
			MerkleRoot: "e0028eb9648db56b1ac77cf090b99048a8007e2bb64b68f092c03c7f56a662c7",
			BlockHash:  "00000ffd590b1485b3caadc19b22e6379c733355108f107a430458cdf3407ab6",
		},
	}

	for _, test := range tests {
		var current big.Int
		var hash []byte

		blk := CreateBlock(&test.Params)

		if test.Params.Algo == "x11" {
			hash = ComputeX11(blk.Serialize())
			CheckHash(t, test.BlockHash, hash)
		} else {
			CheckHash(t, test.BlockHash, blk.Hash)
		}

		CheckHash(t, test.MerkleRoot, blk.MerkleRoot)

		// Check difficulty as well
		target := ComputeTarget(test.Params.Bits)

		switch test.Params.Algo {
		case "sha256":
			hash = ComputeSha256(ComputeSha256(blk.Serialize()))
		case "scrypt":
			hash = ComputeScrypt(blk.Serialize())
		case "x11":
			hash = ComputeX11(blk.Serialize())
		}

		current.SetBytes(Reverse(hash))
		if 1 != target.Cmp(&current) {
			t.Error("Target not reached.")
		}
	}
}
