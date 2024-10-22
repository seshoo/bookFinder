package domain

type Topic struct {
	Id    string `json:"id" binding:"required"`
	Title string `json:"title" binding:"required"`
	Link  string `json:"link" binding:"required"`
	Text  string `json:"text" binding:"required"`
}
