package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func handleThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fn := filepath.Join(pf.Path, filename)

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		http.NotFound(w, r)
		log.Printf("404 not found: %s", fn)
		return
	}

	reader, err := smallThumbCache.GetThumb(&pf, filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	buf := make([]byte, 1024*32)
	for {
		iread, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), 500)
				return
			}
			return
		}
		_, err = w.Write(buf[:iread])
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
