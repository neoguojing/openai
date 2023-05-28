package models

import (
	"log"

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

type ChatRecord struct {
	gorm.Model
	Request   string `gorm:"uniqueIndex"`
	Reply     string
	MediaType MediaType
	FilePath  string
}

func (o *ChatRecord) CreateChatRecord(request string, reply string, mediaType MediaType) error {
	chatRecord := &ChatRecord{
		Request:   request,
		Reply:     reply,
		MediaType: mediaType,
	}
	if err := db.Create(chatRecord).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (o *ChatRecord) GetChatRecord(id uint) (*ChatRecord, error) {
	chatRecord := &ChatRecord{}
	if err := db.First(chatRecord, id).Error; err != nil {
		log.Println(err)
		return nil, err
	}
	return chatRecord, nil
}

func (o *ChatRecord) UpdateChatRecord(request string, reply string,
	mediaType MediaType) error {
	chatRecord := &ChatRecord{}
	if err := db.First(chatRecord, request).Error; err != nil {
		log.Println(err)
		return err
	}
	chatRecord.Request = request
	chatRecord.Reply = reply
	chatRecord.MediaType = mediaType
	if err := db.Save(chatRecord).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (o *ChatRecord) DeleteChatRecord(id uint) error {
	chatRecord := &ChatRecord{}
	if err := db.First(chatRecord, id).Error; err != nil {
		log.Println(err)
		return err
	}
	if err := db.Delete(chatRecord).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}
