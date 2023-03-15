package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// Onion is an interface for the layers of messages each encrypted inside a
// OnionSkin, which provides the cipher for the inner layers inside it.
type Onion interface {
	Magic() string
	Encode(s *octet.Splice) (e error)
	Decode(s *octet.Splice) (e error)
	Len() int
	Wrap(inner Onion)
	Handle(s *octet.Splice, p Onion, ng *Engine) (e error)
}

type Transport interface {
	Chain(t Transport) Transport
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}