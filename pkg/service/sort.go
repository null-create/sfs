package service

import "sort"

/*
implementations based off
https://stackoverflow.com/questions/18695346/how-can-i-sort-a-mapstringint-by-its-values
*/

// --------------- sorted batch building

// converts a slice of files to map[*File]int64 where the int is the files size
func (b *Batch) SliceToMap(files []*File) map[*File]int64 {
	m := make(map[*File]int64)
	for _, f := range files {
		m[f] = f.Size()
	}
	return m
}

type Pair struct {
	File *File // file object
	Size int64 // size of file
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Size < p[j].Size }

// A function to turn a map into a PairList, then sort and return it.
func (b *Batch) SortMapByValue(m map[*File]int64) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(p)
	return p
}
