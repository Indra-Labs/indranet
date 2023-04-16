package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

const (
	ConfirmationMagic = "cn"
	ConfirmationLen   = magic.Len + nonce.IDLen + 1
)

type Confirmation struct {
	ID   nonce.ID
	Load byte
}

func confirmationGen() Codec             { return &Confirmation{} }
func init()                              { Register(ConfirmationMagic, confirmationGen) }
func (x *Confirmation) Len() int         { return ConfirmationLen }
func (x *Confirmation) Wrap(inner Onion) {}
func (x *Confirmation) GetOnion() Onion  { return x }

func (x *Confirmation) Magic() string { return ConfirmationMagic }

func (x *Confirmation) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Load,
	)
	s.Magic(ConfirmationMagic).ID(x.ID).Byte(x.Load)
	return
}

func (x *Confirmation) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ConfirmationLen-magic.Len,
		ConfirmationMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).ReadByte(&x.Load)
	return
}

func (x *Confirmation) Handle(s *splice.Splice, p Onion,
	ni interface{}) (e error) {
	
	ng := ni.(*Engine)
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.PendingResponses.ProcessAndDelete(x.ID, nil, s.GetAll())
	return
}

func (x *Confirmation) Account(res *Data, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = x.ID
	return
}
