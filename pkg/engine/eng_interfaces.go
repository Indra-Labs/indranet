package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	coding.Codec
	Wrap(inner Onion)
	Handle(s *splice.Splice, p Onion, ni interface{}) (e error)
	Account(res *Data, sm *SessionManager, s *SessionData,
		last bool) (skip bool, sd *SessionData)
}
