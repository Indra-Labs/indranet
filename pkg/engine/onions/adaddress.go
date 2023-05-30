package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"time"
)

const (
	AddressAdMagic = "adad"
	AddressAdLen   = nonce.IDLen +
		splice.AddrLen + 1 +
		slice.Uint64Len +
		crypto.SigLen
)

// AddressAd entries are stored with an index generated by concatenating the bytes
// of the public key with a string path "/address/N" where N is the index of the
// address. This means hidden service introducers for values over zero.
// Hidden services have no value in the zero index, which is "<hash>/address/0".
type AddressAd struct {
	ID        nonce.ID            // To ensure no repeating message
	Multiaddr multiaddr.Multiaddr // We only use a netip.AddrPort though.
	Index     byte                // This is the index in the slice from Peer.
	Expiry    time.Time           // zero for relay's public address (32 bit).
	Sig       crypto.SigBytes
}

func (x *AddressAd) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {

	return false, nil
}

func (x *AddressAd) Decode(s *splice.Splice) (e error) {
	var addr *netip.AddrPort
	s.ReadID(&x.ID).ReadAddrPort(&addr).ReadByte(&x.Index).ReadTime(&x.Expiry)
	return
}

func (x *AddressAd) Encode(s *splice.Splice) (e error) {
	x.Splice(s.Magic(AddressAdMagic))
	return
}

func (x *AddressAd) GetOnion() interface{}                               { return nil }
func (x *AddressAd) Gossip(sm *sess.Manager, c qu.C)                     {}
func (x *AddressAd) Handle(s *splice.Splice, p Onion, ni Ngin) (e error) { return nil }
func (x *AddressAd) Len() int                                            { return AddressAdLen }
func (x *AddressAd) Magic() string                                       { return "" }

func (x *AddressAd) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	if e = x.Encode(s); fails(e) {
		return
	}
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	if len(b) != crypto.SigLen {
		return fmt.Errorf("signature incorrect length, got %d expected %d",
			len(b), crypto.SigLen)
	}
	copy(x.Sig[:], b)
	return nil
}

func (x *AddressAd) Splice(s *splice.Splice) {
	var e error
	var ip, port string
	if ip, e = x.Multiaddr.ValueForProtocol(multiaddr.P_IP4); fails(e) {
	}
	if ip == "" {
		if ip, e = x.Multiaddr.ValueForProtocol(multiaddr.P_IP6); fails(e) {
			return
		}
	}
	// There is really no alternative to TCP so, TCP it is.
	if port, e = x.Multiaddr.ValueForProtocol(multiaddr.P_TCP); fails(e) {
		return
	}
	var addr netip.AddrPort
	if addr, e = netip.ParseAddrPort(ip + ":" + port); fails(e) {
	}
	s.ID(x.ID).AddrPort(&addr).Byte(x.Index).Time(x.Expiry)
}

func (x *AddressAd) Validate(s *splice.Splice) (pub *crypto.Pub) {
	h := sha256.Single(s.GetRange(0, nonce.IDLen+splice.AddrLen+1+
		slice.Uint64Len))
	var e error
	if pub, e = x.Sig.Recover(h); fails(e) {
	}
	return
}

func (x *AddressAd) Wrap(inner Onion) {}
