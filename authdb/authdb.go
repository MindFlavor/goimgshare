package authdb

import (
	"github.com/stretchr/signature"
	"time"
)

type Signature string

type AuthToken struct {
	Sig        Signature
	Expiration time.Time
}

type DB map[Signature]AuthToken

func New() DB {
	return make(map[Signature]AuthToken)
}

func (db DB) Register(s string, expiration time.Time) AuthToken {
	sig := Signature(signature.RandomKey(64))
	db[sig] = AuthToken{sig, expiration}
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
