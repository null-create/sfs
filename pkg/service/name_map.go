package service

// used to store the association between a file's name and its UUID
//
// key = UUID, value = file name (TODO: or file path?)
type NameMap map[string]string

func newNameMap(file string, uuid string) NameMap {
	nm := make(NameMap, 1)
	nm[uuid] = file
	return nm
}
