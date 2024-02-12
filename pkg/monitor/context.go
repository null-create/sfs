package monitor

import (
	"io/fs"
	"path/filepath"
)

// used for keeping track of current
// items in the directory as its being monitored
type DirCtx struct {
	currItems map[string]EItem
}

func NewDirCtx() *DirCtx {
	return &DirCtx{
		currItems: make(map[string]EItem),
	}
}

func (ctx *DirCtx) Clear() {
	ctx.currItems = nil
	ctx.currItems = make(map[string]EItem)
}

func (ctx *DirCtx) HasItem(itemName string) bool {
	if _, ok := ctx.currItems[itemName]; ok {
		return true
	}
	return false
}

// adds all new fs.DirEntry objects to the current context and returns
// a slice of the newly added entries.
func (ctx *DirCtx) AddItems(new []fs.DirEntry, dirPath string) []EItem {
	diffs := make([]EItem, 0)
	for _, item := range new {
		if _, exist := ctx.currItems[item.Name()]; !exist {
			eitem := EItem{
				name: item.Name(),
				Path: filepath.Join(dirPath, item.Name()),
			}
			diffs = append(diffs, eitem)
			ctx.currItems[eitem.Name()] = eitem
		}
	}
	return diffs
}
