package indra

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) session(on *session.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	log.T.C(func() string {
		return fmt.Sprint("incoming session",
			spew.Sdump(on.PreimageHash()))
	})
	pi := en.FindPendingPreimage(on.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such
		// messages arrive at the same time, and we end up with
		// duplicate sessions.
		en.DeletePendingPayment(pi.Preimage)
		log.T.F("Adding session %x\n", pi.ID)
		en.AddSession(traffic.NewSession(pi.ID,
			en.Node.Peer, pi.Amount, on.Header, on.Payload, on.Hop))
		en.handleMessage(BudgeUp(b, *c), on)
	} else {
		log.T.Ln("dropping session message without payment")
	}
}
