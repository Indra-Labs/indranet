package engine

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/codec/ad/addresses"
	"github.com/indra-labs/indra/pkg/codec/ad/intro"
	"github.com/indra-labs/indra/pkg/codec/ad/load"
	"github.com/indra-labs/indra/pkg/codec/ad/peer"
	"github.com/indra-labs/indra/pkg/codec/ad/services"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"testing"
	"time"

	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func pauza() { time.Sleep(time.Second / 2) }

func TestEngine_PeerStore(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	const nTotal = 30
	var e error
	var engines []*Engine
	var cleanup func()
	engines, cleanup, e = CreateAndStartMockEngines(nTotal)
	adz := engines[0].Listener.Host.Addrs()
	addrs := make([]*netip.AddrPort, len(adz))
	for i := range adz {
		addy, _ := multi.AddrToAddrPort(adz[i])
		addrs[i] = &addy
	}
	pauza()
	newAddressAd := addresses.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		addrs,
		time.Now().Add(time.Hour*24*7))
	sa := splice.New(newAddressAd.Len())
	if e = newAddressAd.Encode(sa); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sa.GetAll()); fails(e) {
		t.FailNow()
	}
	pauza()
	newIntroAd := intro.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		engines[0].Mgr().GetLocalNode().Identity.Pub,
		20000, 443,
		time.Now().Add(time.Hour*24*7))
	si := splice.New(newIntroAd.Len())
	if e = newIntroAd.Encode(si); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(si.GetAll()); fails(e) {
		t.FailNow()
	}
	pauza()
	newLoadAd := load.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		17,
		time.Now().Add(time.Hour*24*7))
	sl := splice.New(newLoadAd.Len())
	if e = newLoadAd.Encode(sl); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sl.GetAll()); fails(e) {
		t.FailNow()
	}
	pauza()
	newPeerAd := peer.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		20000,
		time.Now().Add(time.Hour*24*7))
	log.D.S("peer ad", newPeerAd)
	sp := splice.New(newPeerAd.Len())
	if e = newPeerAd.Encode(sp); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sp.GetAll()); fails(e) {
		t.FailNow()
	}
	pauza()
	newServiceAd := services.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		[]services.Service{{20000, 54321}, {10000, 42221}},
		time.Now().Add(time.Hour*24*7))
	ss := splice.New(newServiceAd.Len())
	if e = newServiceAd.Encode(ss); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(ss.GetAll()); fails(e) {
		t.FailNow()
	}
	pauza()
	cleanup()
	pauza()
}
