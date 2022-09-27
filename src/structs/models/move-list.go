package models

type MoveListModel struct {
	WorkerId string `json:"worker_id" binding:"required"`
	Id       string `json:"id" binding:"required"`
	Queue    string `json:"queue" binding:"required"`
	GotError bool   `json:"got_error"`
}
