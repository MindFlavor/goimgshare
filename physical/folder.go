package physical

import (
	"encoding/json"
	"io"

	"github.com/mindflavor/testauth/logical"
)

// File is a phyisical representation
// of a shared file
type Folder struct {
	*logical.Folder
	Path string
}

type Folders []*Folder

// Save allows you to save the slice
func (f Folders) Save(w io.Writer) {
	json.NewEncoder(w).Encode(f)
}

// Load loads the slice from the reader
func Load(r io.Reader) (Folders, error) {
	var ps []*Folder
	if err := json.NewDecoder(r).Decode(&ps); err != nil {
		return nil, err
	}

	return ps, nil
}

func (f Folders) ToLogical() logical.Folders {
	log := make([]*logical.Folder, 0, len(f))

	for _, item := range f {
		log = append(log, item.Folder)
	}

	return log
}
