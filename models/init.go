package models

import (
	"github.com/neoguojing/log"

	"github.com/neoguojing/gormboot/v2"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	recoder *Recorder
)

func init() {
	gormboot.DefaultDB.RegisterModel(&Role{}, &ChatRecord{})
	db = gormboot.DefaultDB.AutoMigrate().DB()
	recoder = NewRecorder()
	log.Infof("telegram db path：%s", tgDBPath)
}

type Operation string

const (
	Create Operation = "create"
	Update Operation = "update"
)

type Element struct {
	Operation  Operation
	ChatRecord *ChatRecord
}

type Recorder struct {
	syncer chan Element
}

func NewRecorder() *Recorder {
	r := &Recorder{
		syncer: make(chan Element, 100),
	}
	go r.loop()
	return r
}

func (r *Recorder) loop() {
	for {
		select {
		case elem, ok := <-r.syncer:
			if !ok {
				return
			}
			switch elem.Operation {
			case Create:
				err := elem.ChatRecord.CreateChatRecord()
				if err != nil {
					log.Error(err.Error())
				}
			case Update:
				err := elem.ChatRecord.UpdateFrequency(elem.ChatRecord.Request, elem.ChatRecord.Frequency)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

}

func (r *Recorder) Exit() {
	close(r.syncer)
}

func (r *Recorder) Send(elem Element) {
	r.syncer <- elem
}

func GetRecorder() *Recorder {
	return recoder
}
