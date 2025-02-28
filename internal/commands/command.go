package commands

import (
	"context"
	"github.com/seshoo/bookFinder/internal/repository/elastic"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
	"os"
)

type LogOptions struct {
	ToFile    bool `long:"to-file" env:"TO_FILE" description:"log in file"`
	ToConsole bool `long:"to-console" env:"TO_CONSOLE" description:"log in console"`
}

type DataProvider struct {
	UrlTemplate string `long:"dp-url-tmp" env:"DP_URL_TMP" default:"https://rutracker.net/forum/viewtopic.php?t=%s" description:"Template of url"`
}

type CommonOpts struct {
	Dbg       bool       `short:"d" long:"debug" description:"debug mode"`
	LogConfig LogOptions `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	Dp        DataProvider
	Elastic   struct {
		Host          string `long:"host" env:"HOST" default:"http://localhost:9200" description:"Elasticsearch host"`
		Username      string `long:"username" env:"USERNAME" default:"admin" description:"Elasticsearch username"`
		Password      string `long:"password" env:"PASSWORD" default:"admin" description:"Elasticsearch password"`
		Index         string `long:"index" env:"INDEX" default:"books" description:"Elasticsearch index"`
		MappingConfig string `long:"mapping" env:"MAPPING" default:"./settings/elasticmapping.json" description:"Elasticsearch mapping"`
	} `group:"Elastic" namespace:"elastic" env-namespace:"ELASTIC"`
}

func (c *CommonOpts) initElasticRepository(ctx context.Context) (*elastic.ElasticRepository, error) {
	var err error

	db, err := elasticsearch.NewElastic(elasticsearch.Config{
		Host:      c.Elastic.Host,
		Username:  c.Elastic.Username,
		Password:  c.Elastic.Password,
		IndexName: c.Elastic.Index,
	})
	if err != nil {
		return nil, err
	}

	mapping, err := os.ReadFile(c.Elastic.MappingConfig)
	if err != nil {
		return nil, err
	}

	err = db.Prepare(ctx, mapping)
	if err != nil {
		return nil, err
	}

	repositories := elastic.NewElasticRepository(db)

	return repositories, nil
}
