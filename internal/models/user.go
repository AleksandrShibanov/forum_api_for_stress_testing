package models

type User struct {
	Nickname string `json:"nickname,omitempty"`
	Fullname string `json:"fullname,omitempty"`
	About    string `json:"about,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Users = []User

type UserUpdate struct {
	Fullname *string
	About    *string
	Email    *string
}
