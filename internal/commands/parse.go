package commands

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/seshoo/bookFinder/internal/service"
)

type ParseCommand struct {
	CommonOpts
}

func (c *ParseCommand) Execute(_ []string) error {
	fmt.Println("Parse Command")
	fmt.Printf("Debug Mode: %t\n", c.Dbg)

	var ctx = context.Background()

	repositories, err := c.initElasticRepository(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to init elastic repository")
	}

	services := service.NewServices(service.Deps{
		DpUrlTmp:     c.Dp.UrlTemplate,
		Repositories: repositories,
	})

	_ = services

	return nil
}
