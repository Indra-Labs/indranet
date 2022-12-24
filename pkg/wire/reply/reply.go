package reply

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "rl"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + 1 + net.IPv4len
	_           types.Onion = &OnionSkin{}
)

// OnionSkin is reply messages, distinct from forward.OnionSkin messages in that the
// header encryption uses a different secret than the payload. The magic bytes
// signal this to the relay that receives this, which then looks up the
// PayloadHey matching the To address in the message header. And lastly, each
// step the relay budges up it's message to the front of the packet and puts
// csprng random bytes into the remainder to the same length.
type OnionSkin struct {
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return magicbytes.Len + len(x.IP) + 1 + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	b[*c] = byte(len(x.IP))
	copy(b[c.Inc(1):c.Inc(len(x.IP))], x.IP)
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b[*c:]), MinLen, string(Magic))
	}
	ipLen := b[*c]
	x.IP = net.IP(b[c.Inc(1):c.Inc(int(ipLen))])
	return
}
