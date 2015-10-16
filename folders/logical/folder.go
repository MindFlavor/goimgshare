package logical

// Folder is a logical representation only
// of a shared file
type Folder struct {
	ID   string
	Name string
}

type Folders []*Folder
