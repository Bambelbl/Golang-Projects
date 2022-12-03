package user

type User struct {
	ID       string `json:"id" bson:"id"`
	Username string `json:"username" bson:"username"`
	password string `schema:"-" bson:"password"`
}

// go install github.com/golang/mock/mockgen@v1.6.0
//
//go:generate mockgen -source=user.go -destination=repo_mock.go -package=user UsersRepo
type UsersRepo interface {
	Authorize(login, pass string) (*User, error)
	AddUser(id, login, pass string) error
}
