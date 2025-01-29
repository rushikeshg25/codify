package main

import "time"

type Codeground struct {
	ID             string    `json:"id"`
	UserID         int       `json:"user_id"`
	Name           string    `json:"name"`
	CodegroundType string    `json:"codeground_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (c *Codeground) CreateCodeground(){}

func (c *Codeground) DeleteCodeground(){}

func (c *Codeground) StartCodeground(){}

func (c *Codeground) StopCodeground(){}