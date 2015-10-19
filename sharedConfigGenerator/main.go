package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mindflavor/goimgshare/folders/physical"
)

var autMails map[string]bool
var pfs physical.Folders
var id int

func main() {
	root := os.Args[1]
	id = 0

	autMails := make(map[string]bool)
	for _, item := range os.Args[2:] {
		autMails[item] = true
	}

	pfs = physical.New()

	addFolder(root)

	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(pfs)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("%s", buf.Bytes())
}

func addFolder(path string) {
	log.Printf("Adding %s", path)

	// add item
	f := physical.Folder{Path: path, AuthorizedMails: autMails}
	f.ID = fmt.Sprintf("%d", id)
	id++

	pfs[f.ID] = f

	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			addFolder(filepath.Join(path, file.Name()))
		}
	}
}
