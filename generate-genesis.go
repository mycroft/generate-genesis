package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"golang.org/x/crypto/scrypt"
)

var (
	algo             string
	psz              string
	coins            uint64
	pubkey           string
	timestamp, nonce uint
	bits             string
)

func init() {
	flag.StringVar(&algo, "algo", "sha256", "Algo to use: sha256, scrypt")
	flag.StringVar(&psz, "psz", "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks", "pszTimestamp")
	flag.Uint64Var(&coins, "coins", uint64(50*100000000), "Number of coins")
	flag.StringVar(&pubkey, "pubkey", "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f", "Pubkey (required)")
	flag.UintVar(&timestamp, "timestamp", 1231006505, "Timestamp to use")
	flag.UintVar(&nonce, "nonce", 2083236893, "Nonce value")
	flag.StringVar(&bits, "bits", "1d00ffff", "Bits")
}

func ComputeSha256(content []byte) []byte {
	m := sha256.New()
	m.Write(content)

	return m.Sum(nil)
}

func ComputeScrypt(content []byte) []byte {
	scryptHash, err := scrypt.Key(content, content, 1024, 1, 1, 32)

	if err != nil {
		panic(err)
	}

	return scryptHash
}

func Reverse(in []byte) []byte {
	out := make([]byte, len(in))

	for i := 0; i < len(in); i++ {
		out[i] = in[len(in)-i-1]
	}

	return out
}

type GenesisParams struct {
	Algo      string
	Psz       string
	Coins     uint64
	Pubkey    string
	Timestamp uint32
	Nonce     uint32
	Bits      uint32
}

func ComputeTarget(bits uint32) big.Int {
	var target big.Int

	target_bytes := make([]byte, bits>>24)
	binary.BigEndian.PutUint32(target_bytes, uint32(bits%(1<<24)<<8))

	target.SetBytes(target_bytes)

	return target
}

func main() {
	flag.Parse()

	var hash []byte
	var current big.Int

	if psz == "" {
		fmt.Printf("Require a psz. Please set -psz")
		os.Exit(1)
	}

	bits_uint32, err := strconv.ParseUint(bits, 16, 32)
	if err != nil {
		panic(err)
	}

	params := new(GenesisParams)
	params.Algo = algo
	params.Psz = psz
	params.Coins = coins
	params.Pubkey = pubkey
	params.Timestamp = uint32(timestamp)
	params.Nonce = uint32(nonce)
	params.Bits = uint32(bits_uint32)

	blk := CreateBlock(params)
	target := ComputeTarget(blk.Bits)

	for {
		switch params.Algo {
		case "sha256":
			hash = ComputeSha256(ComputeSha256(blk.Serialize()))
		case "scrypt":
			hash = ComputeScrypt(blk.Serialize())
		}

		current.SetBytes(Reverse(hash))
		if 1 == target.Cmp(&current) {
			break
		}

		blk.Nonce++
		if blk.Nonce == 0x00 {
			blk.Timestamp++
		}
	}

	fmt.Printf("Ctrl Hash:\t0x%x\n", Reverse(hash))
	target_hash := make([]byte, 32)
	copy(target_hash[32-len(target.Bytes()):], target.Bytes())
	fmt.Printf("Target:\t\t0x%x\n", target_hash)
	fmt.Printf("Blk Hash:\t0x%x\n", Reverse(blk.Hash))
	fmt.Printf("Mkl Hash:\t0x%x\n", Reverse(blk.MerkleRoot))
	fmt.Printf("Nonce:\t\t%d\n", blk.Nonce)
	fmt.Printf("Timestamp:\t%d\n", blk.Timestamp)
	fmt.Printf("Pubkey:\t\t%s\n", pubkey)
	fmt.Printf("Coins:\t\t%d\n", coins)
	fmt.Printf("Psz:\t\t'%s'\n", psz)
}
