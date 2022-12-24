package exit

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

// Type exit messages are the layer of a message after two Forward packets that
// provides an exit address and
type Type struct {
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Port uint16
	// Cipher is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Cipher [3]sha256.Hash
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	types.Onion
}

var (
	Magic              = slice.Bytes("exi")
	MinLen             = magicbytes.Len + slice.Uint16Len + 3*sha256.Len
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return MinLen + x.Bytes.Len() +
		x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	port := slice.NewUint16()
	slice.EncodeUint16(port, int(x.Port))
	copy(o[*c:c.Inc(slice.Uint16Len)], port)
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[0][:])
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[1][:])
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[1][:])
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x.Bytes))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(x.Bytes))], x.Bytes)
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