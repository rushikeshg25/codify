package main

import (
	"fmt"
	"log"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type Codeground struct {
	Id             string `json:"id"`
	UserId         int    `json:"user_id"`
	Name           string `json:"name"`
	CodegroundType string `json:"codeground_type"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	K8s            *K8s
}

/*
CreateCodeground provisions the resources a codeground needs, in dependency
order, and unwinds them if any step fails.

Previously every step logged its error and carried on, so a Deployment that
could not be scheduled still got a Service and an Ingress: a routable hostname
pointing at nothing, and a user staring at a spinner with no way to find out
why.

The Secret has to exist before the Deployment that mounts it, and the
NetworkPolicy before the pod it constrains, so that a pod is never briefly
reachable on an open network.
*/
func (c *Codeground) CreateCodeground() error {
	steps := []struct {
		name string
		fn   func(Codeground) error
	}{
		{"network policy", c.K8s.CreateNetworkPolicy},
		{"secret", c.K8s.CreateSecret},
		{"deployment", c.K8s.CreateDeployment},
		{"service", c.K8s.CreateService},
		{"ingress", c.K8s.CreateIngress},
	}

	for _, step := range steps {
		if err := step.fn(*c); err != nil {
			log.Printf("codeground %s: %s failed, rolling back: %v", c.Id, step.name, err)
			if rbErr := c.DeleteCodeground(); rbErr != nil {
				// Not fatal: the reconciler sweeps whatever is left behind.
				log.Printf("codeground %s: rollback incomplete: %v", c.Id, rbErr)
			}
			return fmt.Errorf("create %s: %w", step.name, err)
		}
	}

	return nil
}

// DeleteCodeground tears down every resource belonging to a codeground.
//
// Resources that are already gone are not an error: teardown has to be safe to
// retry, both for the rollback path above and for the reconciler.
func (c *Codeground) DeleteCodeground() error {
	steps := []struct {
		name string
		fn   func(Codeground) error
	}{
		{"ingress", c.K8s.DeleteIngress},
		{"service", c.K8s.DeleteService},
		{"deployment", c.K8s.DeleteDeployment},
		{"secret", c.K8s.DeleteSecret},
		{"network policy", c.K8s.DeleteNetworkPolicy},
	}

	var failures []error
	for _, step := range steps {
		if err := step.fn(*c); err != nil && !k8serrors.IsNotFound(err) {
			failures = append(failures, fmt.Errorf("delete %s: %w", step.name, err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("codeground %s teardown: %v", c.Id, failures)
	}
	return nil
}

func (c *Codeground) StartCodeground() error { return c.CreateCodeground() }

// StopCodeground tears down the running k8s resources for a codeground.
// Currently equivalent to a delete; persistent state is not retained.
func (c *Codeground) StopCodeground() error {
	return c.DeleteCodeground()
}
