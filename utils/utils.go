package utils

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"time"
)

func Own(X *xgbutil.XUtil) error {
	log.Printf("Setting selection owner...")
	var err error

	xTime, err := currentTime(X)
	if err != nil {
		return err
	}

	selAtom, err := managerAtom(X)
	if err != nil {
		return err
	}

	err = xproto.SetSelectionOwnerChecked(X.Conn(), X.Dummy(), selAtom, xTime).Check()
	if err != nil {
		return err
	}

	// Now we've got to make sure that we *actually* got ownership.
	log.Printf("Getting selection owner...")
	reply, err := xproto.GetSelectionOwner(X.Conn(), selAtom).Reply()
	if err != nil {
		return err
	}

	if reply.Owner != X.Dummy() {
		return fmt.Errorf(
			"Could not acquire ownership with SetSelectionOwner. "+
				"GetSelectionOwner claims that '%d' is the owner, but '%d' "+
				"needs to be.", reply.Owner, X.Dummy())
	}

	log.Println("ZEDSDEAD has window manager ownership!")
	announce(X)
	return err
}

// managerAtom returns an xproto.Atom of the manager selection atom.
// Usually it's "WM_S0", where "0" is the screen number.
func managerAtom(X *xgbutil.XUtil) (xproto.Atom, error) {
	name := fmt.Sprintf("WM_S%d", X.Conn().DefaultScreen)
	return xprop.Atm(X, name)
}

func announce(X *xgbutil.XUtil) {
	typAtom, err := xprop.Atm(X, "MANAGER")
	if err != nil {
		log.Println(err)
		return
	}
	manSelAtom, err := managerAtom(X)
	if err != nil {
		log.Println(err)
		return
	}
	cm, err := xevent.NewClientMessage(32, X.RootWin(), typAtom,
		int(X.TimeGet()), int(manSelAtom), int(X.Dummy()))
	xproto.SendEvent(X.Conn(), false, X.RootWin(),
		xproto.EventMaskStructureNotify, string(cm.Bytes()))
}

// currentTime forcefully causes a PropertyNotify event to fire on the root
// window, then scans the event queue and picks up the time.
//
// It is NOT SAFE to call this function in a place other than Wingo's
// initialization. Namely, this function subverts xevent's queue and reads
// events directly from X.
func currentTime(X *xgbutil.XUtil) (xproto.Timestamp, error) {
	wmClassAtom, err := xprop.Atm(X, "WM_CLASS")
	if err != nil {
		return 0, err
	}

	stringAtom, err := xprop.Atm(X, "STRING")
	if err != nil {
		return 0, err
	}

	// Make sure we're listening to PropertyChange events on the root window.
	err = xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskPropertyChange)
	if err != nil {
		return 0, fmt.Errorf(
			"Could not listen to Root window events (PropertyChange): %s", err)
	}

	// Do a zero-length append on a property as suggested by ICCCM 2.1.
	err = xproto.ChangePropertyChecked(
		X.Conn(), xproto.PropModeAppend, X.RootWin(),
		wmClassAtom, stringAtom, 8, 0, nil).Check()
	if err != nil {
		return 0, err
	}

	// Now look for the PropertyNotify generated by that zero-length append
	// and return the timestamp attached to that event.
	// Note that we do this outside of xgbutil/xevent, since ownership
	// is literally the first thing we do after connecting to X.
	// (i.e., we don't have our event handling system initialized yet.)
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-timeout:
			return 0, fmt.Errorf(
				"Expected a PropertyNotify event to get a valid timestamp, " +
					"but never received one.")
		default:
			ev, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			if propNotify, ok := ev.(xproto.PropertyNotifyEvent); ok {
				X.TimeSet(propNotify.Time) // why not?
				return propNotify.Time, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	panic("unreachable")
}
