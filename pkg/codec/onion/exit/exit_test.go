package exit

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"math/rand"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/tests"
)

func TestOnions_Exit(t *testing.T) {
	var e error
	prvs, pubs := crypto.GetCipherSet()
	ciphers := crypto.GenCiphers(prvs, pubs)
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = tests.GenMessage(512, "aoeu"); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	n3 := crypto.Gen3Nonces()
	p := uint16(rand.Uint32())
	id := nonce.NewID()
	ep := &ExitPoint{
		Routing: &Routing{
			Sessions: [3]*sessions.Data{},
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	on := ont.Assemble([]ont.Onion{New(id, p, msg, ep)})
	s := ont.Encode(on)
	s.SetCursor(0)
	var onc codec.Codec
	if onc = reg.Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ex *Exit
	var ok bool
	if ex, ok = onc.(*Exit); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ex.ID != id {
		t.Error("Keys did not decode correctly")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	plH := sha256.Single(ex.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
		t.FailNow()
	}
}