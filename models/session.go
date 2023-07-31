package models

import (
	"encoding/json"
	"time"

	"github.com/neoguojing/log"
	"github.com/neoguojing/openai/utils"
	"gorm.io/gorm"
)

type Session struct {
	ID        string                 `sql:"unique_index"`
	Data      string                 `sql:"type:text"`
	Values    map[string]interface{} `gorm:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time `sql:"index"`
}

func (m *Session) TableName() string {
	return "sessions"
}

func (s *Session) SetSession(id string, k string, v interface{}, duration time.Duration) {
	s.ID = id
	s.Values[k] = v
	s.ExpiresAt = time.Now().Add(duration)
}

func lruCallback(key string, item *utils.CacheItem) {
	if item == nil {
		return
	}

	if item.Value == nil {
		return
	}

	session := item.Value.(*Session)
	data, _ := json.Marshal(session.Values)
	session.Data = string(data)
	session.UpdatedAt = time.Now()
	log.Infof("save the session %v", *session)
	db.Where("id = ?", key).Updates(map[string]interface{}{"data": session.Data, "updated_at": session.UpdatedAt})
}

type SessionManager struct {
	sessionCache  *utils.LRUCache
	cleanInterval time.Duration
	exitChan      chan struct{}
}

func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessionCache:  utils.NewLRUCache(100, lruCallback),
		cleanInterval: time.Minute,
	}

	go m.job()
	return m
}

// GetSession 从缓存中获取Session,缓存未命中从数据库加载
func (m *SessionManager) GetSession(id string) *Session {
	if s := m.sessionCache.Get(id); s != nil {
		return s.(*Session)
	}

	var s Session
	s.ID = id
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Values = map[string]interface{}{}
	err := db.Model(&Session{}).Where("id = ?", id).First(&s).Error
	if err == nil {
		json.Unmarshal([]byte(s.Data), &s.Values)
	} else if err == gorm.ErrRecordNotFound {
		db.Save(&s)
	}

	m.sessionCache.Set(id, &s, 0)
	return &s
}

// DeleteSession 从数据库和缓存中删除
func (m *SessionManager) DeleteSession(id string) error {
	err := db.Model(&Session{}).Where("id = ?", id).Delete(&Session{}).Error
	if err == nil {
		m.sessionCache.Delete(id)
	}
	return err
}

// DeleteSession 从数据库和缓存中删除
func (m *SessionManager) UpdateSession(id, data string) error {
	err := db.Model(&Session{}).Where("id = ?", id).UpdateColumn("data", data).Error
	if err != nil {
		return err
	}
	return err
}

func (m *SessionManager) job() {
	for {

		select {
		case <-m.exitChan:
			return
		default:
		}

		var sessions []Session
		db.Model(&Session{}).Where("expires_at is not null and expires_at < ?", time.Now()).Find(&sessions)

		// 删除已过期的Session
		for _, s := range sessions {
			m.DeleteSession(s.ID)
		}

		// 睡眠一段时间后继续下一轮清理
		time.Sleep(m.cleanInterval)
	}
}

func (m *SessionManager) Close() {
	close(m.exitChan)
}
