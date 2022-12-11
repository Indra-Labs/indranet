// Package ciph manages encryption ciphers and encrypting blobs of data.
package ciph

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/ecdh"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// GetBlock returns a block cipher with a secret generated from the provided
// keys using ECDH.
func GetBlock(from *prv.Key, to *pub.Key) (block cipher.Block) {
	secret := ecdh.Compute(from, to)
	block, _ = aes.NewCipher(secret[:])
	return
}

// BlockFromHash creates an AES block cipher from an sha256.Hash
func BlockFromHash(h sha256.Hash) (block cipher.Block) {
	block, _ = aes.NewCipher(h[:])
	return
}

// Encipher XORs the data with the block stream. This encrypts unencrypted data
// and decrypts encrypted data. If the cipher.Block is nil, it panics (this
// should never happen).
func Encipher(blk cipher.Block, n nonce.IV, b []byte) {
	if blk == nil {
		panic("Encipher called without a block cipher provided")
	} else {
		cipher.NewCTR(blk, n[:]).XORKeyStream(b, b)
	}
}
