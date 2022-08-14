package models

type UserRequest struct {
	Method string `json:"method"`
	Route  string `json:"route"`
}
