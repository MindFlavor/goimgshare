package logical

// Folder is a logical representation only
// of a shared file
type Folder struct {
	ID   string
	Name string
}

// Folders is the folder list type
type Folders []*Folder

func (slice Folders) Len() int {
	return len(slice)
}

func (slice Folders) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice Folders) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
