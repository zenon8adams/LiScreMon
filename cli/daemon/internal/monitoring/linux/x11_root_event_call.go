package monitoring

import (
	"fmt"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
)

var ()

func registerRootWindowForEvents(x11Conn *xgbutil.XUtil) {

	xevent.MapNotifyFun(mapNotifyEventFuncRoot).Connect(x11Conn, x11Conn.RootWin())

	xevent.PropertyNotifyFun(propertyNotifyEventFuncRoot).Connect(x11Conn, x11Conn.RootWin())
}
func propertyNotifyEventFuncRoot(x11Conn *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
	monitor.rootPropertyNotifyHandler(x11Conn, ev, netActiveWindowAtom, netClientStackingAtom)
}

func mapNotifyEventFuncRoot(x11Conn *xgbutil.XUtil, ev xevent.MapNotifyEvent) {
	fmt.Printf("\nrootMapNotifyHandler ev.window:%v ======++++++====> ev.event:%v\n", ev.Window, ev.Event)
	if ev.OverrideRedirect { // window is a popup
		return
	}

	if transientFor, err := xprop.PropValWindow(xprop.GetProperty(x11Conn, ev.Window, "WM_TRANSIENT_FOR")); err == nil && transientFor != 0 {
		fmt.Println("This window is transient for window", transientFor)
		return // window can be treated as a popup
	}

	if windowTypes, err := ewmh.WmWindowTypeGet(x11Conn, ev.Window); err == nil || len(windowTypes) >= 1 {
		for i := 0; i < len(windowTypes); i++ {
			if windowTypes[i] == "_NET_WM_WINDOW_TYPE_NORMAL" {
				// _NET_WM_WINDOW_TYPE_NORMAL indicates that this is a normal, top-level window, either managed or override-redirect.
				// Managed windows with neither _NET_WM_WINDOW_TYPE nor WM_TRANSIENT_FOR set MUST be taken as this type.
				// Override-redirect windows without _NET_WM_WINDOW_TYPE, must be taken as this type, whether or not they have WM_TRANSIENT_FOR set.
				// https://specifications.freedesktop.org/wm-spec/latest/ar01s05.html#idm45584883008224:~:text=override%2Dredirect%20windows.-,_NET_WM_WINDOW_TYPE_NORMAL,-indicates%20that%20this
				rootMapNotifyHandler(x11Conn, ev)
				return
			}
		}
	}
}
