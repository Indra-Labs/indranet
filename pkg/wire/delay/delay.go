package delay

import (
	"time"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "dl"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + slice.Uint64Len
	_           types.Onion = &OnionSkin{}
)

// A OnionSkin is a 32 byte value.
type OnionSkin struct {
	time.Duration
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(_ types.Onion) {}
func (x *OnionSkin) Len() int             { return MinLen }

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	slice.EncodeUint64(b[*c:c.Inc(slice.Uint64Len)], uint64(x.Duration))
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), MinLen-magicbytes.Len,
			string(Magic))
	}
	x.Duration = time.Duration(
		slice.DecodeUint64(b[*c:c.Inc(slice.Uint64Len)]))
	return
}
