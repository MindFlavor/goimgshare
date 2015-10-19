package physical

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mindflavor/goimgshare/folders/logical"
)

// Folder is a phyisical representation
// of a shared file
type Folder struct {
	*logical.Folder
	Path            string
	AuthorizedMails map[string]bool
}

// Folders is the folder list type
type Folders map[string]Folder

// New creates a new Folders list
func New() Folders {
	return make(Folders)
}

// Save allows you to save the slice
func (f Folders) Save(w io.Writer) {
	// convert to array first
	a := make([]Folder, 0, len(f))
	for _, val := range f {
		a = append(a, val)
	}

	json.NewEncoder(w).Encode(a)
}

// Load loads the slice from the reader
func Load(r io.Reader) (Folders, error) {
	var ps []Folder
	if err := json.NewDecoder(r).Decode(&ps); err != nil {
		return nil, err
	}

	f := make(Folders)
	for _, item := range ps {
		if _, found := f[item.ID]; found {
			return nil, fmt.Errorf("Duplicate PhysicalFolder ID found parsing the input stream: %s", item.ID)
		}

		f[item.ID] = item
	}

	return f, nil
}

// ToLogical translates the Physical folders
// into Logical ones (ie without path).
func (f Folders) ToLogical() logical.Folders {
	log := make([]*logical.Folder, 0, len(f))

	for _, item := range f {
		log = append(log, item.Folder)
	}

	return log
}

// IsAuthorized returns true if the
// mail can access folderID
func (f Folders) IsAuthorized(folderID string, mail string) bool {
	ret, found := f[folderID]
	if !found {
		return false
	}

	retAuth, found := ret.AuthorizedMails[mail]
	if !found || !retAuth {
		return false
	}

	return true
}
