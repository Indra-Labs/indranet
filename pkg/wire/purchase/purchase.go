package purchase

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Type purchase is a message that requests a session key, which will activate
// when a payment for it has been done, or it will time out after some period to
// allow unused codes to be flushed.
type Type struct {
	NBytes uint64
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	types.Onion
}

var (
	Magic              = slice.Bytes("prc")
	MinLen             = magicbytes.Len + slice.Uint64Len + sha256.Len*3
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return MinLen + x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	value := slice.NewUint64()
	slice.EncodeUint64(value, x.NBytes)
	x.Onion.Encode(o, c)
}

func (x *Type) Decode(b slice.Bytes) (e error) {

	magic := Magic
	if !magicbytes.CheckMagic(b, magic) {
		return magicbytes.WrongMagic(x, b, magic)
	}
	minLen := MinLen
	if len(b) < minLen {
		return magicbytes.TooShort(len(b), minLen, string(magic))
	}
	sc := slice.Cursor(0)
	c := &sc
	_ = c

	return
}