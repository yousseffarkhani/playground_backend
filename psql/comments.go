package psql

import (
	"fmt"
	"log"

	"github.com/yousseffarkhani/playground/backend2/server"
)

func (db *playgroundDatabase) GetComment(PlaygroundID, commentID int) server.Comment {
	var comment server.Comment
	db.Where("playground_id=? AND id=?", PlaygroundID, commentID).Find(&comment)
	return comment
}

func (db *playgroundDatabase) GetAllComments(PlaygroundID int) server.Comments {
	var comments server.Comments
	db.Where("playground_id=?", PlaygroundID).Find(&comments)
	return comments
}

func (db *playgroundDatabase) AddComment(playgroundID int, newComment server.Comment) error {
	if err := db.Create(&newComment).Error; err != nil {
		return fmt.Errorf("Couldn't add comment, %s", err)
	}
	log.Println("New comment added successfully")
	return nil
}

func (db *playgroundDatabase) DeleteComment(playgroundID, commentID int) error {
	if err := db.Unscoped().Where("playground_id=? AND id=?", playgroundID, commentID).Delete(&server.Comment{}).Error; err != nil {
		return fmt.Errorf("Couldn't delete comment, %s", err)
	}
	log.Println("Comment successfully deleted")
	return nil
}

func (db *playgroundDatabase) ModifyComment(updatedComment server.Comment) error {
	if err := db.Save(&updatedComment).Error; err != nil {
		return fmt.Errorf("Couldn't update comment, %s", err)
	}
	log.Println("Comment successfully updated")
	return nil
}

func (db *playgroundDatabase) GetLastCommentID() int {
	var max float64
	row := db.Table("comments").Select("max(id)").Row()
	row.Scan(&max)
	return int(max)
}
