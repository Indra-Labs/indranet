package forward

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Type forward is just an IP address and a wrapper for another message.
type Type struct {
	net.IP
	types.Onion
}

var (
	Magic              = slice.Bytes("fwd")
	MinLen             = magicbytes.Len + 1 + net.IPv4len
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + len(x.IP) + 1 + x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	o[*c] = byte(len(x.IP))
	copy(o[c.Inc(1):c.Inc(len(x.IP))], x.IP)
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