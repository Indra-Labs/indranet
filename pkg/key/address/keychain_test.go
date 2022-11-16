package address

import (
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
)

func TestReceiveCache_Add(t *testing.T) {
	rc := NewReceiveCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		if e = rc.Add(NewReceiver(prvKey)); check(e) {
			t.Error(e)
		}
	}
	if rc.Len() != makeCount {
		t.Error("did not find expected number of entries in cache")
	}
}

func TestReceiveCache_Delete(t *testing.T) {
	rc := NewReceiveCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		if e = rc.Add(NewReceiver(prvKey)); check(e) {
			t.Error(e)
		}
	}
	for _ = range rc.Index {
		if e := rc.Delete(rc.ReceiveEntries[0].Bytes); check(e) {
			t.Error(e)
		}
	}
	if rc.Len() != 0 {
		t.Error("did not find expected number of entries in cache")
	}
}

func TestReceiveCache_Find(t *testing.T) {
	rc := NewReceiveCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		if e = rc.Add(NewReceiver(prvKey)); check(e) {
			t.Error(e)
		}
	}
	for i := range rc.Index {
		if rc.Find(rc.Index[i]) == nil {
			t.Error("failed to find an entry")
		}
	}
}

func TestSendCache_Add(t *testing.T) {
	sc := NewSendCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		pubKey := pub.Derive(prvKey)
		if e = sc.Add(pubKey.ToBytes()); check(e) {
			t.Error(e)
		}
	}
	if sc.Len() != makeCount {
		t.Error("did not find expected number of entries in cache")
	}
}

func TestSendCache_Delete(t *testing.T) {
	sc := NewSendCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		pubKey := pub.Derive(prvKey)
		if e = sc.Add(pubKey.ToBytes()); check(e) {
			t.Error(e)
		}
	}
	for _ = range sc.Index {
		if e := sc.Delete(sc.SendEntries[0].Sender.ToBytes()); check(e) {
			t.Error(e)
		}
	}
	if sc.Len() != 0 {
		t.Error("did not find expected number of entries in cache")
	}
}

func TestSendCache_Find(t *testing.T) {
	sc := NewSendCache()
	const makeCount = 1000
	for i := 0; i < makeCount; i++ {
		prvKey, e := prv.GenerateKey()
		if check(e) {
			t.Error(e)
		}
		pubKey := pub.Derive(prvKey)
		if e = sc.Add(pubKey.ToBytes()); check(e) {
			t.Error(e)
		}
	}
	for i := range sc.Index {
		if sc.Find(sc.Index[i]) == nil {
			t.Error("failed to find an entry")
		}
	}
}
