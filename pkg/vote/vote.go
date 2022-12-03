package vote

type Vote struct {
	UserID string `json:"user" bson:"user"`
	Vote   int    `json:"vote" bson:"vote"`
}
