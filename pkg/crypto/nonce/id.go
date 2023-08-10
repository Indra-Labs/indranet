package nonce

import (
	"crypto/rand"
	"encoding/base32"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/util/b32"
	"sync"
)

const IDLen = 8

var (
	counter uint16
	// enc is a raw base32 encoder as IDs have a consistent set of extraneous
	// characters after 13 digits and do not need check bytes as they are compact
	// large numbers used as collision resistant nonces to identify items in lists.
	enc  = base32.NewEncoding(b32.Based32Ciphers).EncodeToString
	idMx sync.Mutex
	seed sha256.Hash
)

// ID is a value generated by the first 8 bytes truncated from the values of a
// hash chain that reseeds from a CSPRNG at first use and every time it
// generates 2^16 (65536) new ID's.
type ID [IDLen]byte

// NewID returns a random 8 byte nonce to be used as identifiers.
func NewID() (t ID) {
	idMx.Lock()
	defer idMx.Unlock()
	if counter == 0 {
		// We reseed when the counter value overflows.
		reseed()
	}
	s := sha256.Single(seed[:])
	copy(seed[:], s[:])
	copy(t[:], seed[:IDLen])
	counter++
	return
}

// String encodes the ID using Based32.
func (id ID) String() string {
	return enc(id[:])[:13]
}

func reseed() {
	if c, e := rand.Read(seed[:]); fails(e) && c != IDLen {
	}
	counter++
}
