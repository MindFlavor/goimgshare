package authdb

import (
	"fmt"
	"github.com/stretchr/signature"
	"log"
	"net/http"
	"time"

	"github.com/mindflavor/goimgshare/folders/physical"
)

type Signature string

type AuthToken struct {
	Sig        Signature
	Expiration time.Time
	email      string
}

type DB map[Signature]AuthToken

func New() DB {
	return make(map[Signature]AuthToken)
}

func (db DB) Register(s string, email string, expiration time.Time) AuthToken {
	sig := Signature(signature.RandomKey(64))
	db[sig] = AuthToken{sig, expiration, email}
	return db[sig]
}

func (db DB) IsRegistered(sig Signature) bool {
	val, found := db[sig]

	if found {
		if time.Now().After(val.Expiration) {
			// remove from map
			delete(db, sig)
			return false
		}

		return true
	}

	return false
}

func (db DB) EmailFromHTTPRequest(r *http.Request) string {
	auth := "unauthenticated"

	cookie, err := r.Cookie("auth")
	if err == nil {
		auth, _ = db.Email(Signature(cookie.Value))
	}

	return auth
}

func (db DB) Email(sig Signature) (string, error) {
	val, found := db[sig]

	if found {
		return val.email, nil
	}

	return "forged identity", fmt.Errorf("Invalid signature %s", sig)
}

func (db DB) IsAuthorized(phyFolders *physical.Folders, r *http.Request, folderID string) bool {
	cookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Authentication not present")
		return false
	}
	if !db.IsRegistered(Signature(cookie.Value)) {
		log.Printf("Authentication not valid or expired")
		return false
	}

	val := db[Signature(cookie.Value)]

	return phyFolders.IsAuthorized(folderID, val.email)
}
