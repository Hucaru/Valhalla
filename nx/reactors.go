package nx

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// ReactorEventInfo represents a single event entry under a reactor state's "event" node.
type ReactorEventInfo struct {
	Type        int32
	State       int32
	ReqItemID   int32 // int name "0"
	ReqItemCnt  int32 // int name "1"
	Probability int32 // int name "2"
	LT          NXVec2
	RB          NXVec2

	ExtraInts map[string]int32
	ExtraVecs map[string]NXVec2
}

// ReactorStateInfo is a reactor state (e.g., "0", "1") and its events.
type ReactorStateInfo struct {
	Events []ReactorEventInfo
}

// ReactorInfo is the reactor data for a single ID.
type ReactorInfo struct {
	ID     int32
	Info   string // /Reactor/<id>.img/info/info
	Action string // /Reactor/<id>.img/action
	States map[int]ReactorStateInfo
	LinkTo int32 // /Reactor/<id>.img/info/link -> other reactor id
}

// NXVec2 is a simple integer vector for NX vector lookups
type NXVec2 struct {
	X int16
	Y int16
}

func extractReactors(nodes []gonx.Node, textLookup []string) map[int32]ReactorInfo {
	out := make(map[int32]ReactorInfo)

	base := "/Reactor"
	ok := gonx.FindNode(base, nodes, textLookup, func(root *gonx.Node) {
		iterateChildren(root, nodes, textLookup, func(reactorNode gonx.Node, fileName string) {
			reactorID, valid := parseIDFromName(fileName)
			if !valid {
				return
			}

			r := ReactorInfo{
				ID:     reactorID,
				States: make(map[int]ReactorStateInfo),
			}

			for i := uint32(0); i < uint32(reactorNode.ChildCount); i++ {
				child := nodes[reactorNode.ChildID+i]
				childName := textLookup[child.NameID]

				switch childName {
				case "info":
					r.Info, r.LinkTo = getReactorInfo(&child, nodes, textLookup)
				case "action":
					r.Action = getStringValue(&child, nodes, textLookup)
				default:
					if stateIdx, err := strconv.Atoi(childName); err == nil {
						state := parseReactorState(&child, nodes, textLookup)
						r.States[stateIdx] = state
					}
				}
			}

			out[reactorID] = r
		})
	})

	if !ok {
		log.Println("Invalid node search:", base)
	}

	for id, r := range out {
		if r.LinkTo == 0 {
			continue
		}
		target, ok := out[r.LinkTo]
		if !ok {
			continue
		}

		if r.Info == "" {
			r.Info = target.Info
		}
		if r.Action == "" {
			r.Action = target.Action
		}
		if len(r.States) == 0 && len(target.States) > 0 {
			r.States = cloneStates(target.States)
		} else {
			for sIdx, s := range target.States {
				if _, exists := r.States[sIdx]; !exists {
					r.States[sIdx] = cloneState(s)
				}
			}
		}

		out[id] = r
	}

	return out
}

func cloneStates(src map[int]ReactorStateInfo) map[int]ReactorStateInfo {
	dst := make(map[int]ReactorStateInfo, len(src))
	for k, v := range src {
		dst[k] = cloneState(v)
	}
	return dst
}

func cloneState(s ReactorStateInfo) ReactorStateInfo {
	if len(s.Events) == 0 {
		return ReactorStateInfo{}
	}
	evs := make([]ReactorEventInfo, len(s.Events))
	for i, e := range s.Events {
		evs[i] = cloneEvent(e)
	}
	return ReactorStateInfo{Events: evs}
}

func cloneEvent(e ReactorEventInfo) ReactorEventInfo {
	ev := e
	if e.ExtraInts != nil {
		ev.ExtraInts = make(map[string]int32, len(e.ExtraInts))
		for k, v := range e.ExtraInts {
			ev.ExtraInts[k] = v
		}
	}
	if e.ExtraVecs != nil {
		ev.ExtraVecs = make(map[string]NXVec2, len(e.ExtraVecs))
		for k, v := range e.ExtraVecs {
			ev.ExtraVecs[k] = v
		}
	}
	return ev
}

func parseIDFromName(name string) (int32, bool) {
	trimmed := strings.TrimSuffix(name, filepath.Ext(name))
	id64, err := strconv.ParseInt(trimmed, 10, 32)
	if err != nil {
		log.Println("Invalid reactor id name:", name, "err:", err)
		return 0, false
	}
	return int32(id64), true
}

func getReactorInfo(node *gonx.Node, nodes []gonx.Node, textLookup []string) (info string, linkTo int32) {
	// /info nodes can contain "info" and optionally "link"
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		name := textLookup[ch.NameID]
		switch name {
		case "info":
			info = getStringValue(&ch, nodes, textLookup)
		case "link":
			linkStr := getStringValue(&ch, nodes, textLookup)
			if id, err := strconv.Atoi(strings.TrimSpace(linkStr)); err == nil {
				linkTo = int32(id)
			}
		}
	}
	return
}

func parseReactorState(node *gonx.Node, nodes []gonx.Node, textLookup []string) ReactorStateInfo {
	st := ReactorStateInfo{}
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		name := textLookup[ch.NameID]
		if name == "event" {
			st.Events = parseReactorEvents(&ch, nodes, textLookup)
		}
	}
	return st
}

func parseReactorEvents(eventNode *gonx.Node, nodes []gonx.Node, textLookup []string) []ReactorEventInfo {
	events := make([]ReactorEventInfo, 0)

	for i := uint32(0); i < uint32(eventNode.ChildCount); i++ {
		evDir := nodes[eventNode.ChildID+i]
		_ = textLookup[evDir.NameID] // index, not needed

		ev := ReactorEventInfo{
			ExtraInts: make(map[string]int32),
			ExtraVecs: make(map[string]NXVec2),
		}

		for j := uint32(0); j < uint32(evDir.ChildCount); j++ {
			opt := nodes[evDir.ChildID+j]
			name := textLookup[opt.NameID]

			switch name {
			case "type":
				ev.Type = gonx.DataToInt32(opt.Data)
			case "state":
				ev.State = gonx.DataToInt32(opt.Data)
			case "0":
				ev.ReqItemID = gonx.DataToInt32(opt.Data)
			case "1":
				ev.ReqItemCnt = gonx.DataToInt32(opt.Data)
			case "2":
				ev.Probability = gonx.DataToInt32(opt.Data)
			case "lt":
				ev.LT = readNXVec2(&opt, nodes, textLookup)
			case "rb":
				ev.RB = readNXVec2(&opt, nodes, textLookup)
			default:
				if isVectorNode(&opt) {
					ev.ExtraVecs[name] = readNXVec2(&opt, nodes, textLookup)
				} else if len(opt.Data) > 0 {
					ev.ExtraInts[name] = gonx.DataToInt32(opt.Data)
				}
			}
		}

		events = append(events, ev)
	}
	return events
}

func isVectorNode(node *gonx.Node) bool {
	if node.ChildCount == 0 {
		return false
	}

	return true == func() bool {
		for i := uint32(0); i < uint32(node.ChildCount); i++ {
			ch := node.ChildID + i
			_ = ch
		}
		return true
	}()
}

func readNXVec2(node *gonx.Node, nodes []gonx.Node, textLookup []string) NXVec2 {
	v := NXVec2{}
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		name := textLookup[ch.NameID]
		switch name {
		case "x":
			v.X = gonx.DataToInt16(ch.Data)
		case "y":
			v.Y = gonx.DataToInt16(ch.Data)
		}
	}
	return v
}

// getStringValue reads a string value from a node whose value is stored in Data as a string ref
func getStringValue(node *gonx.Node, nodes []gonx.Node, textLookup []string) string {
	if len(node.Data) > 0 {
		return textLookup[gonx.DataToUint32(node.Data)]
	}
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		if len(ch.Data) > 0 {
			return textLookup[gonx.DataToUint32(ch.Data)]
		}
	}
	return ""
}
