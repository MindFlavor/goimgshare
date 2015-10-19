package authdb

import (
	"fmt"
	"github.com/stretchr/signature"
	"log"
	"net/http"
	"time"

	"github.com/mindflavor/goimgshare/folders/physical"
)

// Signature is the
// opaque mail signature
type Signature string

// AuthToken contains the
// runtime association between
// Signature and email
type AuthToken struct {
	Sig        Signature
	Expiration time.Time
	email      string
}

// DB is the in-memory
// store of valid emails
// signatures
type DB map[Signature]AuthToken

// New creates a new authdb.DB
func New() DB {
	return make(map[Signature]AuthToken)
}

// Register adds a new email
// to the DB
func (db DB) Register(s string, email string, expiration time.Time) AuthToken {
	sig := Signature(signature.RandomKey(64))
	db[sig] = AuthToken{sig, expiration, email}
	return db[sig]
}

// IsRegistered is true if the
// signature is registered in the
// DB
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

// EmailFromHTTPRequest translates the
// cookie from the http.Request in the
// authenticated email (if present and valid)
func (db DB) EmailFromHTTPRequest(r *http.Request) string {
	auth := "unauthenticated"

	cookie, err := r.Cookie("auth")
	if err == nil {
		auth, _ = db.Email(Signature(cookie.Value))
	}

	return auth
}

// Email extracts the email
// linked to the Signature passed as parameter
// (if present)
func (db DB) Email(sig Signature) (string, error) {
	val, found := db[sig]

	if found {
		return val.email, nil
	}

	return "forged identity", fmt.Errorf("Invalid signature %s", sig)
}

// IsAuthorized returns true if the
// specified user (extracted from the http.Request)
// can rightfully access the folderID in physical.Folders
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
