package psql

import (
	"fmt"
	"log"

	"github.com/yousseffarkhani/playground/backend2/server"
)

func (db *playgroundDatabase) GetAllSubmittedPlaygrounds() server.Playgrounds {
	var playgrounds server.Playgrounds
	db.Where("draft=?", true).Find(&playgrounds)
	return playgrounds
}

func (db *playgroundDatabase) GetSubmittedPlayground(ID int) server.Playground {
	var playground server.Playground
	db.Where("id=? AND draft=?", ID, true).Find(&playground)
	return playground
}

func (db *playgroundDatabase) SubmitPlayground(newPlayground server.Playground) {
	db.Create(&newPlayground)
	log.Println("New playground successfully submitted")
}

func (db *playgroundDatabase) DeleteSubmittedPlayground(ID int) error {
	if err := db.Unscoped().Where("id=? AND draft=?", ID, true).Delete(&server.Playground{}).Error; err != nil {
		return fmt.Errorf("Couldn't delete submitted playground, %s", err)
	}
	log.Println("Submitted playground successfully deleted")
	return nil
}
