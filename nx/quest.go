package nx

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// Quest data
type Quest struct {
	ID     int16
	Name   string
	Parent string
	Order  int16
	Area   int32

	Journal map[int]string

	// Requirements
	Start    CheckBlock // from Check[questID]["0"]
	Complete CheckBlock // from Check[questID]["1"]

	// Actions / rewards
	ActOnStart    ActBlock // from Act[questID]["0"]
	ActOnComplete ActBlock // from Act[questID]["1"]

	// Keys like: "start.0","start.1","complete.0","start.yes.0", etc.
	Say map[string][]string
}

type CheckBlock struct {
	NPC   int32
	Job   int32
	LvMin int32
	LvMax int32
	Pop   int32

	PrevQuests []QuestStateReq
	Items      []ReqItem
	Mobs       []ReqMob
}

type QuestStateReq struct {
	ID    int16
	State int8
}

type ReqItem struct {
	ID    int32
	Count int32
}

type ReqMob struct {
	ID    int32
	Count int32
}

type ActBlock struct {
	Exp       int32
	Money     int32
	Pop       int32
	NextQuest int16
	Fame      int32

	Items []ActItem
}

type ActItem struct {
	ID     int32
	Count  int32
	Prop   int32
	Job    int32
	Gender int32
}

// extractQuests builds the quests map by traversing four quest images under /Quest
func extractQuests(nodes []gonx.Node, text []string) map[int16]Quest {
	out := make(map[int16]Quest)

	parseQuestInfo(out, nodes, text)
	parseQuestCheck(out, nodes, text)
	parseQuestAct(out, nodes, text)
	parseQuestSay(out, nodes, text)

	return out
}

func parseQuestInfo(out map[int16]Quest, nodes []gonx.Node, text []string) {
	const root = "/Quest/QuestInfo.img"
	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			qid64, err := strconv.ParseInt(name, 10, 16)
			if err != nil {
				continue
			}
			q := out[int16(qid64)]
			q.ID = int16(qid64)
			if q.Journal == nil {
				q.Journal = make(map[int]string, 4)
			}

			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				ch := nodes[dir.ChildID+j]
				key := text[ch.NameID]
				switch key {
				case "name":
					q.Name = text[gonx.DataToUint32(ch.Data)]
				case "parent":
					q.Parent = text[gonx.DataToUint32(ch.Data)]
				case "order":
					q.Order = int16(gonx.DataToInt32(ch.Data))
				case "area":
					q.Area = gonx.DataToInt32(ch.Data)
				default:
					if idx, err := strconv.Atoi(key); err == nil {
						q.Journal[idx] = text[gonx.DataToUint32(ch.Data)]
					}
				}
			}
			out[q.ID] = q
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
}

func parseQuestCheck(out map[int16]Quest, nodes []gonx.Node, text []string) {
	const root = "/Quest/Check.img"
	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			qid64, err := strconv.ParseInt(name, 10, 16)
			if err != nil {
				continue
			}
			q := out[int16(qid64)]
			q.ID = int16(qid64)

			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				phaseDir := nodes[dir.ChildID+j]
				phaseName := text[phaseDir.NameID] // "0" (start) or "1" (complete)
				block := CheckBlock{}
				for k := uint32(0); k < uint32(phaseDir.ChildCount); k++ {
					entry := nodes[phaseDir.ChildID+k]
					key := text[entry.NameID]
					switch key {
					case "npc":
						block.NPC = gonx.DataToInt32(entry.Data)
					case "job":
						block.Job = gonx.DataToInt32(entry.Data)
					case "lvmin":
						block.LvMin = gonx.DataToInt32(entry.Data)
					case "lvmax":
						block.LvMax = gonx.DataToInt32(entry.Data)
					case "pop":
						block.Pop = gonx.DataToInt32(entry.Data)
					case "item":
						block.Items = append(block.Items, parseReqItems(&entry, nodes, text)...)
					case "mob":
						block.Mobs = append(block.Mobs, parseReqMobs(&entry, nodes, text)...)
					case "quest":
						block.PrevQuests = append(block.PrevQuests, parseReqQuests(&entry, nodes, text)...)
					}
				}
				if phaseName == "0" {
					q.Start = block
				} else if phaseName == "1" {
					q.Complete = block
				}
			}
			out[q.ID] = q
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
}

func parseReqItems(node *gonx.Node, nodes []gonx.Node, text []string) []ReqItem {
	var ret []ReqItem
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		dir := nodes[node.ChildID+i]
		var it ReqItem
		for j := uint32(0); j < uint32(dir.ChildCount); j++ {
			f := nodes[dir.ChildID+j]
			switch text[f.NameID] {
			case "id":
				it.ID = gonx.DataToInt32(f.Data)
			case "count":
				it.Count = gonx.DataToInt32(f.Data)
			}
		}
		if it.ID != 0 {
			ret = append(ret, it)
		}
	}
	return ret
}

func parseJobList(node *gonx.Node, nodes []gonx.Node, text []string) []int32 {
	out := make([]int32, 0, node.ChildCount)
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		// children are typically numbered ("0","1",...) with a scalar int job id
		if ch.ChildCount == 0 {
			out = append(out, gonx.DataToInt32(ch.Data))
		} else {
			// some builds wrap the value in a nested dir with a single child "job": value
			for j := uint32(0); j < uint32(ch.ChildCount); j++ {
				inner := nodes[ch.ChildID+j]
				if text[inner.NameID] == "job" && inner.ChildCount == 0 {
					out = append(out, gonx.DataToInt32(inner.Data))
				}
			}
		}
	}
	return out
}

func parseReqMobs(node *gonx.Node, nodes []gonx.Node, text []string) []ReqMob {
	var ret []ReqMob
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		dir := nodes[node.ChildID+i]
		var m ReqMob
		for j := uint32(0); j < uint32(dir.ChildCount); j++ {
			f := nodes[dir.ChildID+j]
			switch text[f.NameID] {
			case "id":
				m.ID = gonx.DataToInt32(f.Data)
			case "count":
				m.Count = gonx.DataToInt32(f.Data)
			}
		}
		if m.ID != 0 {
			ret = append(ret, m)
		}
	}
	return ret
}

func parseReqQuests(node *gonx.Node, nodes []gonx.Node, text []string) []QuestStateReq {
	var ret []QuestStateReq
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		dir := nodes[node.ChildID+i]
		var qr QuestStateReq
		for j := uint32(0); j < uint32(dir.ChildCount); j++ {
			f := nodes[dir.ChildID+j]
			switch text[f.NameID] {
			case "id":
				qr.ID = int16(gonx.DataToInt32(f.Data))
			case "state":
				qr.State = int8(gonx.DataToInt32(f.Data))
			}
		}
		if qr.ID != 0 {
			ret = append(ret, qr)
		}
	}
	return ret
}

func parseQuestAct(out map[int16]Quest, nodes []gonx.Node, text []string) {
	const root = "/Quest/Act.img"
	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			qid64, err := strconv.ParseInt(name, 10, 16)
			if err != nil {
				continue
			}
			q := out[int16(qid64)]
			q.ID = int16(qid64)

			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				phaseDir := nodes[dir.ChildID+j]
				phaseName := text[phaseDir.NameID] // "0" (on accept) or "1" (on complete)
				block := ActBlock{}
				for k := uint32(0); k < uint32(phaseDir.ChildCount); k++ {
					entry := nodes[phaseDir.ChildID+k]
					key := text[entry.NameID]
					switch key {
					case "exp":
						block.Exp = gonx.DataToInt32(entry.Data)
					case "money":
						block.Money = gonx.DataToInt32(entry.Data)
					case "pop":
						block.Pop = gonx.DataToInt32(entry.Data)
						block.Fame = block.Pop
					case "nextQuest":
						block.NextQuest = int16(gonx.DataToInt32(entry.Data))
					case "item":
						block.Items = append(block.Items, parseActItems(&entry, nodes, text)...)
					}
				}
				if phaseName == "0" {
					q.ActOnStart = block
				} else if phaseName == "1" {
					q.ActOnComplete = block
				}
			}
			out[q.ID] = q
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
}

func parseActItems(node *gonx.Node, nodes []gonx.Node, text []string) []ActItem {
	var ret []ActItem
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		dir := nodes[node.ChildID+i]
		var ai ActItem
		for j := uint32(0); j < uint32(dir.ChildCount); j++ {
			f := nodes[dir.ChildID+j]
			switch text[f.NameID] {
			case "id":
				ai.ID = gonx.DataToInt32(f.Data)
			case "count":
				ai.Count = gonx.DataToInt32(f.Data)
			case "prop":
				ai.Prop = gonx.DataToInt32(f.Data)
			case "job":
				ai.Job = gonx.DataToInt32(f.Data)
			case "gender":
				ai.Gender = gonx.DataToInt32(f.Data)
			}
		}
		if ai.ID != 0 {
			ret = append(ret, ai)
		}
	}
	return ret
}

func parseQuestSay(out map[int16]Quest, nodes []gonx.Node, text []string) {
	const root = "/Quest/Say.img"
	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			qid64, err := strconv.ParseInt(name, 10, 16)
			if err != nil {
				continue
			}
			q := out[int16(qid64)]
			q.ID = int16(qid64)
			if q.Say == nil {
				q.Say = make(map[string][]string, 8)
			}

			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				phaseDir := nodes[dir.ChildID+j]
				phaseName := text[phaseDir.NameID] // "0"(start) or "1"(complete)
				prefix := "start"
				if phaseName == "1" {
					prefix = "complete"
				}
				// Walk children; leaf nodes (ChildCount==0) are strings
				for k := uint32(0); k < uint32(phaseDir.ChildCount); k++ {
					entry := nodes[phaseDir.ChildID+k]
					key := text[entry.NameID]
					if entry.ChildCount == 0 {
						// numbered lines "0","1","2",...
						if _, err := strconv.Atoi(key); err == nil {
							q.Say[prefix+"."+key] = append(q.Say[prefix+"."+key], text[gonx.DataToUint32(entry.Data)])
						}
					} else {
						// nested branches like yes/no/stop/lost; collect recursively
						collectSayStrings(&entry, nodes, text, prefix+"."+key, q.Say)
					}
				}
			}
			out[q.ID] = q
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
}

func collectSayStrings(node *gonx.Node, nodes []gonx.Node, text []string, base string, acc map[string][]string) {
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		ch := nodes[node.ChildID+i]
		key := text[ch.NameID]
		if ch.ChildCount == 0 {
			if _, err := strconv.Atoi(key); err == nil {
				acc[base] = append(acc[base], text[gonx.DataToUint32(ch.Data)])
			}
		} else {
			collectSayStrings(&ch, nodes, text, base+"."+key, acc)
		}
	}
}
