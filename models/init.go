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
}

type Recorder struct {
	syncer chan ChatRecord
}

func NewRecorder() *Recorder {
	r := &Recorder{
		syncer: make(chan ChatRecord, 100),
	}
	go r.loop()
	return r
}

func (r *Recorder) loop() {
	for {
		select {
		case record := <-r.syncer:
			err := record.CreateChatRecord()
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (r *Recorder) Exit() {
	close(r.syncer)
}

func (r *Recorder) Send(record ChatRecord) {
	r.syncer <- record
}

func GetRecorder() *Recorder {
	return recoder
}
