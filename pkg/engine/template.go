package engine

import "git-indra.lan/indra-labs/indra/pkg/splice"

const (
	EndMagic = "!!"
	EndLen   = 0
)

func EndGen() Codec             { return &End{} }
func init()                     { Register(EndMagic, EndGen) }
func (x *End) Magic() string    { return EndMagic }
func (x *End) Len() int         { return EndLen }
func (x *End) Wrap(inner Onion) {}
func (x *End) GetOnion() Onion  { return x }

type End struct{}

func NewEnd() *End {
	return &End{}
}

func (x *End) Encode(s *splice.Splice) (e error) {
	return
}

func (x *End) Decode(s *splice.Splice) (e error) {
	return
}

func (x *End) Handle(s *splice.Splice, p Onion,
	ng *Engine) (e error) {
	
	return
}

func (x *End) Account(res *SendData, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	return
}
