package main

import "log"

type Codeground struct {
	Id             string `json:"id"`
	UserId         int    `json:"user_id"`
	Name           string `json:"name"`
	CodegroundType string `json:"codeground_type"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	K8s            *K8s
}

func logError(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func (c *Codeground) CreateCodeground() {
	var err error
	err = c.K8s.CreateDeployment(*c)
	logError(err)
	err = c.K8s.CreateService(*c)
	logError(err)
	err = c.K8s.CreateIngress(*c)
	logError(err)
}

func (c *Codeground) DeleteCodeground() {}

func (c *Codeground) StartCodeground() {}

func (c *Codeground) StopCodeground() {}
