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

// A slice of Item objects that implements sort.Interface to sort by Value.
type ItemList []Item

func (l ItemList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ItemList) Len() int           { return len(l) }
func (l ItemList) Less(i, j int) bool { return l[i].Size < l[j].Size }

// sort a map of files by file size and return as an ItemList
// ItemList is a slice of Item objects containing a pointer to the file object,
// and its size. These types are primarily used just for sorting.
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
