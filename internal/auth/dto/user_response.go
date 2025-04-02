package dto

type UserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Id    int64  `json:"id"`
}
