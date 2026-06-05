package docker

import (
	"context"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
)

type Container struct {
	ID       string
	Name     string
	Image    string
	Status   string
	State    string
	Created  time.Time
	Started  time.Time
	Restarts int
	ExitCode int
}

type Client struct {
	dc *dockerclient.Client
}

func NewClient(socketPath string) (*Client, error) {
	dc, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost("unix://"+socketPath),
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &Client{dc: dc}, nil
}

func (c *Client) Containers(ctx context.Context) ([]Container, error) {
	list, err := c.dc.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	result := make([]Container, 0, len(list))
	for _, ct := range list {
		name := ""
		if len(ct.Names) > 0 {
			name = ct.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		insp, err := c.dc.ContainerInspect(ctx, ct.ID)
		restarts := 0
		exitCode := 0
		var started time.Time
		var created time.Time
		if err == nil {
			restarts = insp.RestartCount
			exitCode = insp.State.ExitCode
			started, _ = time.Parse(time.RFC3339Nano, insp.State.StartedAt)
			created, _ = time.Parse(time.RFC3339Nano, insp.Created)
		}

		result = append(result, Container{
			ID:       ct.ID[:12],
			Name:     name,
			Image:    ct.Image,
			Status:   ct.Status,
			State:    ct.State,
			Created:  created,
			Started:  started,
			Restarts: restarts,
			ExitCode: exitCode,
		})
	}
	return result, nil
}
