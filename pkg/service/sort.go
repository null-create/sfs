package service

import "sort"

/*
implementations based off
https://stackoverflow.com/questions/18695346/how-can-i-sort-a-mapstringint-by-its-values
*/

// --------------- sorted batch building --------------------

type Item struct {
	File *File // file object
	Size int64 // size of file
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type ItemList []Item

func (p ItemList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ItemList) Len() int           { return len(p) }
func (p ItemList) Less(i, j int) bool { return p[i].Size < p[j].Size }

// sort a map of files by file size and return as a PairList
func (b *Batch) SortMapByValue(m map[*File]int64) ItemList {
	i := 0
	p := make(ItemList, len(m))
	for k, v := range m {
		p[i] = Item{k, v}
		i++
	}
	sort.Sort(p)
	return p
}
