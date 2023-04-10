package crypto

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"sync"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/b32/based32"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// ComputeSharedSecret computes an Elliptic Curve Diffie-Hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func ComputeSharedSecret(prv *Prv, pub *Pub) sha256.Hash {
	return sha256.Single(
		secp256k1.GenerateSharedSecret(
			(*secp256k1.PrivateKey)(prv), (*secp256k1.PublicKey)(pub),
		),
	)
}

const (
	BlindLen = 3
	HashLen  = 5
	Len      = BlindLen + HashLen
)

func (c PubKey) CopyBlinder() (blinder Blinder) {
	copy(blinder[:], c[:BlindLen])
	return
}

// PubKey is the blinded hash of a public key used to conceal a message public
// key from attackers.
type PubKey [Len]byte

type Blinder [BlindLen]byte
type Hash [HashLen]byte

// GetCloak returns a value which a receiver with the private key can identify
// the association of a message with the peer in order to retrieve the private
// key to generate the message cipher.
//
// The three byte blinding factor concatenated in front of the public key
// generates the 5 bytes at the end of the PubKey code. In this way the source
// public key it relates to is hidden to any who don't have this public key,
// which only the parties know.
func GetCloak(s *Pub) (c PubKey) {
	var blinder Blinder
	var n int
	var e error
	if n, e = rand.Read(blinder[:]); fails(e) && n != BlindLen {
		panic("no entropy")
	}
	c = Cloak(blinder, s.ToBytes())
	return
}

func Cloak(b Blinder, key PubBytes) (c PubKey) {
	h := sha256.Single(append(b[:], key[:]...))
	copy(c[:BlindLen], b[:BlindLen])
	copy(c[BlindLen:BlindLen+HashLen], h[:HashLen])
	return
}

// Match uses the cached public key and the provided blinding factor to match
// the source public key so the packet address field is only recognisable to the
// intended recipient.
func Match(r PubKey, k PubBytes) bool {
	var b Blinder
	copy(b[:], r[:BlindLen])
	hash := Cloak(b, k)
	return r == hash
}

const (
	PrvKeyLen = secp256k1.PrivKeyBytesLen
)

// Prv is a private key.
type Prv secp256k1.PrivateKey
type PrvBytes [PrvKeyLen]byte

// GeneratePrvKey a private key.
func GeneratePrvKey() (prv *Prv, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); fails(e) {
		return
	}
	return (*Prv)(p), e
}

func (p *Prv) ToBase32() (s string) {
	b := p.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); fails(e) {
	}
	ss := []byte(s[1:])
	return string(ss)
}

func PrvFromBase32(s string) (k *Prv, e error) {
	ss := []byte(s)
	var b slice.Bytes
	b, e = based32.Codec.Decode("a" + string(ss))
	k = PrvKeyFromBytes(b)
	return
}

// PrvKeyFromBytes converts a byte slice into a private key.
func PrvKeyFromBytes(b []byte) *Prv {
	return (*Prv)(secp256k1.PrivKeyFromBytes(b))
}

// Zero out a private key to prevent key scraping from memory.
func (p *Prv) Zero() { (*secp256k1.PrivateKey)(p).Zero() }

// ToBytes returns the Bytes serialized form. It zeroes the original bytes.
func (p *Prv) ToBytes() (b PrvBytes) {
	br := (*secp256k1.PrivateKey)(p).Serialize()
	copy(b[:], br[:PrvKeyLen])
	// // zero the original
	// copy(br, zeroPrv())
	return
}

func zeroPrv() []byte {
	z := PrvBytes{}
	return z[:]
}

// Zero zeroes out a private key in serial form.
func (pb PrvBytes) Zero() { copy(pb[:], zeroPrv()) }

const (
	// PubKeyLen is the length of the serialized key. It is an ECDSA compressed
	// key.
	PubKeyLen = secp256k1.PubKeyBytesLenCompressed
)

var enc = base32.NewEncoding(Charset).EncodeToString

const Charset = "abcdefghijklmnopqrstuvwxyz234679"

type (
	// Pub is a public key.
	Pub secp256k1.PublicKey
	// PubBytes is the serialised form of a public key.
	PubBytes [PubKeyLen]byte
)

func (pb PubBytes) String() (s string) {
	var e error
	if s, e = based32.Codec.Encode(pb[:]); fails(e) {
	}
	ss := []byte(s)
	// Reverse text order to get all starting ciphers.
	for i := 0; i < len(s)/2; i++ {
		ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	}
	return color.LightGreen.Sprint(string(ss))
}

func (k *Pub) String() (s string) {
	return k.ToBase32()
}

// DerivePub generates a public key from the prv.Pub.
func DerivePub(prv *Prv) *Pub {
	if prv == nil {
		return nil
	}
	return (*Pub)((*secp256k1.PrivateKey)(prv).PubKey())
}

// PubFromBytes converts a byte slice into a public key, if it is valid and on the
// secp256k1 elliptic curve.
func PubFromBytes(b []byte) (pub *Pub, e error) {
	var p *secp256k1.PublicKey
	if p, e = secp256k1.ParsePubKey(b); fails(e) {
		return
	}
	pub = (*Pub)(p)
	return
}

// ToBytes returns the compressed 33 byte form of the pubkey as used in wire and
// storage forms.
func (k *Pub) ToBytes() (p PubBytes) {
	b := (*secp256k1.PublicKey)(k).SerializeCompressed()
	copy(p[:], b)
	return
}

func (k *Pub) ToHex() (s string, e error) {
	b := k.ToBytes()
	s = hex.EncodeToString(b[:])
	return
}

func (k *Pub) ToBase32() (s string) {
	b := k.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); fails(e) {
	}
	ss := []byte(s)[3:]
	// // Reverse text order to get all starting ciphers.
	// for i := 0; i < len(s)/2; i++ {
	// 	ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	// }
	return string(ss)
}

func (k *Pub) ToBase32Abbreviated() (s string) {
	s = k.ToBase32()
	s = s[:13] + "..." + s[len(s)-8:]
	return color.LightGreen.Sprint(string(s))
}

func PubFromBase32(s string) (k *Pub, e error) {
	ss := []byte(s)
	var b slice.Bytes
	b, e = based32.Codec.Decode("ayb" + string(ss))
	return PubFromBytes(b)
}

func (pb PubBytes) Equals(qb PubBytes) bool { return pb == qb }

func (k *Pub) ToPublicKey() *secp256k1.PublicKey {
	return (*secp256k1.PublicKey)(k)
}

// Equals returns true if two public keys are the same.
func (k *Pub) Equals(pub2 *Pub) bool {
	return k.ToPublicKey().IsEqual(pub2.ToPublicKey())
}

// SigLen is the length of the signatures used in Indra, compact keys that can have
// the public key extracted from them.
const SigLen = 65

// SigBytes is an ECDSA BIP62 formatted compact signature which allows the recovery
// of the public key from the signature.
type SigBytes [SigLen]byte

// Sign produces an ECDSA BIP62 compact signature.
func Sign(prv *Prv, hash sha256.Hash) (sig SigBytes, e error) {
	copy(sig[:],
		ecdsa.SignCompact((*secp256k1.PrivateKey)(prv), hash[:], true))
	return
}

// Recover the public key corresponding to the signing private key used to
// create a signature on the hash of a message.
func (sig SigBytes) Recover(hash sha256.Hash) (p *Pub, e error) {
	var pk *secp256k1.PublicKey
	// We are only using compressed keys, so we can ignore the compressed
	// bool.
	if pk, _, e = ecdsa.RecoverCompact(sig[:], hash[:]); !fails(e) {
		p = (*Pub)(pk)
	}
	return
}

type KeySet struct {
	sync.Mutex
	Base, Increment *Prv
}

// NewSigner creates a new KeySet which enables (relatively) fast generation of new
// private keys by using scalar addition.
func NewSigner() (first *Prv, ks *KeySet, e error) {
	ks = &KeySet{}
	if ks.Base, e = GeneratePrvKey(); fails(e) {
		return
	}
	if ks.Increment, e = GeneratePrvKey(); fails(e) {
		return
	}
	first = ks.Base
	return
}

// Next adds Increment to Base, assigns the new value to the Base and returns
// the new value.
func (ks *KeySet) Next() (n *Prv) {
	ks.Mutex.Lock()
	next := ks.Base.Key.Add(&ks.Increment.Key)
	ks.Base.Key = *next
	n = &Prv{Key: *next}
	ks.Mutex.Unlock()
	return
}

func (ks *KeySet) Next3() (n [3]*Prv) {
	for i := range n {
		n[i] = ks.Next()
	}
	return
}

func (ks *KeySet) Next2() (n [2]*Prv) {
	for i := range n {
		n[i] = ks.Next()
	}
	return
}