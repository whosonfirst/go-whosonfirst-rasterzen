package main

/*

go run cmd/wof-rasterzen-tiles/main.go -include properties.wof:placetype=country -include properties.wof:placetype=disputed -include properties.wof:placetype=dependency -min-zoom 2 -max-zoom 3 -mode git -seed-tileset-catalog-dsn 'catalog=sqlite dsn=tmp/rasterzen.db' https://github.com/sfomuseum-data/sfomuseum-data-whosonfirst.git | sort
2/1/1
2/1/2
2/2/1
2/2/2
2/3/1
3/2/3
3/3/3
3/3/4
3/4/3
3/4/4
3/5/3
3/6/3
3/7/3

*/

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-string/dsn"
	"github.com/tidwall/gjson"
	rz_seed "github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-rasterzen/seed"
	"github.com/whosonfirst/warning"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	var exclude flags.KeyValueArgs
	var include flags.KeyValueArgs

	flag.Var(&exclude, "exclude", "Exclude records not matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")
	flag.Var(&include, "include", "Include only those records matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")

	seed_tileset_catalog_dsn := flag.String("seed-tileset-catalog-dsn", "catalog=sync", "A valid tile.SeedCatalog DSN string. Required parameters are 'catalog=CATALOG'")

	timings := flag.Bool("timings", false, "Display timings for tile seeding.")
	log_level := flag.String("log-level", "status", "Log level to use for logging")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	writer := io.MultiWriter(os.Stdout)
	logger.AddLogger(writer, *log_level)

	dsn_str := *seed_tileset_catalog_dsn
	dsn_map, err := dsn.StringToDSNWithKeys(dsn_str, "catalog")

	if err != nil {
		logger.Fatal(err)
	}

	if strings.ToUpper(dsn_map["catalog"]) == "SQLITE" {

		tmpfile, err := ioutil.TempFile("", "rasterzen")

		if err != nil {
			logger.Fatal(err)
		}

		err = tmpfile.Close()

		if err != nil {
			logger.Fatal(err)
		}

		dsn_map["dsn"] = tmpfile.Name()
		dsn_str = dsn_map.String()

		defer os.Remove(tmpfile.Name())
	}

	tileset, err := rz_seed.NewTileSetFromDSN(dsn_str)

	if err != nil {
		logger.Fatal(err)
	}

	tileset.Logger = logger
	tileset.Timings = *timings

	index_func := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		principal, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !principal {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {

			if !warning.IsWarning(err) {
				return err
			}

			logger.Warning(err)
		}

		for _, e := range exclude {

			path := e.Key
			test := e.Value

			rsp := gjson.GetBytes(f.Bytes(), path)

			if rsp.Exists() && rsp.String() == test {
				logger.Debug("%s (%s) is being exluded because it matches the %s=%s -exclude test", f.Name(), f.Id(), path, test)
				return nil
			}
		}

		if len(include) > 0 {

			include_ok := false

			for _, i := range include {

				path := i.Key
				test := i.Value

				rsp := gjson.GetBytes(f.Bytes(), path)

				if !rsp.Exists() {
					logger.Debug("%s (%s) fails -include test because %s does not exist", f.Name(), f.Id(), path)
					continue
				}

				if rsp.String() != test {
					logger.Debug("%s (%s) fails -include test because %s != %s (is %s)", f.Name(), f.Id(), path, test, rsp.String())
					continue
				}

				include_ok = true
				break
			}

			if !include_ok {
				logger.Debug("%s (%s) is being excluded because all -include tests failed", f.Name(), f.Id())
				return nil
			}
		}

		gather_func, err := seed.NewGatherTilesFeatureFunc(f, *min_zoom, *max_zoom)

		if err != nil {
			return err
		}

		err = gather_func(ctx, tileset)

		if err != nil {
			return err
		}

		return nil
	}

	idx, err := index.NewIndexer(*mode, index_func)

	if err != nil {
		logger.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := idx.IndexPath(path)

		if err != nil {
			logger.Fatal(err)
		}
	}

	tileset.Range(func(key interface{}, value interface{}) bool {

		str_key := key.(string)
		fmt.Println(str_key)
		return true
	})

	os.Exit(0)
}
