package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

func (p *PlaygroundServer) getAllCommentsHandler(w http.ResponseWriter, r *http.Request) {
	playgroundID, err := extractParameterFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = encodeToJson(w, p.database.GetAllComments(playgroundID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	playgroundID, err := extractParameterFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	commentID, err := extractParameterFromRequest(r, "commentID")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = encodeToJson(w, p.database.GetComment(playgroundID, commentID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		username := claims.Username
		playgroundID, err := extractParameterFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = r.ParseForm()
		if err != nil {
			log.Println("Couldn't parse request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if content := strings.TrimSpace(r.FormValue("comment")); content != "" {
			newComment := Comment{
				PlaygroundID:     playgroundID,
				Content:          content,
				Author:           username,
				ID:               len(p.database.GetAllComments(playgroundID)) + 1,
				TimeOfSubmission: time.Now(),
			}

			err := p.database.AddComment(playgroundID, newComment)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		playgroundID, err := extractParameterFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		commentID, err := extractParameterFromRequest(r, "commentID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		comment := p.database.GetComment(playgroundID, commentID)
		if comment.Author != claims.Username {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = p.database.DeleteComment(playgroundID, commentID)
		if err != nil {
			log.Printf("Impossible de supprimer le commentaire, %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) modifyCommentHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		playgroundID, err := extractParameterFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		commentID, err := extractParameterFromRequest(r, "commentID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		updatedComment, err := NewCommentFromJSON(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		updatedComment.Content = strings.TrimSpace(updatedComment.Content)
		updatedComment.TimeOfSubmission = time.Now()
		updatedComment.ID = commentID
		updatedComment.Author = claims.Username
		// func (p *Playground) UpdateComment(updatedComment Comment) error {
		// 	for index, comment := range p.Comments {
		// 		if comment.ID == updatedComment.ID {
		// 			if comment.Author != updatedComment.Author {
		// 				return errors.New("Not the same author")
		// 			}
		// 			content := strings.TrimSpace(updatedComment.Content)
		// 			if content == "" {
		// 				return ErrEmptyField
		// 			}
		// 			p.Comments[index].Content = content
		// 			p.Comments[index].TimeOfSubmission = updatedComment.TimeOfSubmission
		// 			return nil
		// 		}
		// 	}
		// 	return errors.New("Couldn't find comment")
		// }
		err = p.database.ModifyComment(playgroundID, commentID, updatedComment)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func NewCommentFromJSON(input io.Reader) (Comment, error) {
	var comment Comment
	err := json.NewDecoder(input).Decode(&comment)
	if err != nil {
		return Comment{}, fmt.Errorf("Unable to parse input %q into slice, '%v'", input, err)
	}
	return comment, nil
}
