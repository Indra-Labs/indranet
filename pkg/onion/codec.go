package onion

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/ecdh"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/delay"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"github.com/indra-labs/indra/pkg/wire/session"
	"github.com/indra-labs/indra/pkg/wire/token"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func Encode(on types.Onion) (b slice.Bytes) {
	b = make(slice.Bytes, on.Len())
	var sc slice.Cursor
	c := &sc
	on.Encode(b, c)
	return
}

func Peel(b slice.Bytes, c *slice.Cursor) (on types.Onion, e error) {
	switch b[*c:c.Inc(magicbytes.Len)].String() {
	case session.MagicString:
		o := &session.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case confirm.MagicString:
		o := &confirm.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case delay.MagicString:
		o := &delay.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case exit.MagicString:
		o := &exit.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case forward.MagicString:
		o := &forward.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case layer.MagicString:
		var o layer.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case reverse.MagicString:
		o := &reverse.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case response.MagicString:
		o := response.New()
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case token.MagicString:
		o := token.NewOnionSkin()
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	default:
		e = fmt.Errorf("message magic not found")
		log.T.C(func() string {
			return fmt.Sprintln(e) + spew.Sdump(b.ToBytes())
		})
		return
	}
	return
}

func GenCiphers(prvs [3]*prv.Key, pubs [3]*pub.Key) (ciphers [3]sha256.Hash) {
	for i := range prvs {
		ciphers[2-i] = ecdh.Compute(prvs[i], pubs[i])
	}
	return
}

func Gen3Nonces() (n [3]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}