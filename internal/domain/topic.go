package domain

type Topic struct {
	Id    string `json:"id" binding:"required"`
	Title string `json:"title" binding:"required"`
	Link  string `json:"link" binding:"required"`
	Text  string `json:"text" binding:"required"`
}

type Topics []Topic

func (ts Topics) ids() []string {
	ids := make([]string, len(ts))
	for i, t := range ts {
		ids[i] = t.Id
	}
	return ids
}
