package sig

import (
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Len is the length of the signatures used in Indra, compact keys that can have
// the public key extracted from them, thus eliminating the need to separately
// specify it in messages.
const Len = 65

// Bytes is an ECDSA BIP62 formatted compact signature which allows the recovery
// of the public key from the signature. This allows messages to avoid adding
// extra bytes to also specify the public key of the signer.
type Bytes []byte

func New() Bytes { return make(Bytes, Len) }

// IsValid checks that the signature is the correct length. This avoids needing
// to copy into a static array. Static arrays save on this code because they
// automatically must be correct.
func (sig Bytes) IsValid() (e error) {
	if len(sig) == Len {
		return
	}
	return fmt.Errorf(
		"signature incorrect length, expect %d, got %d",
		Len, len(sig))
}

// FromBytes checks if signature bytes are the correct length to be a signature.
func FromBytes(sig Bytes) (e error) { return sig.IsValid() }

// Sign produces an ECDSA BIP62 compact signature.
func Sign(prv *prv.Key, hash sha256.Hash) (sig Bytes, e error) {
	if check(hash.Valid()) {
		return
	}
	sig = ecdsa.SignCompact((*secp256k1.PrivateKey)(prv), hash, true)
	return
}

// Recover the public key corresponding to the signing private key used to
// create a signature on the hash of a message.
func (sig Bytes) Recover(hash sha256.Hash) (p *pub.Key, e error) {
	var pk *secp256k1.PublicKey
	// We are only using compressed keys, so we can ignore the compressed
	// bool.
	if pk, _, e = ecdsa.RecoverCompact(sig, hash); check(e) {
		p = (*pub.Key)(pk)
	}
	return
}
