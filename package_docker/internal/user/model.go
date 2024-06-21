package user

type User struct {
	ID           string `json:"id" bson:"_id,omitempty"`
	Username     string `json:"username" bson:"username"`
	PasswordHash string `json:"" bson:"password"`
	Email        string `json:"email" bson:"email"`
	IsAdmin      bool   `json:"root" bson:"root"`
}
type CreateUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
