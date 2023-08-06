package forward

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
)

func TestOnions_Forward(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	ipSizes := []int{net.IPv6len, net.IPv4len}
	for i := range ipSizes {
		n := nonce.New()
		ip := net.IP(n[:ipSizes[i]])
		var adr netip.Addr
		if ipSizes[i] == net.IPv4len {
			ip = ip.To4()
		}
		if ipSizes[i] == net.IPv6len {
			ip = ip.To16()
		}
		var ok bool
		if adr, ok = netip.AddrFromSlice(ip); !ok {
			t.Error("unable to get netip.Addrs")
			t.FailNow()
		}
		port := uint16(rand.Uint32())
		ap := netip.AddrPortFrom(adr, port)
		var ma multiaddr.Multiaddr
		var e error
		if ma, e = multi.AddrFromAddrPort(ap); fails(e) {
			t.FailNow()
		}
		log.D.S("ma", ma)
		on := ont.Assemble([]ont.Onion{New(ma)})
		s := codec.Encode(on)
		log.D.S("forward", s.GetAll().ToBytes())
		s.SetCursor(0)
		var onr codec.Codec
		if onr = reg.Recognise(s); onr == nil {
			t.Error("did not unwrap")
			t.FailNow()
		}
		if e := onr.Decode(s); fails(e) {
			t.Error("did not decode")
			t.FailNow()

		}
		var cf *Forward
		if cf, ok = onr.(*Forward); !ok {
			t.Error("did not unwrap expected type expected *Forward got",
				reflect.TypeOf(onr))
			t.FailNow()
		}
		if cf.Multiaddr.String() != ma.String() {
			t.Error("reverse Addresses did not unwrap correctly")
			t.FailNow()
		}
	}
}
