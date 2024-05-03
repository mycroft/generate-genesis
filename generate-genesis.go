package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"strconv"

	quark "github.com/mycroft/goquarkhash"
	x11 "gitlab.com/nitya-sattva/go-x11"
	"golang.org/x/crypto/scrypt"
)

var (
	algo             string
	psz              string
	coins            uint64
	pubkey           string
	timestamp, nonce uint
	bits             string
	profile          string
	threads          int
	verbose          bool
)

func init() {
	flag.StringVar(&algo, "algo", "sha256", "Algo to use: sha256, scrypt, x11, quark")
	flag.StringVar(&psz, "psz", "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks", "pszTimestamp")
	flag.Uint64Var(&coins, "coins", uint64(50*100000000), "Number of coins")
	flag.StringVar(&pubkey, "pubkey", "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f", "Pubkey (required)")
	flag.UintVar(&timestamp, "timestamp", 1231006505, "Timestamp to use")
	flag.UintVar(&nonce, "nonce", 2083236893, "Nonce value")
	flag.StringVar(&bits, "bits", "1d00ffff", "Bits")
	flag.StringVar(&profile, "profile", "", "Write profile information into file (debug)")
	flag.IntVar(&threads, "threads", 4, "Number of threads to use")
	flag.BoolVar(&verbose, "verbose", false, "Show some messages")
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

func ComputeX11(content []byte) []byte {
	out := make([]byte, 32)

	hasher := x11.New()
	hasher.Hash(content, out)

	return out
}

func ComputeQuark(content []byte) []byte {
	return quark.QuarkHash(content)
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

type Job struct {
	StartingNonce uint32
	MaxNonce      uint32
	Timestamp     uint32
}

func main() {
	flag.Parse()

	if profile != "" {
		f, err := os.Create(profile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if psz == "" {
		fmt.Printf("Require a psz. Please set -psz")
		os.Exit(1)
	}

	jobs_num := threads
	jobs := make(chan Job, jobs_num)
	results := make(chan bool, jobs_num)

	for i := 0; i < jobs_num; i++ {
		go SearchWorker(i, jobs, results)
	}

	nonce_current := uint32(nonce)
	nonce_iterator := uint32(1024000)

	for {
		var res bool
		if jobs_num > 0 {
			next_max_nonce := nonce_current + nonce_iterator
			jobs <- Job{
				StartingNonce: nonce_current,
				MaxNonce:      next_max_nonce,
				Timestamp:     uint32(timestamp),
			}
			if next_max_nonce < nonce_current {
				timestamp++
				fmt.Println("nonce was reset. Timestamp is now", timestamp)
			}
			nonce_current = next_max_nonce

			jobs_num--
		} else if jobs_num == 0 {
			// Wait for a job to be completed
			res = <-results
			jobs_num++
		}

		if res {
			break
		}
	}
}

func SearchWorker(instance int, jobs <-chan Job, results chan<- bool) {
	var current big.Int
	var found bool

	bits_uint32, err := strconv.ParseUint(bits, 16, 32)
	if err != nil {
		panic(err)
	}

	target := ComputeTarget(uint32(bits_uint32))

	for job := range jobs {
		if verbose {
			fmt.Printf(
				"Worker %2d: Nonce: %10d to Nonce: %10d timestamp: %10d\n",
				instance,
				job.StartingNonce,
				job.MaxNonce,
				job.Timestamp,
			)
		}
		params := new(GenesisParams)
		params.Algo = algo
		params.Psz = psz
		params.Coins = coins
		params.Pubkey = pubkey
		params.Timestamp = job.Timestamp
		params.Nonce = job.StartingNonce
		params.Bits = uint32(bits_uint32)

		tx := CreateTransaction(
			params.Psz,
			params.Coins,
			params.Pubkey,
		)
		tx.ComputeHash()

		blk := CreateBlock(params, tx)

		for {
			switch params.Algo {
			case "sha256":
				blk.ComputeHash()
			case "scrypt":
				blk.Hash = ComputeScrypt(blk.Serialize())
			case "x11":
				blk.Hash = ComputeX11(blk.Serialize())
			case "quark":
				blk.Hash = quark.QuarkHash(blk.Serialize())
			}

			current.SetBytes(Reverse(blk.Hash))
			if 1 == target.Cmp(&current) {
				found = true
				PrintFound(blk, blk.Hash, target)
				break
			}

			blk.Nonce++
			if blk.Nonce == job.MaxNonce {
				break
			}
		}

		if found == true {
			results <- true
			break
		}

		results <- false
	}

}

func PrintFound(blk *Block, hash []byte, target big.Int) {
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
