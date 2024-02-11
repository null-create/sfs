package monitor

import "io/fs"

// used for keeping track of current
// items in the directory as its being monitored
type DirCtx struct {
	currItems map[string]fs.DirEntry
}

func NewDirCtx() *DirCtx {
	return &DirCtx{
		currItems: make(map[string]fs.DirEntry),
	}
}

func (ctx *DirCtx) Clear() {
	ctx.currItems = nil
	ctx.currItems = make(map[string]fs.DirEntry)
}

func (ctx *DirCtx) HasItem(itemName string) bool {
	if _, ok := ctx.currItems[itemName]; ok {
		return true
	}
	return false
}

func (ctx *DirCtx) AddItems(items []fs.DirEntry) {
	for _, item := range items {
		if !ctx.HasItem(item.Name()) {
			ctx.currItems[item.Name()] = item
		}
	}
}

// finds all items that aren't current being monitored and returns
// a map containing the items
func (ctx *DirCtx) GetDiffs(new []fs.DirEntry) []fs.DirEntry {
	diffs := make([]fs.DirEntry, 0)
	for _, item := range new {
		if _, exist := ctx.currItems[item.Name()]; !exist {
			diffs = append(diffs, item)
		}
	}
	return diffs
}
