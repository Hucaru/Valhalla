package nx

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

func GetPackages() map[int32][]int32 {
	return packages
}

func extractPackages(nodes []gonx.Node, text []string) map[int32][]int32 {
	const root = "/Etc/CashPackage.img"
	out := make(map[int32][]int32)

	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			pkgSN64, err := strconv.ParseInt(name, 10, 32)
			if err != nil {
				continue
			}
			var list []int32
			// Some formats use child "SN" with numbered children
			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				ch := nodes[dir.ChildID+j]
				if text[ch.NameID] != "SN" {
					continue
				}
				for k := uint32(0); k < uint32(ch.ChildCount); k++ {
					entry := nodes[ch.ChildID+k]
					if entry.ChildCount == 0 {
						list = append(list, gonx.DataToInt32(entry.Data))
					}
				}
			}
			if len(list) > 0 {
				out[int32(pkgSN64)] = list
			}
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
	return out
}
