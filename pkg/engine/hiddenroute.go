package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	HiddenRouteMagic = "hr"
	HiddenRouteLen   = magic.Len + pub.KeyLen + cloak.Len
)

func HiddenRoutePrototype() Onion { return &HiddenRoute{} }

func init() { Register(HiddenRouteMagic, HiddenRoutePrototype) }

type HiddenRoute struct {
	*pub.Key
	recv  *pub.Key
	Cloak cloak.PubKey
	Onion
}

func (o Skins) HiddenRoute(addr, recv *pub.Key) Skins {
	return append(o, &HiddenRoute{
		Key:   addr,
		recv:  recv,
		Cloak: cloak.GetCloak(recv),
	})
}

func (x *HiddenRoute) Magic() string { return HiddenRouteMagic }

func (x *HiddenRoute) Encode(s *octet.Splice) (e error) {
	s.Magic(HiddenRouteMagic).Pubkey(x.Key).Cloak(x.recv)
	if x.Onion != nil {
		return x.Onion.Encode(s)
	}
	return
}

func (x *HiddenRoute) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), HiddenRouteLen-magic.Len,
		HiddenRouteMagic); check(e) {
		
		s.ReadPubkey(&x.Key).
			ReadCloak(&x.Cloak)
		return
	}
	return
}

func (x *HiddenRoute) Len() int { return HiddenRouteLen }

func (x *HiddenRoute) Wrap(inner Onion) { x.Onion = inner }

func (x *HiddenRoute) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln("handling HiddenRoute")
	return
}

func MakeHiddenRoute(addr, recv *pub.Key, header slice.Bytes,
	r *Routing) (o Skins) {
	
	o = o.Triple(header).HiddenRoute(addr, recv).RoutingHeader(r)
	return
}