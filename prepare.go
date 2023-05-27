package openai

import (
	"log"

	"github.com/neoguojing/openai/role"
)

func (o *Chat) Prepare(roleName string) *Chat {
	roles, err := role.SearchRoleByName(roleName)
	if err != nil {
		log.Panicln(err)
		return nil
	}

	if len(roles) == 0 {
		log.Panicln("roles was empty")
		return nil
	}
	chatResponse, err := o.Complete(roles[0].Desc)
	if err != nil {
		log.Panicln(err)
		return nil
	}
	log.Println(chatResponse.Choices[0].Message.Content)
	return o
}
