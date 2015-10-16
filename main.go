// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/mindflavor/goimgshare/authdb"
	"github.com/mindflavor/goimgshare/folders/logical"
	"github.com/mindflavor/goimgshare/folders/physical"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
)

var aDB authdb.DB
var phyFolders physical.Folders

func main() {
	//	generateSampleConfigurationFile()
	//	return

	gomniauth.SetSecurityKey(signature.RandomKey(64))

	aDB = authdb.New()

	// load folders
	file, err := os.Open("C:\\temp\\config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	phyFolders, err = physical.Load(file)
	if err != nil {
		panic(err)
	}
	// end load folders

	gomniauth.WithProviders(
		google.New("1076712416430-kst5ildq694fa4ntin0t4f432pnfitcp.apps.googleusercontent.com", "FFEq-52VNT4NAUybziEs09vd", "http://localhost:8080/auth/google/callback"),
		github.New("3d1e6ba69036e0624b61", "7e8938928d802e7582908a5eadaaaf22d64babf1", "http://localhost:8080/auth/github/callback"),
		facebook.New("537611606322077", "f9f4d77b3d3f4f5775369f5c9f88f65e", "http://localhost:8080/auth/facebook/callback"),
	)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", logHandler(requireAuth(handleStatic(staticDirectoryAuth, "index.html"))))
	router.HandleFunc("/folders", logHandler(requireAuth(handleFolders)))

	// images
	rImages := regexp.MustCompile("(.*[.]jpg$)|(.*[.]gif$)|(.*[.]png$)")
	router.HandleFunc("/images/{folderid}", logHandler(requireAuth(handleFolder(
		func(strtomatch string) bool {
			return rImages.MatchString(strings.ToLower(strtomatch))
		}))))

	//videos
	rVideos := regexp.MustCompile("(.*[.]mp4$)|(.*[.]mkv$)|(.*[.]avi$)")
	router.HandleFunc("/videos/{folderid}", logHandler(requireAuth(handleFolder(
		func(strtomatch string) bool {
			return rVideos.MatchString(strings.ToLower(strtomatch))
		}))))

	// everyting else
	router.HandleFunc("/extra/{folderid}", logHandler(requireAuth(handleFolder(
		func(strtomatch string) bool {
			return !(rVideos.MatchString(strings.ToLower(strtomatch)) || rImages.MatchString(strings.ToLower(strtomatch)))
		}))))

	// serve the files
	router.HandleFunc("/file/{folderid}/{fn}", logHandler(requireAuth(handleStaticFile)))

	// register all the static content with NO authentication
	files, err := ioutil.ReadDir(path.Join(staticDirectory, staticDirectoryNoAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := fmt.Sprintf("/%s/%s", staticDirectory, file.Name())
		log.Printf("Registering %s with noauth", path)
		router.HandleFunc(path, logHandler(handleStatic(staticDirectoryNoAuth, file.Name())))
	}

	// register all the static content with authentication
	files, err = ioutil.ReadDir(path.Join(staticDirectory, staticDirectoryAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := fmt.Sprintf("/%s/%s", staticDirectory, file.Name())
		log.Printf("Registering %s with auth", path)
		router.HandleFunc(path, logHandler(requireAuth(handleStatic(staticDirectoryAuth, file.Name()))))
	}

	http.HandleFunc("/", logHandler(requireAuth(handleStatic(staticDirectoryAuth, "index.html"))))

	providers := []string{"google", "github", "facebook"}
	for _, provider := range providers {
		router.HandleFunc(fmt.Sprintf("/auth/%s/login", provider), loginHandler(provider))
		router.HandleFunc(fmt.Sprintf("/auth/%s/callback", provider), callbackHandler(provider))
	}

	log.Printf("Server running...")
	http.ListenAndServe(":8080", router)
}

func loginHandler(providerName string) http.HandlerFunc {
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		state := gomniauth.NewState("after", "success")

		// This code borrowed from goweb example and not fixed.
		// if you want to request additional scopes from the provider,
		// pass them as login?scope=scope1,scope2
		//options := objx.MSI("scope", ctx.QueryValue("scope"))

		authURL, err := provider.GetBeginAuthURL(state, nil)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func callbackHandler(providerName string) http.HandlerFunc {
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		omap, err := objx.FromURLQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		creds, err := provider.CompleteAuth(omap)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			// This code borrowed from goweb example and not fixed.
			// get the state
			state, err := gomniauth.StateFromParam(ctx.QueryValue("state"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// redirect to the 'after' URL
			afterUrl := state.GetStringOrDefault("after", "error?e=No after parameter was set in the state")
		*/

		// load the user
		user, userErr := provider.GetUser(creds)

		if userErr != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Authenticated as %s", user.Email())

		sig := aDB.Register(user.Email(), user.Email(), time.Now().AddDate(0, 0, 1))
		cookie := http.Cookie{Name: "auth", Value: string(sig.Sig), Expires: sig.Expiration, Path: "/"}
		http.SetCookie(w, &cookie)

		//		data := fmt.Sprintf("%#v", user)
		//		io.WriteString(w, data)

		// redirect
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func generateSampleConfigurationFile() {
	ps := physical.New()
	ps["001"] = physical.Folder{&logical.Folder{ID: "001", Name: "C:\\temp"}, "C:\\temp", map[string]bool{"francesco.cogno@gmail.com": true, "valentina.campora@gmail.com": true}}
	ps["002"] = physical.Folder{&logical.Folder{ID: "002", Name: "D:\\temp\\pic"}, "D:\\temp\\pic", map[string]bool{"francesco.cogno@gmail.com": true, "prova@test.com": true}}

	file, err := os.Create("C:\\temp\\config.json")
	if err != nil {
		panic(err)
	}

	ps.Save(file)
	file.Close()
}

func logHandler(inner func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner(w, r)

		log.Printf("%s\t%s\t\t\t\t\t\t%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}
