package indra

import (
	"time"

	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) exit(ex *exit.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var e error
	var result slice.Bytes
	h := sha256.Single(ex.Bytes)
	log.T.S(h)
	if e = en.SendTo(ex.Port, ex.Bytes); check(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5) // todo: timeout/retries etc
	select {
	case result = <-en.ReceiveFrom(ex.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt. The client recognises
	// the context of the response by the hash of the request message.
	en.Lock()
	res := onion.Encode(&response.Layer{
		Hash:  h,
		ID:    en.ID,
		Load:  en.Load,
		Bytes: result,
	})
	en.Unlock()
	rb := FormatReply(b[*c:c.Inc(crypt.ReverseHeaderLen)],
		res, ex.Ciphers, ex.Nonces)
	switch on := prev.(type) {
	case *crypt.Layer:
		sess := en.FindSessionByHeader(on.ToPriv)
		if sess == nil {
			break
		}
		for i := range sess.Services {
			if ex.Port != sess.Services[i].Port {
				continue
			}
			in := sess.Services[i].RelayRate *
				lnwire.MilliSatoshi(len(b)) / 2 / 1024 / 1024
			out := sess.Services[i].RelayRate *
				lnwire.MilliSatoshi(len(rb)) / 2 / 1024 / 1024
			log.D.Ln(sess.AddrPort.String(), "exit send")
			en.DecSession(sess.ID, in+out)
			break
		}
	}
	en.handleMessage(rb, ex)
}