package pub

import (
	"encoding/hex"
	"unsafe"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type (
	// Key is a public key.
	Key secp256k1.PublicKey
	// Bytes is the serialised form of a public key.
	Bytes []byte
)

const (
	// KeyLen is the length of the serialized key. It is an ECDSA compressed
	// key.
	KeyLen = secp256k1.PubKeyBytesLenCompressed
)

// Derive generates a public key from the prv.Key.
func Derive(prv *prv.Key) *Key {
	return (*Key)((*secp256k1.PrivateKey)(prv).PubKey())
}

// FromBytes converts a byte slice into a public key, if it is valid and on the
// secp256k1 elliptic curve.
func FromBytes(b []byte) (pub *Key, e error) {
	var p *secp256k1.PublicKey
	if p, e = secp256k1.ParsePubKey(b); check(e) {
		return
	}
	pub = (*Key)(p)
	return
}

// ToBytes returns the compressed 33 byte form of the pubkey as used in wire and
// storage forms.
func (pub *Key) ToBytes() (p Bytes) {
	return (*secp256k1.PublicKey)(pub).SerializeCompressed()
}

const HRP = "indra"

func (pub *Key) ToHex() (s string, e error) {
	b := pub.ToBytes()
	s = hex.EncodeToString(b)
	return
}
func (pb Bytes) Equals(qb Bytes) bool {
	// Ensure lengths are correct.
	if len(pb) == KeyLen && len(qb) == KeyLen {
		return *(*string)(unsafe.Pointer(&pb)) ==
			*(*string)(unsafe.Pointer(&qb))
	}
	return false
}

func (pub *Key) ToPublicKey() *secp256k1.PublicKey {
	return (*secp256k1.PublicKey)(pub)
}

// Equals returns true if two public keys are the same.
func (pub *Key) Equals(pub2 *Key) bool {
	return pub.ToPublicKey().IsEqual(pub2.ToPublicKey())
}
