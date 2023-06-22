package models

import (
	"os"
	"path/filepath"
	"time"

	"github.com/neoguojing/gormboot/v2"
)

var (
	basepath = os.Getenv("DB_PATH")
	dbPath   = filepath.Join(basepath, "telegram.db")
	tgDB     = gormboot.New(gormboot.DefaultSqliteConfig(dbPath))
)

// CREATE TABLE telegram_user_info (
//
//	chat_id INTEGER NOT NULL,
//	username VARCHAR(255),
//	first_name VARCHAR(255),
//	last_name VARCHAR(255),
//	phone_number VARCHAR(255),
//	bio TEXT,
//	accesshash VARCHAR(20),
//	is_bot BOOLEAN,
//	image_path VARCHAR(255),
//	created_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	updated_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	PRIMARY KEY (chat_id)
//
// );

type TelegramUserInfo struct {
	ChatID      int64  `gorm:"primary_key"`
	Username    string `gorm:"type:varchar(255)"`
	FirstName   string `gorm:"type:varchar(255)"`
	LastName    string `gorm:"type:varchar(255)"`
	PhoneNumber string `gorm:"type:varchar(255)"`
	Bio         string `gorm:"type:text"`
	AccessHash  string `gorm:"type:varchar(20)"`
	IsBot       bool   `gorm:"type:boolean"`
	ImagePath   string `gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FindByChatIDOrUsername finds TelegramUserInfo by ChatID or Username
func (t *TelegramUserInfo) FindByChatIDOrUsername(chatID int64, username string) (*TelegramUserInfo, error) {
	db := tgDB.DB()
	if chatID == 0 && username != "" {
		if err := db.Where("username = ?", username).First(t).Error; err != nil {
			return nil, err
		}
	} else if chatID != 0 {
		if err := db.Where("chat_id = ?", chatID).First(t).Error; err != nil {
			return nil, err
		}
	}
	return t, nil
}

// FindByChatIDs finds TelegramUserInfo by multiple ChatIDs
func (t *TelegramUserInfo) FindByChatIDs(chatIDs []int64) ([]TelegramUserInfo, error) {
	db := tgDB.DB()
	var userInfos []TelegramUserInfo
	if len(chatIDs) > 0 {
		if err := db.Where("chat_id IN (?)", chatIDs).Find(&userInfos).Error; err != nil {
			return nil, err
		}
	}
	return userInfos, nil
}

// CREATE TABLE telegram_profile (
//
//	chat_id INTEGER NOT NULL,
//	keywords VARCHAR(255),
//	summary TEXT,
//	urls TEXT,
//	names TEXT,
//	location TEXT,
//	count INTEGER,
//	created_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	updated_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	PRIMARY KEY (chat_id)

type TelegramProfile struct {
	ChatID    int64  `gorm:"primary_key"`
	Keywords  string `gorm:"type:varchar(255)"`
	Summary   string `gorm:"type:text"`
	Urls      string `gorm:"type:text"`
	Names     string `gorm:"type:text"`
	Location  string `gorm:"type:text"`
	Count     int    `gorm:"type:integer"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *TelegramProfile) FindByKeywords(keywords []string, limit int, offset int) ([]TelegramProfile, error) {
	db := tgDB.DB()
	var profiles []TelegramProfile
	if len(keywords) > 0 {
		if limit > 20 {
			limit = 20
		} else if limit < 1 {
			limit = 6
		}
		query := "keywords LIKE ?"
		for i := 1; i < len(keywords); i++ {
			query += " OR keywords LIKE ?"
		}
		args := make([]interface{}, len(keywords))
		for i, keyword := range keywords {
			args[i] = "%" + keyword + "%"
		}
		if err := db.Where(query, args...).Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

func (t *TelegramProfile) FindByLocations(locations []string, limit int, offset int) ([]TelegramProfile, error) {
	db := tgDB.DB()
	var profiles []TelegramProfile
	if len(locations) > 0 {
		if limit > 20 {
			limit = 20
		} else if limit < 1 {
			limit = 6
		}
		query := "location LIKE ?"
		for i := 1; i < len(locations); i++ {
			query += " OR location LIKE ?"
		}
		args := make([]interface{}, len(locations))
		for i, location := range locations {
			args[i] = "%" + location + "%"
		}
		if err := db.Where(query, args...).Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

// FindByLocationAndKeyword finds TelegramProfile by location and keyword
func (t *TelegramProfile) FindByLocationAndKeyword(locations []string, keywords []string, limit int, offset int) ([]TelegramProfile, error) {
	db := tgDB.DB()
	var profiles []TelegramProfile
	if limit > 20 {
		limit = 20
	} else if limit < 1 {
		limit = 6
	}

	query := ""
	args := make([]interface{}, len(locations)*len(keywords)*2)
	for i, location := range locations {
		for j, keyword := range keywords {
			if i > 0 || j > 0 {
				query += " OR "
			}
			query += "(location LIKE ? AND keywords LIKE ?)"
			args[(i*len(keywords)+j)*2] = "%" + location + "%"
			args[(i*len(keywords)+j)*2+1] = "%" + keyword + "%"
		}
	}

	if err := db.Where(query, args...).Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// CREATE TABLE telegram_chat (
// 	chat_id INTEGER NOT NULL,
// 	name VARCHAR(100),
// 	chat_type VARCHAR(20),
// 	username VARCHAR(255),
// 	accesshash VARCHAR(20),
// 	created_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
// 	updated_at DATETIME,
// 	PRIMARY KEY (chat_id)
// );

type TelegramChat struct {
	ChatID     int64  `gorm:"primary_key"`
	Name       string `gorm:"type:varchar(100)"`
	ChatType   string `gorm:"type:varchar(20)"`
	Username   string `gorm:"type:varchar(255)"`
	AccessHash string `gorm:"type:varchar(20)"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// FindByChatID finds TelegramChat by ChatID
func (t *TelegramChat) FindByChatID(chatID int64) (*TelegramChat, error) {
	db := tgDB.DB()
	if err := db.Where("chat_id = ?", chatID).First(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

// CREATE TABLE telegram_chat_message (
//
//	id INTEGER NOT NULL,
//	chat_id INTEGER,
//	group_id INTEGER,
//	from_id INTEGER,
//	to_id INTEGER,
//	message TEXT,
//	media_type VARCHAR(20),
//	media_path VARCHAR(255),
//	created_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	updated_at DATETIME DEFAULT (CURRENT_TIMESTAMP),
//	PRIMARY KEY (id)
//
// );
type TelegramChatMessage struct {
	ID        int64  `gorm:"primary_key"`
	ChatID    int64  `gorm:"type:integer"`
	GroupID   int64  `gorm:"type:integer"`
	FromID    int64  `gorm:"type:integer"`
	ToID      int64  `gorm:"type:integer"`
	Message   string `gorm:"type:text"`
	MediaType string `gorm:"type:varchar(20)"`
	MediaPath string `gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
