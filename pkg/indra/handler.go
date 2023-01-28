package indra

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"

	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/balance"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/delay"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/getbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) handler() (out bool) {
	log.T.C(func() string {
		return en.AddrPort.String() +
			" awaiting message"
	})
	var prev types.Onion
	select {
	case <-en.C.Wait():
		en.Cleanup()
		out = true
		break
	case b := <-en.Node.Receive():
		en.handleMessage(b, prev)
	case p := <-en.PaymentChan:
		log.T.S("incoming payment", en.AddrPort.String(), p)
		topUp := false
		en.IterateSessions(func(s *traffic.Session) bool {
			if s.Preimage == p.Preimage {
				s.IncSats(p.Amount)
				topUp = true
				log.T.F("topping up %x with %d mSat",
					s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			en.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %x",
				p.Preimage)
		}
	}
	return
}

func (en *Engine) handleMessage(b slice.Bytes, prev types.Onion) {
	// process received message
	var on types.Onion
	var e error
	c := slice.NewCursor()
	if on, e = onion.Peel(b, c); check(e) {
		return
	}
	switch on := on.(type) {
	case *balance.Layer:
		log.T.C(recLog(on, b, en))
		en.balance(on, b, c, prev)
	case *confirm.Layer:
		log.T.C(recLog(on, b, en))
		en.confirm(on, b, c, prev)
	case *crypt.Layer:
		log.T.C(recLog(on, b, en))
		en.crypt(on, b, c, prev)
	case *delay.Layer:
		log.T.C(recLog(on, b, en))
		en.delay(on, b, c, prev)
	case *exit.Layer:
		log.T.C(recLog(on, b, en))
		en.exit(on, b, c, prev)
	case *forward.Layer:
		log.T.C(recLog(on, b, en))
		en.forward(on, b, c, prev)
	case *getbalance.Layer:
		log.T.C(recLog(on, b, en))
		en.getBalance(on, b, c, prev)
	case *reverse.Layer:
		log.T.C(recLog(on, b, en))
		en.reverse(on, b, c, prev)
	case *response.Layer:
		log.T.C(recLog(on, b, en))
		en.response(on, b, c, prev)
	case *session.Layer:
		log.T.C(recLog(on, b, en))
		en.session(on, b, c, prev)
	default:
		log.I.S("unrecognised packet", b)
	}
}

// utility functions

func recLog(on types.Onion, b slice.Bytes, cl *Engine) func() string {
	return func() string {
		return cl.AddrPort.String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			spew.Sdump(b.ToBytes())
	}
}