package address

import (
	"fmt"
	"time"

	"github.com/Indra-Labs/indra/pkg/key/pub"
)

type SendEntry struct {
	*Sender
	time.Time
}

type SendEntries []*SendEntry

func (se SendEntries) Delete(index int) (so SendEntries) {
	if len(se)-1 > index {
		return append(se[:index], se[index+1:]...)
	}
	return se
}

type Index []pub.Bytes

func (ie Index) Delete(index int) (io Index) {
	if len(ie)-1 > index {
		return append(ie[:index], ie[index+1:]...)
	}
	return ie
}

// SendCache is a cache of public keys received from a correspondent node that
// will be used as addressees for the cipher half concealed by a cloaked address
// in a message.
//
// Index stores the key bytes in the same sequence as the SendEntries so it can
// be scanned for matches and then its index used to access the related pub.Key.
type SendCache struct {
	SendEntries
	Index
}

func NewSendCache() *SendCache { return &SendCache{} }

func (sc *SendCache) Add(pb pub.Bytes) (e error) {
	var s *Sender
	if s, e = FromBytes(pb); check(e) {
		return
	}
	se := &SendEntry{Sender: s, Time: time.Now()}
	sc.SendEntries = append(sc.SendEntries, se)
	sc.Index = append(sc.Index, pb)
	return
}

func (sc *SendCache) Find(k pub.Bytes) (se *SendEntry) {
out:
	for i := range sc.Index {
		if k.Equals(sc.Index[i]) {
			se = sc.SendEntries[i]
			break out
		}
	}
	return
}

func (sc *SendCache) Delete(k pub.Bytes) (e error) {
	for i := range sc.Index {
		if k.Equals(sc.Index[i]) {
			sc.SendEntries = sc.SendEntries.Delete(i)
			sc.Index = sc.Index.Delete(i)
			return
		}
	}
	e = fmt.Errorf("key %x not found for deletion", k)
	return
}

type ReceiveEntry struct {
	*Receiver
	time.Time
}

type ReceiveEntries []*ReceiveEntry

func (re ReceiveEntries) Delete(index int) (ro ReceiveEntries) {
	if len(re)-1 > index {
		return append(re[:index], re[index+1:]...)
	}
	return re
}

// ReceiveCache is a cache of the Receiver entries created for adding return
// addresses in messages.
type ReceiveCache struct {
	ReceiveEntries
	Index
}

func NewReceiveCache() *ReceiveCache { return &ReceiveCache{} }

func (rc *ReceiveCache) Add(r *Receiver) (e error) {
	re := &ReceiveEntry{Receiver: r, Time: time.Now()}
	rc.ReceiveEntries = append(rc.ReceiveEntries, re)
	rc.Index = append(rc.Index, pub.Derive(r.Key).ToBytes())
	return
}

func (rc *ReceiveCache) Find(k pub.Bytes) (se *ReceiveEntry) {
out:
	for i := range rc.Index {
		if k.Equals(rc.Index[i]) {
			se = rc.ReceiveEntries[i]
			break out
		}
	}
	return
}

func (rc *ReceiveCache) Delete(k pub.Bytes) (e error) {
	for i := range rc.Index {
		if k.Equals(rc.Index[i]) {
			rc.ReceiveEntries = rc.ReceiveEntries.Delete(i)
			rc.Index = rc.Index.Delete(i)
			return
		}
	}
	e = fmt.Errorf("key %x not found for deletion", k)
	return
}
