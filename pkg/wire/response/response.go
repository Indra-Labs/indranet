package response

import (
	"reflect"
	"unsafe"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "rs"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + slice.Uint32Len
	_           types.Onion = OnionSkin{}
)

// OnionSkin messages are what are carried back via Reply messages from an Exit.
type OnionSkin slice.Bytes

func (x OnionSkin) Inner() types.Onion   { return nil }
func (x OnionSkin) Insert(_ types.Onion) {}
func (x OnionSkin) Len() int             { return MinLen + len(x) }

func (x OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x)-slice.Uint32Len)
	copy(b[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(b[*c:c.Inc(len(x))], x)
}

func (x OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b[*c:]), MinLen, string(Magic))
	}
	responseLen := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	xd := OnionSkin(b[*c:c.Inc(responseLen)])
	// replace current slice header using unsafe.
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&x))
	hdr.Data = (*reflect.SliceHeader)(unsafe.Pointer(&xd)).Data
	return
}
