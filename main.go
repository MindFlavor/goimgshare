// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/mindflavor/goimgshare/authdb"
	"github.com/mindflavor/goimgshare/config"
	"github.com/mindflavor/goimgshare/folders/physical"
	"github.com/mindflavor/goimgshare/thumb"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
)

const (
	goimgshareConfigPathEnv = "GOIMGSHARECONF"
)

var aDB authdb.DB
var phyFolders physical.Folders
var smallThumbCache, avgThumbCache *thumb.Cache
var conf *config.Config

func main() {
	var configFileName string
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	} else {
		configFileName = os.Getenv(goimgshareConfigPathEnv)
	}

	if configFileName == "" {
		_, file := filepath.Split(os.Args[0])
		log.Printf("Syntax error. Must specify the configuration file either as ")
		log.Printf("environmental variable (%s) or as first command argument.", goimgshareConfigPathEnv)
		log.Fatalf("%s program exiting.", file)
		return
	}

	log.Printf("Opening %s", configFileName)

	file, err := os.Open(configFileName)
	if err != nil {
		panic(fmt.Sprintf("Cannot open configuration file: %s ", err))
	}
	defer file.Close()

	conf, err = config.Load(file)
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration file: %s ", err))
	}

	gomniauth.SetSecurityKey(signature.RandomKey(64))

	aDB = authdb.New()

	smallThumbCache = thumb.New(conf.ThumbnailCacheFolder, conf.SmallThumbnailSize, conf.SmallThumbnailSize)
	avgThumbCache = thumb.New(conf.ThumbnailCacheFolder, conf.AverageThumbnailSize, conf.AverageThumbnailSize)

	// load folders
	fmt.Printf("conf.SharedFoldersConfigurationFile == %s", conf.SharedFoldersConfigurationFile)
	file, err = os.Open(conf.SharedFoldersConfigurationFile)
	if err != nil {
		panic(fmt.Sprintf("Cannot open shared folder configuration file: %s ", err))
	}
	defer file.Close()
	phyFolders, err = physical.Load(file)
	if err != nil {
		panic(err)
	}
	// end load folders

	var prov []common.Provider
	var providers []string

	if conf.Google != nil {
		providers = append(providers, "google")
		prov = append(prov, google.New(conf.Google.ClientID, conf.Google.Secret, conf.Google.ReturnURL))
	}
	if conf.Facebook != nil {
		providers = append(providers, "facebook")
		prov = append(prov, facebook.New(conf.Facebook.ClientID, conf.Facebook.Secret, conf.Facebook.ReturnURL))
	}
	if conf.Github != nil {
		providers = append(providers, "github")
		prov = append(prov, github.New(conf.Github.ClientID, conf.Github.Secret, conf.Github.ReturnURL))
	}

	gomniauth.WithProviders(prov...)

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

	// handle thumbnails
	router.HandleFunc("/smallthumb/{folderid}/{fn}", logHandler(requireAuth(handleThumbnail)))
	router.HandleFunc("/avgthumb/{folderid}/{fn}", logHandler(requireAuth(handleThumbnail)))

	// register all the static content with NO authentication
	files, err := ioutil.ReadDir(filepath.Join(conf.InternalHTTPFilesPath, staticDirectory, staticDirectoryNoAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := path.Join("/", staticDirectory, file.Name())
		log.Printf("Registering %s with noauth", path)

		if conf.LogInternalHTTPFilesAccess {
			router.HandleFunc(path, logHandler(handleStatic(staticDirectoryNoAuth, file.Name())))
		} else {
			router.HandleFunc(path, handleStatic(staticDirectoryNoAuth, file.Name()))
		}
	}

	// register all the static content with authentication
	files, err = ioutil.ReadDir(filepath.Join(conf.InternalHTTPFilesPath, staticDirectory, staticDirectoryAuth))
	if err != nil {
		panic(fmt.Sprintf("Cannot access static content folder: %s", err))
	}

	for _, file := range files {
		path := path.Join("/", staticDirectory, file.Name())
		log.Printf("Registering %s with auth", path)

		if conf.LogInternalHTTPFilesAccess {
			router.HandleFunc(path, logHandler(requireAuth(handleStatic(staticDirectoryAuth, file.Name()))))
		} else {
			router.HandleFunc(path, requireAuth(handleStatic(staticDirectoryAuth, file.Name())))
		}
	}

	http.HandleFunc("/", logHandler(requireAuth(handleStatic(staticDirectoryAuth, "index.html"))))

	for _, provider := range providers {
		router.HandleFunc(fmt.Sprintf("/auth/%s/login", provider), loginHandler(provider))
		router.HandleFunc(fmt.Sprintf("/auth/%s/callback", provider), callbackHandler(provider))
	}

	router.HandleFunc("/supportedAuths", logHandler(handleSupportedAuths))

	if conf.HttpsCertificateFile != "" && conf.HttpsCertificateKeyFile != "" {
		log.Printf("Starting encrypted TLS webserver on port %d...", conf.Port)
		if err := http.ListenAndServeTLS(fmt.Sprintf(":%d", conf.Port), conf.HttpsCertificateFile, conf.HttpsCertificateKeyFile, router); err != nil {
			log.Fatalf("ERROR starting webserver: %s", err)
		}
	} else {
		log.Printf("Starting non encrypted webserver on port %d...", conf.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router); err != nil {
			log.Fatalf("ERROR starting webserver: %s", err)
		}
	}

	//	http.ListenAndServeTLS()
}

func loginHandler(providerName string) http.HandlerFunc {
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		state := gomniauth.NewState("after", "success")

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

func generateSampleFolderFile() {
	ps := physical.New()
	ps["001"] = physical.Folder{Path: "C:\\temp", AuthorizedMails: map[string]bool{"francesco.cogno@gmail.com": true, "valentina.campora@gmail.com": true}}
	ps["001"].ID = "001"
	ps["001"].Name = "temp"

	ps["002"] = physical.Folder{Path: "D:\\temp\\pic", AuthorizedMails: map[string]bool{"francesco.cogno@gmail.com": true, "prova@test.com": true}}
	ps["002"].ID = "002"
	ps["002"].Name = "pic"

	file, err := os.Create("/home/MINDFLAVOR/mindflavor/shared_folders.json")
	if err != nil {
		panic(err)
	}

	ps.Save(file)
	file.Close()
}

func generateSampleConfigurationFile() {
	config := config.Config{
		Port: 8080,
		InternalHTTPFilesPath:          "/home/MINDFLAVOR/mindflavor/go/src/github.com/goimgshare",
		SharedFoldersConfigurationFile: "/home/MINDFLAVOR/mindflavor/shared_folders.json",
		ThumbnailCacheFolder:           "/home/MINDFLAVOR/mindflavor/thumbs",
		CacheInternalHTTPFiles:         false,
		LogInternalHTTPFilesAccess:     true,
		SmallThumbnailSize:             500,
		AverageThumbnailSize:           1000,
		Google: &config.AuthProvider{
			ClientID:  "1076712416430-kst5ildq694fa4ntin0t4f432pnfitcp.apps.googleusercontent.com",
			Secret:    "FFEq-52VNT4NAUybziEs09vd",
			ReturnURL: "http://localhost:8080/auth/google/callback",
		},
	}

	file, err := os.Create("/home/MINDFLAVOR/mindflavor/config.json")
	if err != nil {
		panic(err)
	}

	defer file.Close()
	config.Save(file)
}

func logHandler(inner func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := aDB.EmailFromHTTPRequest(r)

		start := time.Now()

		inner(w, r)

		log.Printf("%s\t\t%s\t%s\t\t\t\t\t\t%s",
			auth,
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}
