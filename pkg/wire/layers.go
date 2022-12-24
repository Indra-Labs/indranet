package wire

import (
	"net/netip"
	"time"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/ecdh"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/message"
	"github.com/Indra-Labs/indra/pkg/wire/noop"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/reply"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
)

func GenCiphers(prvs [3]*prv.Key, pubs [3]*pub.Key) (ciphers [3]sha256.Hash) {
	for i := range prvs {
		ciphers[i] = ecdh.Compute(prvs[i], pubs[i])
	}
	return
}

type OnionSkins []types.Onion

func (o OnionSkins) Cipher(hdr, pld *pub.Key) OnionSkins {
	return append(o, &cipher.OnionSkin{
		Header:  hdr,
		Payload: pld,
		Onion:   &noop.OnionSkin{},
	})
}

func (o OnionSkins) Confirmation(id nonce.ID) OnionSkins {
	return append(o, &confirmation.OnionSkin{ID: id})
}

func (o OnionSkins) Delay(d time.Duration) OnionSkins {
	return append(o, &delay.OnionSkin{Duration: d,
		Onion: &noop.OnionSkin{}})
}

func (o OnionSkins) Exit(port uint16, prvs [3]*prv.Key, pubs [3]*pub.Key,
	payload slice.Bytes) OnionSkins {

	return append(o, &exit.OnionSkin{
		Port:    port,
		Ciphers: GenCiphers(prvs, pubs),
		Bytes:   payload,
		Onion:   &noop.OnionSkin{},
	})
}
func (o OnionSkins) Forward(addr *netip.AddrPort) OnionSkins {
	return append(o, &forward.OnionSkin{AddrPort: addr, Onion: &noop.OnionSkin{}})
}
func (o OnionSkins) Message(to *address.Sender, from *prv.Key) OnionSkins {
	return append(o, &message.OnionSkin{
		To:    to,
		From:  from,
		Onion: &noop.OnionSkin{},
	})
}
func (o OnionSkins) Purchase(nBytes uint64, ciphers [3]sha256.Hash) OnionSkins {
	return append(o, &purchase.OnionSkin{
		NBytes:  nBytes,
		Ciphers: ciphers,
		Onion:   &noop.OnionSkin{},
	})
}
func (o OnionSkins) Reply(ip *netip.AddrPort) OnionSkins {
	return append(o, &reply.OnionSkin{AddrPort: ip, Onion: &noop.OnionSkin{}})
}
func (o OnionSkins) Response(res slice.Bytes) OnionSkins {
	rs := response.OnionSkin(res)
	return append(o, &rs)
}
func (o OnionSkins) Session(hdr, pld *pub.Key) OnionSkins {
	return append(o, &session.OnionSkin{
		HeaderKey:  hdr,
		PayloadKey: pld,
		Onion:      &noop.OnionSkin{},
	})
}
func (o OnionSkins) Token(tok sha256.Hash) OnionSkins {
	return append(o, (*token.OnionSkin)(&tok))
}

// Assemble inserts the slice of OnionSkin s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o OnionSkins) Assemble() (on types.Onion) {
	// First item is the outer layer.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Insert(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}
