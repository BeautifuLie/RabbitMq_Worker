package model

type Joke struct {
	Title string `json:"title" bson:"title"`
	Body  string `json:"body" bson:"body"`
	Score int    `json:"score" bson:"score"`
	ID    string `json:"id" bson:"id"`
}
