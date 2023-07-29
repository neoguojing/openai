package session

import (
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
	"net/http"
	"github.com/neoguojing/log"
)

const (
	SESSION_OFFICE_ACCOUNT = "office-account"
)


type Session struct {
	sessionStore *gormstore.Store 
}

func NewSession(db *gorm.DB,secret string)*Session{
	return &Session{
		sessionStore : gormstore.New(db, []byte(secret)),
	}
}

func (s *Session)OfficeaccountHandler(w http.ResponseWriter, r *http.Request,openid string) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := s.sessionStore.Get(r, openid)
	// Set some session values.
	if  session.Values["count"] == nil {
		session.Values["count"] = 1
	} else {
		count := session.Values["count"].(int)
		session.Values["count"] = count+1
	}
	
	log.Infof("OfficeaccountHandler session:%v",session)
	// Save it before we write to the response/return from the handler.
	err := s.sessionStore.Save(r, w,session)
	if err != nil {
		log.Error(err.Error())
		return
	}
}