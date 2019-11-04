Genesis block generator
=======================

Introduction
------------

This tool provides a convenient way to generate Genesis block for bitcoin-clone (altcoin) crypto currencies.


Build
-----

```shell
$ go get -d -v
$ go build
$ go test
```

Usage
-----

```shell
$ ./generate-genesis -h
Usage of ./generate-genesis:
  -algo string
        Algo to use: sha256, scrypt, x11, quark (default "sha256")
  -bits string
        Bits (default "1d00ffff")
  -coins uint
        Number of coins (default 5000000000)
  -nonce uint
        Nonce value (default 2083236893)
  -profile string
        Write profile information into file (debug)
  -psz string
        pszTimestamp (default "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")
  -pubkey string
        Pubkey (required) (default "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f")
  -threads int
        Number of threads to use (default 4)
  -timestamp uint
        Timestamp to use (default 1231006505)
```

Samples
-------

### Bitcoin (default)

```shell
$ ./generate-genesis
Ctrl Hash:      0x000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
Target:         0x00000000ffff0000000000000000000000000000000000000000000000000000
Blk Hash:       0x000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
Mkl Hash:       0x4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
Nonce:          2083236893
Timestamp:      1231006505
Pubkey:         04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f
Coins:          5000000000
Psz:            'The Times 03/Jan/2009 Chancellor on brink of second bailout for banks'
```

### Litecoin

```shell
$ ./generate-genesis -algo scrypt -bits 1e0ffff0 -coins 5000000000 -nonce 2084524480 -timestamp 1317972665 -pubkey 040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9 -psz "NY Times 05/Oct/2011 Steve Jobs, Apple’s Visionary, Dies at 56"
Ctrl Hash:      0x0000050c34a64b415b6b15b37f2216634b5b1669cb9a2e38d76f7213b0671e00
Target:         0x00000ffff0000000000000000000000000000000000000000000000000000000
Blk Hash:       0x0a2efd19744ffdff263e7223faf3a212c6040acfdc03d6b09f3e1ed1dd6f8272
Mkl Hash:       0x97ddfbbae6be97fd6cdf3e7ca13232a3afff2353e29badfab7f73011edd4ced9
Nonce:          2084524493
Timestamp:      1317972665
Pubkey:         040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9
Coins:          5000000000
Psz:            'NY Times 05/Oct/2011 Steve Jobs, Apple’s Visionary, Dies at 56'
```

### Dash (x11)

```shell
$ ./generate-genesis -algo x11 -bits 1e0ffff0 -coins 5000000000 -psz "Wired 09/Jan/2014 The Grand Experiment Goes Live: Overstock.com Is Now Accepting Bitcoins" -pubkey "040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9" -timestamp 1390095618 -nonce 28917600
Ctrl Hash:      0x00000ffd590b1485b3caadc19b22e6379c733355108f107a430458cdf3407ab6
Target:         0x00000ffff0000000000000000000000000000000000000000000000000000000
Blk Hash:       0x00000ffd590b1485b3caadc19b22e6379c733355108f107a430458cdf3407ab6
Mkl Hash:       0xe0028eb9648db56b1ac77cf090b99048a8007e2bb64b68f092c03c7f56a662c7
Nonce:          28917698
Timestamp:      1390095618
Pubkey:         040184710fa689ad5023690c80f3a49c8f13f8d45b8c857fbcbc8bc4a8e4d3eb4b10f4d4604fa08dce601aaf0f470216fe1b51850b4acf21b179c45070ac7b03a9
Coins:          5000000000
Psz:            'Wired 09/Jan/2014 The Grand Experiment Goes Live: Overstock.com Is Now Accepting Bitcoins'
```
