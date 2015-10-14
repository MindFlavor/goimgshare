// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/mindflavor/testauth/authdb"
	"github.com/mindflavor/testauth/logical"
	"github.com/mindflavor/testauth/physical"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
)

var aDB authdb.DB

func main() {
	gomniauth.SetSecurityKey(signature.RandomKey(64))

	aDB = authdb.New()

	gomniauth.WithProviders(
		google.New("1076712416430-kst5ildq694fa4ntin0t4f432pnfitcp.apps.googleusercontent.com", "FFEq-52VNT4NAUybziEs09vd", "http://localhost:8080/auth/google/callback"),
		github.New("3d1e6ba69036e0624b61", "7e8938928d802e7582908a5eadaaaf22d64babf1", "http://localhost:8080/auth/github/callback"),
		facebook.New("537611606322077", "f9f4d77b3d3f4f5775369f5c9f88f65e", "http://localhost:8080/auth/facebook/callback"),
	)

	http.HandleFunc("/folders", logHandler(requireAuth(handleFolders), "folders"))

	// register all the static content with NO authentication
	files, err := ioutil.ReadDir(path.Join(staticDirectory, staticDirectoryNoAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := fmt.Sprintf("/%s/%s", staticDirectory, file.Name())
		log.Printf("Registering %s with noauth", path)
		http.HandleFunc(path, logHandler(handleStatic(staticDirectoryNoAuth, file.Name()), file.Name()))
	}

	// register all the static content with authentication
	files, err = ioutil.ReadDir(path.Join(staticDirectory, staticDirectoryAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := fmt.Sprintf("/%s/%s", staticDirectory, file.Name())
		log.Printf("Registering %s with auth", path)
		http.HandleFunc(path, logHandler(requireAuth(handleStatic(staticDirectoryAuth, file.Name())), file.Name()))
	}

	http.HandleFunc("/", logHandler(requireAuth(handleStatic(staticDirectoryAuth, "index.html")), "/"))

	providers := []string{"google", "github", "facebook"}
	for _, provider := range providers {
		http.HandleFunc(fmt.Sprintf("/auth/%s/login", provider), loginHandler(provider))
		http.HandleFunc(fmt.Sprintf("/auth/%s/callback", provider), callbackHandler(provider))
	}

	log.Printf("Server running...")
	http.ListenAndServe(":8080", nil)
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

		sig := aDB.Register(user.Email(), time.Now().AddDate(0, 0, 1))
		cookie := http.Cookie{Name: "auth", Value: string(sig.Sig), Expires: sig.Expiration, Path: "/"}
		http.SetCookie(w, &cookie)

		//		data := fmt.Sprintf("%#v", user)
		//		io.WriteString(w, data)

		// redirect
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func generateSampleConfigurationFile() {
	var ps physical.Folders
	ps = append(ps, &physical.Folder{&logical.Folder{ID: "001", Name: "C:\\temp"}, "C:\\temp"})
	ps = append(ps, &physical.Folder{&logical.Folder{ID: "002", Name: "D:\\temp\\pic"}, "D:\\temp\\pic"})

	file, err := os.Create("C:\\temp\\config.json")
	if err != nil {
		panic(err)
	}

	ps.Save(file)
	file.Close()
}

func logHandler(inner func(w http.ResponseWriter, r *http.Request), name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner(w, r)

		log.Printf("%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	}
}
