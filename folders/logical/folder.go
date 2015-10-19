package logical

// Folder is a logical representation only
// of a shared file
type Folder struct {
	ID   string
	Name string
}

// Folders is the folder list type
type Folders []*Folder
