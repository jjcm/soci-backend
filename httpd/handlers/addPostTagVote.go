package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jjcm/soci-backend/models"
)

// PostTagVoteAdditionRequest this is the shape of the JSON request that is needed to
// create a vote for post tag
type PostTagVoteAdditionRequest struct {
	PostURL string `json:"post"`
	TagName string `json:"tag"`
}

// AddPostTagVote - protected http handler
// the user associated with the passed auth token can create a new post-tag
func AddPostTagVote(w http.ResponseWriter, r *http.Request) {
	// any non GET handlers need to attach CORS headers. I always forget about that
	CorsAdjustments(&w)

	// silly AJAX prflight, here's where we can put in the CORS requirements
	if r.Method == "OPTIONS" {
		SendResponse(w, "", 200)
		return
	}

	if r.Method != "POST" {
		SendResponse(w, MakeError("You can only POST to the post creation route"), 405)
		return
	}

	// decode the request parameters 'post_id' and 'tag_id'
	var request PostTagVoteAdditionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&request)

	// get the user id from context
	userID := r.Context().Value("user_id").(int)

	// find the structure of user, post, tag with user id, post url and tag name
	user, post, tag, err := findUserPostTag(userID, request.PostURL, request.TagName)
	if err != nil {
		sendSystemError(w, err)
		return
	}

	// find the PostTagVote by post id, tag id, user id
	postTagVote := models.PostTagVote{}
	if err := postTagVote.FindByUK(post.ID, tag.ID, user.ID); err != nil {
		sendSystemError(w, fmt.Errorf("Query post-tag-vote: %v", err))
		return
	}
	// if there is existed vote rows, just return directly
	if postTagVote.PostID > 0 {
		sendSystemError(w, fmt.Errorf("PostTagVote is existed"))
		return
	}

	needUpdatePost := true
	// check if this is the first PostTagVote by user for the specific post
	votes, err := postTagVote.GetVotesByPostUser(post.ID, user.ID)
	if err != nil {
		sendSystemError(w, fmt.Errorf("Query votes: %v", err))
		return
	}
	if len(votes) > 0 {
		needUpdatePost = false
	}

	// prepare the PostTagVote for insertion
	postTagVote.Post = post
	postTagVote.PostID = post.ID
	postTagVote.PostURL = post.URL
	postTagVote.Tag = tag
	postTagVote.TagID = tag.ID
	postTagVote.TagName = tag.Name
	postTagVote.Voter = user
	postTagVote.VoterID = user.ID
	postTagVote.VoterName = user.Name

	// do many database operations with transaction
	if err = models.WithTransaction(func(tx models.Transaction) error {
		// insert the PostTagVote to database
		if err := postTagVote.CreatePostTagVote(); err != nil {
			return fmt.Errorf("Create PostTagVote: %v", err)
		}

		// increment the score for PostTag
		postTag := models.PostTag{}
		if err := postTag.IncrementScore(post.ID, tag.ID); err != nil {
			return fmt.Errorf("Increment PostTag's score: %v", err)
		}

		// check if it needs to increment the score of post
		if needUpdatePost {
			// increment the score of Post
			if err := post.IncrementScore(post.ID); err != nil {
				return fmt.Errorf("Increment Post's score: %v", err)
			}
		}

		return nil
	}); err != nil {
		sendSystemError(w, err)
		return
	}

	SendResponse(w, postTagVote, 200)
}