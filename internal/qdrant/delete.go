package qdrant

import (
	"context"
	"fmt"
)

func (c *Client) DeleteByFilter(ctx context.Context, collection string, filter Filter) error {
	body := map[string]any{"filter": filter}
	return c.post(ctx, fmt.Sprintf("/collections/%s/points/delete?wait=true", collection), body, nil)
}
