package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	quark "github.com/mycroft/goquarkhash"
	x11 "gitlab.com/nitya-sattva/go-x11"
	"golang.org/x/crypto/scrypt"
)

var (
	algo              string
	psz               string
	coins             uint64
	pubkey            string
	timestamp, nonce  uint
	bits              string
	profile           string
	maxprocs, workers int
	stepsize          int
	verbose           bool

	// Parsed from bits by validateFlags, read by the workers.
	bits_value uint32
)

// Algos accepted by -algo. Every entry must be handled by the switch in SearchWorker.
var algos = []string{"sha256", "scrypt", "x11", "quark"}

func IsKnownAlgo(algo string) bool {
	for _, known := range algos {
		if algo == known {
			return true
		}
	}

	return false
}

func init() {
	flag.StringVar(&algo, "algo", "sha256", fmt.Sprintf("Algo to use: %s", strings.Join(algos, ", ")))
	flag.StringVar(&psz, "psz", "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks", "pszTimestamp")
	flag.Uint64Var(&coins, "coins", uint64(50*100000000), "Number of coins")
	flag.StringVar(&pubkey, "pubkey", "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f", "Pubkey (required)")
	flag.UintVar(&timestamp, "timestamp", 1231006505, "Timestamp to use")
	flag.UintVar(&nonce, "nonce", 2083236893, "Nonce value")
	flag.StringVar(&bits, "bits", "1d00ffff", "Bits")
	flag.StringVar(&profile, "profile", "", "Write profile information into file (debug)")
	flag.IntVar(&workers, "workers", 0, "Number of workers (goroutine) to use (if unset, use the CPU numbers)")
	flag.IntVar(&maxprocs, "maxprocs", 0, "Number of max CPUs that are simultaneously used")
	flag.IntVar(&stepsize, "stepsize", 1024*1000, "Number of hashes computed per worker job")
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

func ComputeTarget(bits uint32) *big.Int {
	var target big.Int

	// Below 4 the mantissa does not fit the target, and above 32 the target exceeds
	// 2^256, which every possible hash would satisfy. validateFlags rejects both for
	// the CLI; this guards the other callers.
	if exponent := bits >> 24; exponent < 4 || exponent > 32 {
		panic(fmt.Sprintf("bits exponent %d is out of the [4, 32] range", exponent))
	}

	target_bytes := make([]byte, bits>>24)
	binary.BigEndian.PutUint32(target_bytes, uint32(bits%(1<<24)<<8))

	target.SetBytes(target_bytes)

	return &target
}

type Job struct {
	StartingNonce uint32
	MaxNonce      uint32
	Timestamp     uint32
}

// validateFlags checks every flag that would otherwise panic or hang deeper in the
// pipeline. It also decodes bits into bits_value for the workers.
func validateFlags() error {
	if psz == "" {
		return fmt.Errorf("require a psz. Please set -psz")
	}

	if !IsKnownAlgo(algo) {
		return fmt.Errorf("unknown algo '%s'. Please set -algo to one of: %s", algo, strings.Join(algos, ", "))
	}

	parsed_bits, err := strconv.ParseUint(bits, 16, 32)
	if err != nil {
		return fmt.Errorf("invalid bits '%s': expected a 32 bit hexadecimal value, eg 1d00ffff", bits)
	}

	if exponent := parsed_bits >> 24; exponent < 4 || exponent > 32 {
		return fmt.Errorf("invalid bits '%s': exponent (high byte) is %d, expected between 4 and 32", bits, exponent)
	}

	bits_value = uint32(parsed_bits)

	if _, err := hex.DecodeString(pubkey); err != nil {
		return fmt.Errorf("invalid pubkey: expected a hexadecimal ecdsa key: %s", err)
	}

	if stepsize <= 0 {
		return fmt.Errorf("invalid stepsize %d: expected greater than 0", stepsize)
	}

	if workers < 0 {
		return fmt.Errorf("invalid workers %d: expected 0 (use the CPU count) or greater", workers)
	}

	if maxprocs < 0 {
		return fmt.Errorf("invalid maxprocs %d: expected 0 (use the default) or greater", maxprocs)
	}

	return nil
}

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if profile != "" {
		f, err := os.Create(profile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if 0 != maxprocs {
		runtime.GOMAXPROCS(maxprocs)
	}

	jobs_num := runtime.NumCPU()
	if workers != 0 {
		jobs_num = workers
	}

	jobs := make(chan Job, jobs_num)
	results := make(chan bool, jobs_num)

	for i := 0; i < jobs_num; i++ {
		go SearchWorker(i, jobs, results)
	}

	nonce_current := uint32(nonce)
	nonce_iterator := uint32(stepsize)

	for {
		var res bool
		if jobs_num > 0 {
			next_max_nonce := nonce_current + nonce_iterator
			wrapped := next_max_nonce < nonce_current

			if wrapped {
				// Clamp the job to the end of the nonce space: MaxNonce 0 is
				// where the worker's counter lands after 0xffffffff. Without
				// the clamp, [0, next_max_nonce) would be searched at the old
				// timestamp and then skipped at the new one.
				next_max_nonce = 0
			}

			jobs <- Job{
				StartingNonce: nonce_current,
				MaxNonce:      next_max_nonce,
				Timestamp:     uint32(timestamp),
			}

			if wrapped {
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

	target := ComputeTarget(bits_value)

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
		params.Bits = bits_value

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
			default:
				// Unreachable: main validates -algo. Guards against an algo
				// being added to algos but not here, which would otherwise
				// leave blk.Hash nil and report a bogus block as found.
				panic(fmt.Sprintf("algo '%s' is not implemented in SearchWorker", params.Algo))
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

func PrintFound(blk *Block, hash []byte, target *big.Int) {
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
