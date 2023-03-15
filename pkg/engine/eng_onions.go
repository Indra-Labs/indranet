package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

type Skins []Onion

var nop = &Tmpl{}

func Encode(on Onion) (s *octet.Splice) {
	s = octet.New(on.Len())
	check(on.Encode(s))
	return
}

// Assemble inserts the slice of Layer s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o Skins) Assemble() (on Onion) {
	// First item is the outer crypt.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Wrap(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}

func (o Skins) ForwardCrypt(s *SessionData, k *prv.Key, n nonce.IV) Skins {
	return o.Forward(s.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, 0)
}

func (o Skins) ReverseCrypt(s *SessionData, k *prv.Key, n nonce.IV,
	seq int) Skins {
	
	return o.Reverse(s.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, seq)
}

type Routing struct {
	Sessions [3]*SessionData
	Keys     [3]*prv.Key
	Nonces   [3]nonce.IV
}

type Headers struct {
	Forward, Return *Routing
	ReturnPubs      [3]*pub.Key
}

func GetHeaders(Client *SessionData, S Circuit,
	KS *signer.KeySet) (h *Headers) {
	
	fwKeys := KS.Next3()
	rtKeys := KS.Next3()
	n := GenNonces(6)
	var rtNonces, fwNonces [3]nonce.IV
	copy(fwNonces[:], n[:3])
	copy(rtNonces[:], n[3:])
	var fwSessions, rtSessions [3]*SessionData
	copy(fwSessions[:], S[:3])
	copy(rtSessions[:], S[3:5])
	rtSessions[2] = Client
	var returnPubs [3]*pub.Key
	returnPubs[0] = S[3].PayloadPub
	returnPubs[1] = S[4].PayloadPub
	returnPubs[2] = Client.PayloadPub
	h = &Headers{
		Forward: &Routing{
			Sessions: fwSessions,
			Keys:     fwKeys,
			Nonces:   fwNonces,
		},
		Return: &Routing{
			Sessions: rtSessions,
			Keys:     rtKeys,
			Nonces:   rtNonces,
		},
		ReturnPubs: returnPubs,
	}
	return
}

type ExitPoint struct {
	*Routing
	ReturnPubs [3]*pub.Key
}

func (h *Headers) ExitPoint() *ExitPoint {
	return &ExitPoint{
		Routing:    h.Return,
		ReturnPubs: h.ReturnPubs,
	}
}

func (o Skins) RoutingHeader(r *Routing) Skins {
	return o.
		ReverseCrypt(r.Sessions[0], r.Keys[0], r.Nonces[0], 3).
		ReverseCrypt(r.Sessions[1], r.Keys[1], r.Nonces[1], 2).
		ReverseCrypt(r.Sessions[2], r.Keys[2], r.Nonces[2], 1)
}

func (o Skins) ForwardSession(s *Node,
	k *prv.Key, n nonce.IV, sess *Session) Skins {
	
	return o.Forward(s.AddrPort).
		Crypt(s.IdentityPub, nil, k, n, 0).
		Session(sess)
}