package engine

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onions/ad/addresses"
	"github.com/indra-labs/indra/pkg/onions/ad/intro"
	"github.com/indra-labs/indra/pkg/onions/ad/load"
	"github.com/indra-labs/indra/pkg/onions/ad/peer"
	"github.com/indra-labs/indra/pkg/onions/ad/services"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestEngine_PeerStore(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	const nTotal = 26
	var cancel func()
	var e error
	var engines []*Engine
	var seed string
	for i := 0; i < nTotal; i++ {
		dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
		if err != nil {
			t.FailNow()
		}
		var eng *Engine
		if eng, cancel, e = CreateMockEngine(seed, dataPath); fails(e) {
			return
		}
		engines = append(engines, eng)
		if i == 0 {
			seed = transport.GetHostAddress(eng.Listener.Host)
		}
		defer os.RemoveAll(dataPath)
		go eng.Start()
	}
	time.Sleep(time.Second)
	adz := engines[0].Listener.Host.Addrs()
	addrs := make([]*netip.AddrPort, len(adz))
	for i := range adz {
		addy, _ := multi.AddrToAddrPort(adz[i])
		addrs[i] = &addy
	}
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
	time.Sleep(time.Second)
	newIntroAd := intro.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		engines[0].Mgr().GetLocalNodeAddress(),
		20000, 443,
		time.Now().Add(time.Hour*24*7))
	si := splice.New(newIntroAd.Len())
	if e = newIntroAd.Encode(si); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(si.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
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
	time.Sleep(time.Second)
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
	time.Sleep(time.Second * 1)
	newServiceAd := services.New(nonce.NewID(),
		engines[0].Mgr().GetLocalNodeIdentityPrv(),
		[]services.Service{{20000, 54321}},
		time.Now().Add(time.Hour*24*7))
	ss := splice.New(newServiceAd.Len())
	if e = newServiceAd.Encode(ss); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(ss.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	cancel()
	for i := range engines {
		engines[i].Shutdown()
	}
}
