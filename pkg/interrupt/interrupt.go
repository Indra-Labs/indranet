package interrupt

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"go.uber.org/atomic"
	"golang.org/x/sys/unix"

	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"

	"github.com/kardianos/osext"
)

var (
	Restart   bool
	requested atomic.Bool
	// ch is used to receive SIGINT (Ctrl+C) signals.
	ch chan os.Signal
	// signals is the list of signals that cause the interrupt
	signals = []os.Signal{syscall.SIGINT}
	// ShutdownRequestChan is a channel that can receive shutdown requests
	ShutdownRequestChan = make(chan struct{})
	// addHandlerChan is used to add an interrupt handler to the list of
	// handlers to be invoked on SIGINT (Ctrl+C) signals.
	addHandlerChan = make(chan HandlerWithSource)
	// HandlersDone is closed after all interrupt handlers run the first
	// time an interrupt is signaled.
	HandlersDone             = make(chan struct{})
	interruptCallbackSources []string
	interruptCallbacks       []func()
	log                      = log2.GetLogger(indra.PathBase)
	check                    = log.E.Chk
)

type HandlerWithSource struct {
	Source string
	Fn     func()
}

// AddHandler adds a handler to call when a SIGINT (Ctrl+C) is received.
func AddHandler(handler func()) {
	// Create the channel and start the main interrupt handler that invokes
	// all other callbacks and exits if not already done.
	_, loc, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf("%s:%d", loc, line)
	log.T.Ln("\n"+msg, "added interrupt handler")
	if ch == nil {
		ch = make(chan os.Signal)
		signal.Notify(ch, signals...)
		go Listener()
	}
	addHandlerChan <- HandlerWithSource{
		msg, handler,
	}
}

// GoroutineDump returns a string with the current goroutine dump in order to
// show what's going on in case of timeout.
func GoroutineDump() string {
	buf := make([]byte, 1<<18)
	n := runtime.Stack(buf, true)
	return string(buf[:n])
}

// Listener listens for interrupt signals, registers interrupt callbacks, and
// responds to custom shutdown signals as required
func Listener() {
	invokeCallbacks := func() {
		var callSrc string
		for i := range interruptCallbackSources {
			callSrc += fmt.Sprintf("\n-> %s running callback %d",
				strings.Split(interruptCallbackSources[i], indra.PathBase)[0],
				i)
		}
		log.T.Ln("running interrupt callbacks", callSrc)
		// run handlers in LIFO order.
		for i := range interruptCallbacks {
			idx := len(interruptCallbacks) - 1 - i
			interruptCallbacks[idx]()
		}
		log.T.Ln("interrupt handlers finished")
		close(HandlersDone)
		if Restart {
			var file string
			var e error
			file, e = osext.Executable()
			if e != nil {
				log.I.Ln(e)
				return
			}
			log.I.Ln("restarting")
			if runtime.GOOS != "windows" {
				e = unix.Exec(file, os.Args, os.Environ())
				if e != nil {
					log.I.Ln(e)
				}
			} else {
				log.I.Ln("doing windows restart")
				var s []string
				s = append(s, os.Args[0])
				s = append(s, os.Args[1:]...)
				cmd := exec.Command(s[0], s[1:]...)
				log.I.Ln("windows restart done")
				if e = cmd.Start(); e != nil {
					log.I.Ln(e)
				}
			}
		}
	}
out:
	for {
		select {
		case sig := <-ch:
			fmt.Print("\r  \r")
			log.W.F("received signal %v", sig)
			requested.Store(true)
			invokeCallbacks()
			break out
		case <-ShutdownRequestChan:
			log.I.Ln("received shutdown request - shutting down...")
			requested.Store(true)
			invokeCallbacks()
			break out
		case handler := <-addHandlerChan:
			interruptCallbacks =
				append(interruptCallbacks, handler.Fn)
			interruptCallbackSources =
				append(interruptCallbackSources, handler.Source)
		case <-HandlersDone:
			break out
		}
	}
}

// Request programmatically requests a shutdown
func Request() {
	_, f, l, _ := runtime.Caller(1)
	log.I.Ln("interrupt requested", f, l, requested.Load())
	if requested.Load() {
		log.I.Ln("requested again")
		return
	}
	requested.Store(true)
	close(ShutdownRequestChan)
	var ok bool
	select {
	case _, ok = <-ShutdownRequestChan:
	default:
	}
	log.I.Ln("shutdownrequestchan", ok)
	if ok {
		close(ShutdownRequestChan)
	}
}

// RequestRestart sets the reset flag and requests a restart
func RequestRestart() {
	Restart = true
	log.I.Ln("requesting restart")
	Request()
}

// Requested returns true if an interrupt has been requested
func Requested() bool { return requested.Load() }
