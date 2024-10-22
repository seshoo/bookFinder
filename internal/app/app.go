package app

import (
	"context"
	"fmt"
	"github.com/seshoo/bookFinder/internal/repository/elastic"
	"github.com/seshoo/bookFinder/internal/service"
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

type Options struct {
	Dbg       bool       `long:"dbg" env:"DBG" description:"debug mode"`
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

func Run(opts Options) error {
	fmt.Printf("opts: %+v", opts)
	var err error

	db, err := elasticsearch.NewElastic(elasticsearch.Config{
		Host:      opts.Elastic.Host,
		Username:  opts.Elastic.Username,
		Password:  opts.Elastic.Password,
		IndexName: opts.Elastic.Index,
	})
	if err != nil {
		return err
	}

	mapping, err := os.ReadFile(opts.Elastic.MappingConfig)
	if err != nil {
		return err
	}

	err = db.Prepare(context.Background(), mapping)
	if err != nil {
		return err
	}

	repositories := elastic.NewElasticRepository(db)
	services := service.NewServices(service.Deps{
		DpUrlTmp:     opts.Dp.UrlTemplate,
		Repositories: repositories,
	})

	_ = services

	return nil
}
