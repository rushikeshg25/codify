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

func (c *Codeground) DeleteCodeground() {
	logError(c.K8s.DeleteIngress(*c))
	logError(c.K8s.DeleteService(*c))
	logError(c.K8s.DeleteDeployment(*c))
}

func (c *Codeground) StartCodeground() {}

// StopCodeground tears down the running k8s resources for a codeground.
// Currently equivalent to a delete; persistent state is not retained.
func (c *Codeground) StopCodeground() {
	c.DeleteCodeground()
}
