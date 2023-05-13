package advert

import (
	"fmt"
	"time"
	
	"github.com/multiformats/go-multiaddr"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	PeerMagic = "peer"
	PeerLen   = magic.Len + nonce.IDLen + slice.Uint64Len + crypto.SigLen
)

// Peer is the root identity document for an Indra peer. It is indexed by the
// Identity field, its public key. The slices found below it are derived via
// concatenation of strings with the keys and hashing to generate a derived
// field index, used to search the DHT for matches.
//
// The data stored for Peer must be signed with the key claimed by the Identity.
// For hidden services the address fields are signed in the DHT by the hidden
// service from their introduction solicitation, and the index from the current
// set is given by the hidden service.
type Peer struct {
	nonce.ID  // To ensure no repeating message
	Identity  crypto.PubBytes
	RelayRate int
	Sig       crypto.SigBytes
	// Addresses - first is address, nil for hidden services,
	// hidden services have more than one, 6 or more are kept active.
	Addresses []*Address
	Services  []Service
}

func (p *Peer) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(p.Len())
	if e = p.Encode(s); fails(e) {
		return
	}
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	if len(b) != crypto.SigLen {
		return fmt.Errorf("signature incorrect length, got %d expected %d",
			len(b), crypto.SigLen)
	}
	copy(p.Sig[:], b)
	return nil
}

func (p *Peer) Magic() string {
	return ""
}

func (p *Peer) Encode(s *splice.Splice) (e error) {
	s.ID(p.ID).Uint64(uint64(p.RelayRate))
	return nil
}

func (p *Peer) Decode(s *splice.Splice) (e error) {
	var v uint64
	s.ReadID(&p.ID).ReadUint64(&v)
	s.ReadSignature(&p.Sig)
	p.RelayRate = int(v)
	return nil
}

func (p *Peer) Validate(s *splice.Splice, pub *crypto.Pub) bool {
	h := sha256.Single(s.GetRange(0, nonce.IDLen+slice.Uint64Len))
	var e error
	var pk *crypto.Pub
	if pk, e = p.Sig.Recover(h); fails(e) {
		return false
	}
	if pub.Equals(pk) {
		return true
	}
	return false
}

func (p *Peer) Len() int {
	return PeerLen
}

func (p *Peer) GetOnion() interface{} {
	return nil
}

// Address entries are stored with an index generated by concatenating the bytes
// of the public key with a string path "/address/N" where N is the index of the
// address. This means hidden service introducers for values over zero.
// Hidden services have no value in the zero index, which is "<hash>/address/0".
type Address struct {
	multiaddr.Multiaddr
	ID     nonce.ID // To ensure no repeating message
	Index  byte
	Expiry time.Time // zero for relay's public address (32 bit).
	crypto.SigBytes
}

const AddressLen = nonce.IDLen + 1 + slice.Uint32Len + crypto.SigLen

func (a *Address) Magic() string { return "" }

func (a *Address) Encode(s *splice.Splice) (e error) {
	// TODO implement me
	panic("implement me")
}

func (a *Address) Decode(s *splice.Splice) (e error) {
	// TODO implement me
	panic("implement me")
}

func (a *Address) Len() int              { return AddressLen }
func (a *Address) GetOnion() interface{} { return nil }

// Service stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the Peer entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Service struct {
	ID        nonce.ID // To ensure no repeating message
	Index     uint16
	Port      uint16
	RelayRate int
	crypto.SigBytes
}

const ServiceLen = nonce.IDLen + 2*slice.Uint16Len + slice.Uint64Len +
	crypto.SigLen

func (sv *Service) Magic() string { return "" }

func (sv *Service) Encode(s *splice.Splice) (e error) {
	// TODO implement me
	panic("implement me")
}

func (sv *Service) Decode(s *splice.Splice) (e error) {
	// TODO implement me
	panic("implement me")
}

func (sv *Service) Len() int              { return ServiceLen }
func (sv *Service) GetOnion() interface{} { return nil }
