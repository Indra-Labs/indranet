package message

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/consts"
	"github.com/indra-labs/indra/pkg/onions/hidden"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	MessageMagic    = "mess"
	ReplyCiphersLen = 2*consts.RoutingHeaderLen +
		6*sha256.Len +
		6*nonce.IVLen
	MessageLen      = magic.Len +
		2*nonce.IDLen +
		2*consts.RoutingHeaderLen +
		ReplyCiphersLen
)

type Message struct {
	Forwards        [2]*sessions.Data
	Address         *crypto.Pub
	ID, Re          nonce.ID
	Forward, Return *hidden.ReplyHeader
	Payload         slice.Bytes
}

func NewMessage() (msg *Message){

	return
}

func (x *Message) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

func (x *Message) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), MessageLen-magic.Len,
		MessageMagic); fails(e) {
		return
	}
	x.Return = &hidden.ReplyHeader{}
	s.ReadPubkey(&x.Address).
		ReadID(&x.ID).ReadID(&x.Re)
	hidden.ReadRoutingHeader(s, &x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces).
		ReadBytes(&x.Payload)
	return
}

func (x *Message) Encode(s *splice.Splice) (e error) {
	log.T.F("encoding %s %x %x %v %s", reflect.TypeOf(x),
		x.ID, x.Re, x.Address, spew.Sdump(x.Forward, x.Return,
			x.Payload.ToBytes()),
	)
	hidden.WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(MessageMagic).
		Pubkey(x.Address).
		ID(x.ID).ID(x.Re)
	hidden.WriteRoutingHeader(s, x.Return.RoutingHeaderBytes).
		Ciphers(x.Return.Ciphers).
		Nonces(x.Return.Nonces).
		Bytes(x.Payload)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		log.D.F("encrypting %s", x.Forward.Ciphers[i])
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
	return
}

func MessageGen() coding.Codec           { return &Message{} }
func (x *Message) GetOnion() interface{} { return x }

func (x *Message) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// Forward payload out to service port.
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (x *Message) Len() int             { return MessageLen + x.Payload.Len() }
func (x *Message) Magic() string        { return MessageMagic }
func (x *Message) Wrap(inner ont.Onion) {}
func init()                             { reg.Register(MessageMagic, MessageGen) }