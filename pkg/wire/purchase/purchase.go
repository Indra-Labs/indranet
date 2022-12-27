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
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "pc"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + slice.Uint64Len + sha256.Len*3
	_           types.Onion = &OnionSkin{}
)

// OnionSkin purchase is a message that requests a session key, which will activate
// when a payment for it has been done, or it will time out after some period to
// allow unused codes to be flushed.
type OnionSkin struct {
	NBytes uint64
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return MinLen + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	value := slice.NewUint64()
	slice.EncodeUint64(value, x.NBytes)
	copy(b[*c:c.Inc(slice.Uint64Len)], value)
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[0][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[1][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[2][:])
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			MinLen-magicbytes.Len, MagicString)
	}
	x.NBytes = slice.DecodeUint64(
		b[*c:c.Inc(slice.Uint64Len)])
	for i := range x.Ciphers {
		bytes := b[*c:c.Inc(sha256.Len)]
		copy(x.Ciphers[i][:], bytes)
		bytes.Zero()
	}
	return
}
