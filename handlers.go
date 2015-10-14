package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/mindflavor/testauth/authdb"
	"github.com/mindflavor/testauth/physical"
)

const staticDirectory = "html"
const staticDirectoryAuth = "auth"
const staticDirectoryNoAuth = "noauth"

var staticCache map[string][]byte

func init() {
	staticCache = make(map[string][]byte)
}

func cors(exec func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		exec(w, r)
	}
}

func requireAuth(exec func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			log.Printf("Authentication required, redirecting to auth page")
			http.Redirect(w, r, "/html/auth.html", http.StatusFound)
			return
		}

		// check if is in role
		if !aDB.IsRegistered(authdb.Signature(cookie.Value)) {
			log.Printf("Authentication not valid or expired, redirecting to auth page")
			http.Redirect(w, r, "/html/auth.html", http.StatusFound)
			return
		}

		exec(w, r)
	}
}

func redirectHandler(redirectTo string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Redirecting from %s to %s", r.URL, redirectTo)
		http.Redirect(w, r, redirectTo, http.StatusMovedPermanently)
	}
}

func handleFolders(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("C:\\temp\\config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	phyFolders, err := physical.Load(file)
	if err != nil {
		log.Printf("Error %s", err)
		jsonifyError(w, err)
		return
	}

	logFolders := phyFolders.ToLogical()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(logFolders)
}

func handleStatic(folder, page string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			if r.URL.Path != "/"+staticDirectory+"/"+page {
				http.NotFound(w, r)
				log.Printf("404 Not found: %s", r.URL)
				return
			}
		}

		localPath := path.Join(staticDirectory, folder, page)

		//		if val, ok := staticCache[path]; ok {
		//			// cache hit
		//			w.Write(val)
		//			return
		//		}

		// cache miss
		f, err := os.Open(localPath)
		if err != nil {
			log.Printf("Err: cannot open %s: %s", localPath, err)
			http.Error(w, err.Error(), 100)
			return
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(f)
		if err != nil {
			panic("Cannot read auth file")
		}

		// check if css so we change the content type
		if strings.ToLower(path.Ext(page)) == ".css" {
			log.Printf("Setting content type to text/css")
			w.Header().Set("Content-Type", "text/css")
		}

		w.Write(buf.Bytes())

		// store in cache
		staticCache[localPath] = buf.Bytes()

	}
}

func jsonifyError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(err)
}
