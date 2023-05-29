package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

const (
	DelayMagic = "dely"
	DelayLen   = magic.Len + slice.Uint64Len
)

type Delay struct {
	time.Duration
	Onion
}

func (x *Delay) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *Delay) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), DelayLen-magic.Len, DelayMagic); fails(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

func (x *Delay) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Duration,
	)
	s.Magic(DelayMagic).Uint64(uint64(x.Duration))
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

func (x *Delay) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(x.Duration):
	}
	ng.HandleMessage(splice.BudgeUp(s), x)
	return
}

func (x *Delay) Len() int         { return DelayLen + x.Onion.Len() }
func (x *Delay) Magic() string    { return DelayMagic }
func (x *Delay) Wrap(inner Onion) { x.Onion = inner }
func delayGen() coding.Codec      { return &Delay{} }
func init()                       { Register(DelayMagic, delayGen) }
