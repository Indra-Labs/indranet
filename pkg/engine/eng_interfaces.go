package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

// Codec is a unit of data that can be read and written from a binary form. All
// Onion are Codec but not all Codec are Onion. Codec is also used for the
// Dispatcher's message headers.
type Codec interface {
	Magic() string
	Encode(s *splice.Splice) (e error)
	Decode(s *splice.Splice) (e error)
	Len() int
	GetOnion() Onion
}

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	Codec
	Wrap(inner Onion)
	Handle(s *splice.Splice, p Onion, ni interface{}) (e error)
	Account(res *Data, sm *SessionManager, s *SessionData,
		last bool) (skip bool, sd *SessionData)
}
