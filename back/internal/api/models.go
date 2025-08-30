package api

type StartSessionRequest struct {
	Model string `json:"model" binding:"required"`
}
