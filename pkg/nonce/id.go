package nonce

import (
	"crypto/rand"

	"github.com/Indra-Labs/indra/pkg/sha256"
)

const IDLen = 8

type ID [IDLen]byte

var seed [sha256.Len]byte
var counter uint16

func reseed() {
	var c int
	var e error
	if c, _ = rand.Read(seed[:]); check(e) && c != IDLen {
		panic(e)
	}
	counter++
}

func NewID() (t ID) {
	if counter == 0 {
		reseed()
	}
	copy(t[:], seed[:IDLen])
	s := sha256.Single(seed[:])
	copy(seed[:], s)
	return
}
