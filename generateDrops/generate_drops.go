package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

type Drop struct {
	IsMesos bool `json:"isMesos"`
	ItemID  int  `json:"itemId"`
	Min     int  `json:"min"`
	Max     int  `json:"max"`
	QuestID int  `json:"questId"`
	Chance  int  `json:"chance"`
}

var (
	// Matches lines like:
	// INSERT INTO drops (monster, dropId, item, minCount, maxCount, money, prob) VALUES (1110101, 2, 2000000, 0, 0, 0, 20000)
	insertRe = regexp.MustCompile(`(?i)insert\s+into\s+drops\s*\(\s*monster\s*,\s*dropId\s*,\s*item\s*,\s*minCount\s*,\s*maxCount\s*,\s*money\s*,\s*prob\s*\)\s*values\s*\(\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*,\s*([+-]?\d+)\s*\)\s*;?\s*$`)
)

func main() {
	var dropsDir string
	var outPath string
	var nxPath string

	flag.StringVar(&dropsDir, "dropsDir", "drops", "Path to the directory containing .sql files with drop INSERTs")
	flag.StringVar(&outPath, "out", "drops.json", "Output JSON file path")
	flag.StringVar(&nxPath, "nx", "", "Optional path to NX file (containing /Item and /Quest) to attach questId to quest items and filter invalid items")
	flag.Parse()

	// Parse SQL-based drops
	drops, err := parseDropsDir(dropsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Optionally enrich with NX: filter invalid items and attach quest IDs
	if strings.TrimSpace(nxPath) != "" {
		if err := enrichWithQuestIDs(nxPath, drops); err != nil {
			fmt.Fprintf(os.Stderr, "warning: NX enrichment failed: %v (continuing without NX filtering/questId mapping)\n", err)
		}
	}

	if err := writeJSON(outPath, drops); err != nil {
		fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %s\n", outPath)
}

func parseDropsDir(dir string) (map[int][]Drop, error) {
	result := make(map[int][]Drop)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
			return nil
		}
		if err := parseSQLFile(path, result); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseSQLFile(path string, acc map[int][]Drop) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Increase buffer in case of very long lines
	const maxCap = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, maxCap)

	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		m := insertRe.FindStringSubmatch(line)
		if m == nil {
			// Not a matching INSERT; skip silently (the file may contain other SQL)
			continue
		}

		monster, err := atoi(m[1])
		if err != nil {
			return fmt.Errorf("line %d: monster: %w", lineNo, err)
		}
		// dropId := m[2] // ignored
		item, err := atoi(m[3])
		if err != nil {
			return fmt.Errorf("line %d: item: %w", lineNo, err)
		}
		minCount, err := atoi(m[4])
		if err != nil {
			return fmt.Errorf("line %d: minCount: %w", lineNo, err)
		}
		maxCount, err := atoi(m[5])
		if err != nil {
			return fmt.Errorf("line %d: maxCount: %w", lineNo, err)
		}
		// money := m[6] // ignored
		prob, err := atoi(m[7])
		if err != nil {
			return fmt.Errorf("line %d: prob: %w", lineNo, err)
		}

		d := Drop{
			IsMesos: item == 0,
			ItemID:  item,
			Min:     minCount,
			Max:     maxCount,
			QuestID: 0,
			Chance:  prob,
		}

		acc[monster] = append(acc[monster], d)
	}

	if err := sc.Err(); err != nil {
		return err
	}

	return nil
}

func writeJSON(path string, data map[int][]Drop) error {
	// Default encoder without extra whitespace, consistent with existing style.
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func atoi(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// ===================== NX quest enrichment & filtering =====================

type nxArchive struct {
	nodes []gonx.Node
	text  []string
}

// enrichWithQuestIDs loads NX data, filters out items that don't exist in NX, and sets QuestID for quest items.
func enrichWithQuestIDs(nxPath string, drops map[int][]Drop) error {
	arc, err := openNX(nxPath)
	if err != nil {
		return err
	}

	// Build set of all valid items from NX
	existingItems := make(map[int]struct{}, 50000)
	if err := collectAllItemIDs(arc, existingItems); err != nil {
		return fmt.Errorf("collect all items: %w", err)
	}

	// Filter: remove any drop where itemId != 0 and not in existingItems
	for mob, list := range drops {
		kept := list[:0]
		for _, d := range list {
			if d.ItemID == 0 {
				kept = append(kept, d)
				continue
			}
			if _, ok := existingItems[d.ItemID]; ok {
				kept = append(kept, d)
			}
		}
		if len(kept) == 0 {
			// keep empty slice for mob to match output shape
			drops[mob] = []Drop{}
		} else {
			drops[mob] = kept
		}
	}

	// Build item quest flag map (itemId -> isQuestItem)
	questItem := make(map[int]bool, 2048)
	if err := collectQuestItems(arc, questItem); err != nil {
		return fmt.Errorf("collect quest items: %w", err)
	}

	// Build item -> questId mapping from Quest/Check.img requirements
	itemToQuest := make(map[int]int, 4096)
	if err := collectItemQuestMap(arc, itemToQuest); err != nil {
		return fmt.Errorf("collect item->quest map: %w", err)
	}

	// Apply quest IDs to remaining drops
	for _, list := range drops {
		for i := range list {
			d := &list[i]
			if d.ItemID == 0 {
				continue
			}
			if !questItem[d.ItemID] {
				continue
			}
			if qid, ok := itemToQuest[d.ItemID]; ok {
				d.QuestID = qid
			}
		}
	}

	return nil
}

// openNX loads an NX archive and returns node and string tables.
func openNX(path string) (*nxArchive, error) {
	nodes, text, _, _, err := gonx.Parse(path)
	if err != nil {
		return nil, err
	}
	return &nxArchive{nodes: nodes, text: text}, nil
}

// collectAllItemIDs traverses NX /Item and /Character to gather every known item ID.
func collectAllItemIDs(arc *nxArchive, out map[int]struct{}) error {
	addByName := func(name string) {
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if id, err := strconv.Atoi(name); err == nil {
			out[id] = struct{}{}
		}
	}

	// Character equips: /Character/<Category>/<ItemID>.img/info
	characterRoots := []string{
		"/Character/Accessory", "/Character/Cap", "/Character/Cape", "/Character/Coat",
		"/Character/Face", "/Character/Glove", "/Character/Hair", "/Character/Longcoat",
		"/Character/Pants", "/Character/PetEquip", "/Character/Ring", "/Character/Shield",
		"/Character/Shoes", "/Character/Weapon",
	}
	for _, base := range characterRoots {
		_ = gonx.FindNode(base, arc.nodes, arc.text, func(node *gonx.Node) {
			iterateChildren(node, arc, func(child gonx.Node, name string) {
				addByName(name)
			})
		})
	}

	// Item groups: /Item/<Group>/<Sub>/<ItemID>.img
	itemGroups := []string{"/Item/Cash", "/Item/Etc", "/Item/Install", "/Item/Consume", "/Item/Pet"}
	for _, base := range itemGroups {
		_ = gonx.FindNode(base, arc.nodes, arc.text, func(node *gonx.Node) {
			// Some have subgroup level, some don't (Pet)
			iterateChildren(node, arc, func(group gonx.Node, groupName string) {
				if group.ChildCount == 0 {
					// leaf items directly under base
					addByName(groupName)
					return
				}
				iterateChildren(&group, arc, func(item gonx.Node, name string) {
					addByName(name)
				})
			})
		})
	}

	return nil
}

// collectQuestItems traverses NX /Item and marks items that have "quest" == 1 (or non-zero)
func collectQuestItems(arc *nxArchive, out map[int]bool) error {
	// Character equips have paths like /Character/<Category>/<ItemID>.img/info
	characterRoots := []string{
		"/Character/Accessory", "/Character/Cap", "/Character/Cape", "/Character/Coat",
		"/Character/Face", "/Character/Glove", "/Character/Hair", "/Character/Longcoat",
		"/Character/Pants", "/Character/PetEquip", "/Character/Ring", "/Character/Shield",
		"/Character/Shoes", "/Character/Weapon",
	}
	for _, base := range characterRoots {
		_ = gonx.FindNode(base, arc.nodes, arc.text, func(node *gonx.Node) {
			iterateChildren(node, arc, func(child gonx.Node, name string) {
				infoPath := base + "/" + name + "/info"
				_ = gonx.FindNode(infoPath, arc.nodes, arc.text, func(info *gonx.Node) {
					if hasQuestFlag(info, arc) {
						if id, ok := parseIDFromName(name); ok {
							out[id] = true
						}
					}
				})
			})
		})
	}

	// Item groups: /Item/<Group>/<Sub>/<ItemID>.img/info (Cash/Etc/Install/Consume) and /Item/Pet/<ItemID>.img/info
	itemGroups := []string{"/Item/Cash", "/Item/Etc", "/Item/Install"}
	for _, base := range itemGroups {
		_ = gonx.FindNode(base, arc.nodes, arc.text, func(node *gonx.Node) {
			iterateChildren(node, arc, func(group gonx.Node, groupName string) {
				iterateChildren(&group, arc, func(item gonx.Node, name string) {
					infoPath := base + "/" + groupName + "/" + name + "/info"
					_ = gonx.FindNode(infoPath, arc.nodes, arc.text, func(info *gonx.Node) {
						if hasQuestFlag(info, arc) {
							if id, ok := parseIDFromName(name); ok {
								out[id] = true
							}
						}
					})
				})
			})
		})
	}

	consumeBase := "/Item/Consume"
	_ = gonx.FindNode(consumeBase, arc.nodes, arc.text, func(node *gonx.Node) {
		iterateChildren(node, arc, func(group gonx.Node, groupName string) {
			iterateChildren(&group, arc, func(item gonx.Node, name string) {
				infoPath := consumeBase + "/" + groupName + "/" + name + "/info"
				_ = gonx.FindNode(infoPath, arc.nodes, arc.text, func(info *gonx.Node) {
					if hasQuestFlag(info, arc) {
						if id, ok := parseIDFromName(name); ok {
							out[id] = true
						}
					}
				})
			})
		})
	})

	petBase := "/Item/Pet"
	_ = gonx.FindNode(petBase, arc.nodes, arc.text, func(node *gonx.Node) {
		iterateChildren(node, arc, func(item gonx.Node, name string) {
			infoPath := petBase + "/" + name + "/info"
			_ = gonx.FindNode(infoPath, arc.nodes, arc.text, func(info *gonx.Node) {
				if hasQuestFlag(info, arc) {
					if id, ok := parseIDFromName(name); ok {
						out[id] = true
					}
				}
			})
		})
	})

	return nil
}

// collectItemQuestMap walks /Quest/Check.img and builds a mapping of itemId -> questId by scanning required items.
func collectItemQuestMap(arc *nxArchive, out map[int]int) error {
	const root = "/Quest/Check.img"
	_ = gonx.FindNode(root, arc.nodes, arc.text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			questDir := arc.nodes[n.ChildID+i]
			raw := arc.text[questDir.NameID]
			qname := strings.TrimSuffix(raw, filepath.Ext(raw))
			qid, err := strconv.Atoi(qname)
			if err != nil {
				continue
			}

			for j := uint32(0); j < uint32(questDir.ChildCount); j++ {
				phaseDir := arc.nodes[questDir.ChildID+j]
				for k := uint32(0); k < uint32(phaseDir.ChildCount); k++ {
					entry := arc.nodes[phaseDir.ChildID+k]
					if arc.text[entry.NameID] != "item" {
						continue
					}
					for m := uint32(0); m < uint32(entry.ChildCount); m++ {
						itemNode := arc.nodes[entry.ChildID+m]
						var itemID int
						for x := uint32(0); x < uint32(itemNode.ChildCount); x++ {
							field := arc.nodes[itemNode.ChildID+x]
							if arc.text[field.NameID] == "id" {
								itemID = int(gonx.DataToInt32(field.Data))
							}
						}
						if itemID != 0 {
							if _, exists := out[itemID]; !exists {
								out[itemID] = qid
							}
						}
					}
				}
			}
		}
	})
	return nil
}

// hasQuestFlag checks if an "info" node contains a non-zero "quest" field.
func hasQuestFlag(info *gonx.Node, arc *nxArchive) bool {
	for i := uint32(0); i < uint32(info.ChildCount); i++ {
		ch := arc.nodes[info.ChildID+i]
		if arc.text[ch.NameID] == "quest" && ch.ChildCount == 0 {
			return gonx.DataToInt64(ch.Data) != 0
		}
	}
	return false
}

// iterateChildren iterates direct children and resolves their names.
func iterateChildren(node *gonx.Node, arc *nxArchive, fn func(child gonx.Node, name string)) {
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		child := arc.nodes[node.ChildID+i]
		name := arc.text[child.NameID]
		fn(child, name)
	}
}

// parseIDFromName extracts an integer id from a node name that may include an extension, e.g. "2000000.img".
func parseIDFromName(name string) (int, bool) {
	trimmed := strings.TrimSuffix(name, filepath.Ext(name))
	id, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return id, true
}
