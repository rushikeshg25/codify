package controller

import "database/sql"

type PlaygroundController struct {
	db *sql.DB
}

func NewPlaygroundController(db *sql.DB) *PlaygroundController {
	return &PlaygroundController{
		db: db,
	}
}