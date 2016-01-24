package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gorilla/mux"
	"github.com/mindflavor/goimgshare/folders/logical"
)

type fnWithSize struct {
	Name string
	Size int64
}

func handleFolders(w http.ResponseWriter, r *http.Request) {
	var logFolders logical.Folders

	// add only authorized folders
	for _, pf := range phyFolders {
		if aDB.IsAuthorized(&phyFolders, r, pf.ID) {
			logFolders = append(logFolders, pf.Folder)
		}
	}

	sort.Sort(logFolders)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(logFolders)
}

func handleFolder(fnAcceptFileName func(strtomatch string) bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		folderid := mux.Vars(r)["folderid"]

		pf, found := phyFolders[folderid]
		if !found {
			http.NotFound(w, r)
			log.Printf("404 Not found: physicalFolder %s", folderid)
			return
		}

		if !aDB.IsAuthorized(&phyFolders, r, folderid) {
			s := fmt.Sprintf("403 Forbidden : you can't access this resource (%s).", folderid)
			http.Error(w, s, 403)
			log.Printf(s)
			return
		}

		files, err := ioutil.ReadDir(pf.Path)
		if err != nil {
			http.Error(w, err.Error(), 1)
			log.Printf("501 error browsing folder %s: %s", pf.ID, err)
			return
		}

		fnwithsizes := make([]fnWithSize, 0, len(files))
		for _, item := range files {
			// skip directories and !fnAcceptFileName
			if !item.IsDir() && fnAcceptFileName(item.Name()) {
				fnwithsizes = append(fnwithsizes, fnWithSize{item.Name(), item.Size()})
			}
		}

		json.NewEncoder(w).Encode(fnwithsizes)
	}
}

func handleStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	folderid := vars["folderid"]
	filename := vars["fn"]

	pf, found := phyFolders[folderid]
	if !found {
		http.NotFound(w, r)
		log.Printf("404 Not found: physicalFolder %s", folderid)
		return
	}

	if !aDB.IsAuthorized(&phyFolders, r, folderid) {
		s := fmt.Sprintf("403 Forbidden : you can't access this resource (%s).", folderid)
		http.Error(w, s, 403)
		log.Printf(s)
		return
	}

	fn := fmt.Sprintf("%s%c%s", pf.Path, filepath.Separator, filename)

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		http.NotFound(w, r)
		log.Printf("404 not found: %s", fn)
		return
	}

	file, err := os.Open(fn)
	if err != nil {
		log.Printf("ERROR: Cannot open file: %q", err)
		http.Error(w, err.Error(), 0)
		return
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Printf("ERROR: Cannot stat file: %q", err)
		http.Error(w, err.Error(), 0)
		return
	}

	log.Printf("GET granted to %s \t%s (%s)", aDB.EmailFromHTTPRequest(r), fn, formatSize(fi.Size()))

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024*64)

	w.Header().Set("Content-Type", contentTypeFromExtension(filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
	w.WriteHeader(http.StatusOK)

	writer := bufio.NewWriter(w)
	defer writer.Flush()

	for {
		read, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("ERROR: Error reading file %q", err)
			return
		}

		if read == 0 {
			break
		}

		_, err = writer.Write(buffer[:read])
		if err != nil {
			log.Printf("ERROR: Error writing static file %q", err)
			return
		}
	}
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%d KB", size/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%d MB", size/(1024*1024))
	}
	return fmt.Sprintf("%d GB", size/(1024*1024*1024))
}
