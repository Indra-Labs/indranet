package slice

import (
	"bytes"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/tests"
)

func TestMessage_ToU64Slice(t *testing.T) {
	var e error
	var msg1 Bytes
	if msg1, _, e = tests.GenMessage(33, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	uMsg1 := msg1.ToU64Slice()
	umsg1 := uMsg1.ToMessage()
	if bytes.Compare(msg1, umsg1) != 0 {
		t.Error("conversion to U64Slice and back to []byte failed")
		t.FailNow()
	}
}

func TestU64Slice_XOR(t *testing.T) {
	const ml = 1024
	var e error
	var msg1 Bytes
	if msg1, _, e = tests.GenMessage(ml, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	hash1 := sha256.Single(msg1)
	uMsg1 := msg1.ToU64Slice()
	var msg2 Bytes
	if msg2, _, e = tests.GenMessage(ml, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// log.I.S(msg2)
	uMsg2 := msg2.ToU64Slice()
	var msg3 Bytes
	if msg3, _, e = tests.GenMessage(ml, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// log.I.S(msg3)
	uMsg3 := msg3.ToU64Slice()
	uMsg1.XOR(uMsg2)
	uMsg1.XOR(uMsg3)
	uMsg1.XOR(uMsg2)
	uMsg1.XOR(uMsg3)
	hash2 := sha256.Single(uMsg1.ToMessage())
	if hash1 != hash2 {
		t.Error("XOR failed")
		t.FailNow()
	}
}