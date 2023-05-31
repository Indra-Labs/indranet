package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
)

func TestOnionSkins_PeerAd(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	peerAd := NewPeerAd(id, pr, 20000)
	s := splice.New(peerAd.Len())
	if e = peerAd.Encode(s); fails(e) {
		t.FailNow()
	}
	s.SetCursor(0)
	var onc coding.Codec
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	log.D.S(onc)
	var pa *PeerAd
	var ok bool
	if pa, ok = onc.(*PeerAd); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if pa.RelayRate != peerAd.RelayRate {
		t.Errorf("relay rate did not decode correctly")
		t.FailNow()
	}
	if !pa.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}
