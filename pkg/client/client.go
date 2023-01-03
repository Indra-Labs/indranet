package client

import (
	"net/netip"
	"sync"
	"time"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	session2 "github.com/Indra-Labs/indra/pkg/session"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirm"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/layer"
	"github.com/Indra-Labs/indra/pkg/wire/noop"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/reverse"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
)

const (
	DefaultDeadline  = 10 * time.Minute
	ReverseLayerLen  = reverse.Len + layer.Len
	ReverseHeaderLen = 3 * ReverseLayerLen
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	*netip.AddrPort
	*node.Node
	node.Nodes
	*address.SendCache
	*address.ReceiveCache
	Circuits
	session2.Sessions
	PendingSessions []nonce.ID
	*confirm.Confirms
	sync.Mutex
	*signer.KeySet
	atomic.Bool
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.HeaderPrv = hdrPrv
	no.HeaderPub = pub.Derive(hdrPrv)
	c = &Client{
		Confirms:     confirm.NewConfirms(),
		Node:         no,
		Nodes:        nodes,
		ReceiveCache: address.NewReceiveCache(),
		C:            qu.T(),
	}
	c.ReceiveCache.Add(address.NewReceiver(hdrPrv))
	return
}

// Start a single thread of the Client.
func (cl *Client) Start() {
out:
	for {
		if cl.runner() {
			break out
		}
	}
}

func (cl *Client) runner() (out bool) {
	select {
	case <-cl.C.Wait():
		cl.Cleanup()
		out = true
		break
	case b := <-cl.Node.Receive():
		// process received message
		var onion types.Onion
		var e error
		c := slice.NewCursor()
		if onion, e = wire.PeelOnion(b, c); check(e) {
			break
		}
		switch on := onion.(type) {
		case *cipher.OnionSkin:
			cl.cipher(on, b, c)
		case *confirm.OnionSkin:
			cl.confirm(on, b, c)
		case *delay.OnionSkin:
			cl.delay(on, b, c)
		case *exit.OnionSkin:
			cl.exit(on, b, c)
		case *forward.OnionSkin:
			cl.forward(on, b, c)
		case *layer.OnionSkin:
			cl.layer(on, b, c)
		case *noop.OnionSkin:
			cl.noop(on, b, c)
		case *purchase.OnionSkin:
			cl.purchase(on, b, c)
		case *reverse.OnionSkin:
			cl.reverse(on, b, c)
		case *response.OnionSkin:
			cl.response(on, b, c)
		case *session.OnionSkin:
			cl.session(on, b, c)
		case *token.OnionSkin:
			cl.token(on, b, c)
		default:
			log.I.S("unrecognised packet", b)
		}
	}
	return
}

func (cl *Client) cipher(on *cipher.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// This either is in a forward only SendKeys message or we are the buyer
	// and these are our session keys.
	log.I.S(on)
	b = append(b[*c:], slice.NoisePad(int(*c))...)
	cl.Node.Send(b)
}

func (cl *Client) confirm(on *confirm.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {
	// When a confirm arrives check if it is registered for and run
	// the hook that was registered with it.
	cl.Confirms.Confirm(on.ID)
}

func (cl *Client) RegisterConfirmation(hook confirm.Hook, on *confirm.OnionSkin) {

	cl.Confirms.Add(&confirm.Callback{
		ID:    on.ID,
		Time:  time.Now(),
		Hook:  hook,
		Onion: on,
	})
}

func (cl *Client) delay(on *delay.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
}

func (cl *Client) exit(on *exit.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// payload is forwarded to a local port and the response is forwarded
	// back with a reverse header.
}

func (cl *Client) forward(on *forward.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// forward the whole buffer received onwards. Usually there will be a
	// layer.OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next
		// part.
		b = append(b[*c:], slice.NoisePad(int(*c))...)
		cl.Node.Send(b)
	} else {
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}
}

func (cl *Client) layer(on *layer.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this is probably an encrypted layer for us.
	rcv := cl.ReceiveCache.FindCloaked(on.Cloak)
	if rcv == nil {
		log.I.Ln("no matching key found from cloaked key")
		return
	}
	on.Decrypt(rcv.Key, b, c)
	b = append(b[*c:], slice.NoisePad(int(*c))...)
	cl.Node.Send(b)
}

func (cl *Client) noop(on *noop.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this won't happen normally
}

func (cl *Client) purchase(on *purchase.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// Create a new Session.
	s := session2.NewSession(on.ID, on.NBytes, DefaultDeadline, cl.KeySet)
	se := &session.OnionSkin{
		ID:         s.ID,
		HeaderKey:  s.HeaderKey.Key,
		PayloadKey: s.PayloadKey.Key,
		Onion:      &noop.OnionSkin{},
	}
	cl.Mutex.Lock()
	cl.Sessions.Add(s)
	cl.Mutex.Unlock()
	header := b[*c:c.Inc(ReverseHeaderLen)]
	// log.I.S(header.ToBytes())
	rb := make(slice.Bytes, ReverseHeaderLen+session.Len)
	cur := slice.NewCursor()
	copy(rb[*cur:cur.Inc(ReverseHeaderLen)], header[:ReverseHeaderLen])
	start := *cur
	se.Encode(rb, cur)

	for i := range on.Ciphers {
		blk := ciph.BlockFromHash(on.Ciphers[i])
		ciph.Encipher(blk, on.Nonces[2-i], rb[start:])
	}
	cl.Node.Send(rb)
}

func (cl *Client) reverse(on *reverse.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	var e error
	var onion types.Onion
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		if onion, e = wire.PeelOnion(b, c); check(e) {
			return
		}
		switch on1 := onion.(type) {
		case *layer.OnionSkin:
			// log.I.F("in reverse %x %s", on1.Nonce[:], on.AddrPort.String())
			// log.I.S(on1)
			start := *c - ReverseLayerLen
			first := *c
			second := first + ReverseLayerLen
			last := second + ReverseLayerLen
			rcv := cl.ReceiveCache.FindCloaked(on1.Cloak)
			// We need to find the PayloadKey to match.
			ses := cl.Sessions.FindPub(rcv.Pub)
			hdrPrv := ses.HeaderPrv
			hdrPub := on1.FromPub
			blk := ciph.GetBlock(hdrPrv, hdrPub)
			// Decrypt using the Payload key and header nonce.
			ciph.Encipher(blk, on1.Nonce,
				b[*c:c.Inc(2*ReverseLayerLen)])
			blk = ciph.GetBlock(ses.PayloadPrv, hdrPub)
			ciph.Encipher(blk, on1.Nonce, b[*c:])
			// shift the header segment upwards and pad the
			// remainder.
			copy(b[start:first], b[first:second])
			copy(b[first:second], b[second:last])
			copy(b[second:last], slice.NoisePad(ReverseLayerLen))
			if b[start:start+2].String() != reverse.MagicString {
				cl.Node.Send(b[last:])
				break
			}
			cl.Node.Send(b[start:])
		default:
			return
		}
	} else {
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}

}

func (cl *Client) response(on *response.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Response is a payload from an exit message.
}

func (cl *Client) session(s *session.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Session is returned from a Purchase message in the reverse layers.
	//
	// Session has a nonce.ID that is given in the last layer of a LN sphinx
	// Bolt 4 onion routed payment path that will cause the seller to
	// activate the accounting onion the two keys it sent back with the
	// nonce, so long as it has not yet expired.
	for i := range cl.PendingSessions {
		if cl.PendingSessions[i] == s.ID {
			// we would make payment and move session to running
			// sessions list.
			log.I.S("session received, now to pay", s)
		}
	}
	// So now we want to pay.
}

func (cl *Client) token(t *token.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// not really sure if we are using these.
	return
}

func (cl *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (cl *Client) Shutdown() {
	if cl.Bool.Load() {
		return
	}
	log.I.Ln("shutting down client", cl.Node.AddrPort.String())
	cl.Bool.Store(true)
	cl.C.Q()
}

func (cl *Client) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range cl.Nodes {
		if as == cl.Nodes[i].Addr {
			cl.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them.

}
