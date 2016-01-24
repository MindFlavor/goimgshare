package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mindflavor/goimgshare/folders/physical"
)

var autMails map[string]bool
var pfs physical.Folders
var id int
var rootName string

func main() {
	root := os.Args[1]
	outfile := os.Args[2]
	id = 0

	autMails = make(map[string]bool)
	for _, item := range os.Args[3:] {
		autMails[item] = true
	}

	pfs = physical.New()

	_, rootName = filepath.Split(root)

	addFolder(root, root)

	file, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	pfs.Save(file)
}

func addFolder(root, path string) {
	// get subpath
	subpath := path[len(root):]

	// add item
	f := physical.Folder{}
	f.Path = path
	f.AuthorizedMails = autMails

	f.ID = fmt.Sprintf("%d", id)
	f.Name = rootName + subpath
	id++

	pfs[f.ID] = f

	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			addFolder(root, filepath.Join(path, file.Name()))
		}
	}
}
