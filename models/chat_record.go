package models

import (
	"github.com/neoguojing/log"

	"gorm.io/gorm"
)

type MediaType string

const (
	Voice   MediaType = "voice"
	Picture MediaType = "picture"
	Text    MediaType = "text"
	Video   MediaType = "video"
	File    MediaType = "file"
)

func (m MediaType) IsValid() bool {
	switch m {
	case Voice, Picture, Text, Video, File:
		return true
	default:
		return false
	}
}

type Platform int

const (
	Wechat     Platform = 1
	Telegram   Platform = 2
	HttpServer Platform = 3
	Chatbot    Platform = 4
)

type ChatRecord struct {
	gorm.Model
	Request   string `gorm:"uniqueIndex"`
	Reply     string
	MediaType MediaType
	FilePath  string
	Platform  Platform
	Frequency int
}

func (o *ChatRecord) CreateChatRecord() error {

	if err := db.Where("request = ?", o.Request).First(o).Error; err == nil {
		o.Frequency += 1
		if err := db.Save(o).Error; err != nil {
			log.Error(err.Error())
			return err
		}
	} else if err := db.Create(o).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (o *ChatRecord) GetChatRecord(id uint) (*ChatRecord, error) {
	chatRecord := &ChatRecord{}
	if err := db.First(chatRecord, id).Error; err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return chatRecord, nil
}

func (o *ChatRecord) UpdateChatRecord(request string, reply string,
	mediaType MediaType) error {
	chatRecord := &ChatRecord{}
	if err := db.Where("request = ?", request).First(chatRecord).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	chatRecord.Request = request
	chatRecord.Reply = reply
	chatRecord.MediaType = mediaType
	if err := db.Save(chatRecord).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (o *ChatRecord) DeleteChatRecord(id uint) error {
	chatRecord := &ChatRecord{}
	if err := db.First(chatRecord, id).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	if err := db.Delete(chatRecord).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (o *ChatRecord) GetChatRecordsByFrequency(offset int, limit int) ([]*ChatRecord, error) {
	chatRecords := []*ChatRecord{}
	if err := db.Order("frequency desc").Offset(offset).Limit(limit).Find(&chatRecords).Error; err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return chatRecords, nil
}

func (o *ChatRecord) UpdateFrequency(request string, frequency int) error {
	chatRecord := &ChatRecord{}
	if err := db.Where("request = ?", request).First(chatRecord).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	chatRecord.Frequency += frequency
	if err := db.Save(chatRecord).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
