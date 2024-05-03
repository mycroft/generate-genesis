package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Transaction struct {
	Size          uint32
	Hash          []byte
	VersionNumber uint32
	InputCount    int
	Input         []TransactionInput
	OutputCount   int
	Output        []TransactionOutput
	LockTime      uint32
}

type TransactionInput struct {
	TxHash         []byte
	TxIndex        uint32
	ScriptLength   int
	Script         []byte
	SequenceNumber uint32
}

type TransactionOutput struct {
	Value        uint64
	ScriptLength int
	Script       []byte
}

func CreateInputScript(psz string) []byte {
	// Signature script (coinbase)
	prefix := []byte{0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04}

	if len(psz) >= 0x4c && len(psz) <= 0xff {
		prefix = append(prefix, byte(0x4c)) // OP_PUSHDATA1
	} else if len(psz) > 0xff {
		panic("Script length is too long")
	}

	prefix = append(prefix, byte(len(psz)))
	prefix = append(prefix, []byte(psz)...)

	return prefix
}

func CreateOutputScript(pubkey_hex string) []byte {
	var script bytes.Buffer

	decoded_len := hex.DecodedLen(len(pubkey_hex))
	pubkey_decoded := make([]byte, decoded_len)

	n, err := hex.Decode(pubkey_decoded, []byte(pubkey_hex))
	if err != nil {
		panic(err)
	}

	if n != 65 && n != 33 {
		fmt.Printf("Warning: Pubkey is not 33 or 65 char long. Are you sure it is a valid ecdsa key?\n")
	}

	script.WriteByte(byte(n))
	script.Write(pubkey_decoded)
	script.WriteByte(0xac) // OP_CHECKSIG

	return script.Bytes()
}

func CreateTransaction(psz string, coins uint64, pubkey_hex string) *Transaction {
	tx := new(Transaction)

	tx.VersionNumber = 1

	inputScript := CreateInputScript(psz)

	tx.InputCount = 1
	tx.Input = append(tx.Input, TransactionInput{
		TxHash:         make([]byte, 32),
		TxIndex:        uint32(0xffffffff),
		ScriptLength:   len(inputScript),
		Script:         inputScript,
		SequenceNumber: 0xffffffff,
	})

	outputScript := CreateOutputScript(pubkey_hex)

	tx.OutputCount = 1
	tx.Output = append(tx.Output, TransactionOutput{
		Value:        coins,
		Script:       outputScript,
		ScriptLength: len(outputScript),
	})

	return tx
}

func (tx *Transaction) ComputeHash() []byte {
	tx.Hash = ComputeSha256(ComputeSha256(tx.Serialize()))
	return tx.Hash
}

func (tx_input *TransactionInput) Serialize() []byte {
	var out bytes.Buffer

	uint32buff := make([]byte, 4)

	out.Write(tx_input.TxHash)

	binary.LittleEndian.PutUint32(uint32buff, tx_input.TxIndex)
	out.Write(uint32buff)
	out.WriteByte(byte(tx_input.ScriptLength))
	out.Write(tx_input.Script)

	binary.LittleEndian.PutUint32(uint32buff, tx_input.SequenceNumber)
	out.Write(uint32buff)

	return out.Bytes()
}

func (tx_output *TransactionOutput) Serialize() []byte {
	var out bytes.Buffer
	uint64buff := make([]byte, 8)

	binary.LittleEndian.PutUint64(uint64buff, tx_output.Value)
	out.Write(uint64buff)
	out.WriteByte(byte(tx_output.ScriptLength))
	out.Write(tx_output.Script)

	return out.Bytes()
}

func (tx *Transaction) Serialize() []byte {
	var out bytes.Buffer

	uint32buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(uint32buff, tx.VersionNumber)

	out.Write(uint32buff)              // Version
	out.WriteByte(byte(tx.InputCount)) // Input count (1)
	out.Write(tx.Input[0].Serialize()) // Input

	out.WriteByte(byte(tx.OutputCount)) // Output count
	out.Write(tx.Output[0].Serialize()) // Input

	binary.LittleEndian.PutUint32(uint32buff, tx.LockTime)
	out.Write(uint32buff) // Locktime

	return out.Bytes()
}
