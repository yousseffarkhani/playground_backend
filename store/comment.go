package store

type Comment struct {
	ID int
	// PlaygroundID int
	Content string
	Author  string
	// Date         string // TODO: Ajouter
}

type Comments []Comment

// func (c Comments) Find(playgroundID int) Comments {
// 	var comments Comments
// 	for _, comment := range c {
// 		if comment.PlaygroundID == playgroundID {
// 			comments = append(comments, comment)
// 		}
// 	}
// 	return comments
// }
