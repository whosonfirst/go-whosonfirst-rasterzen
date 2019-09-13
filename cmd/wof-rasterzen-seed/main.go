package main

// this is an almost stright-clone of go-rasterzen/cmd/rasterzen-seed/main.go
// which means any times there are changes in go-rasterzen we usually have to
// update things here - something more "unified" would be better but I am not
// sure what that is yet... (20190904/thisisaaronland)

import (
	"context"
	"errors"
	"flag"
	_ "fmt"
	"github.com/aaronland/go-string/dsn"
	"github.com/jtacoma/uritemplates"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	rz_seed "github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
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
	"time"
)

type WOFGatherTilesOptions struct {
	Mode    string
	MinZoom int
	MaxZoom int
	Include flags.KeyValueArgs
	Exclude flags.KeyValueArgs
	Paths   []string
}

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	nextzen_apikey := flag.String("nextzen-apikey", "", "A valid Nextzen API key.")
	nextzen_origin := flag.String("nextzen-origin", "", "An optional HTTP 'Origin' host to pass along with your Nextzen requests.")
	nextzen_debug := flag.Bool("nextzen-debug", false, "Log requests (to STDOUT) to Nextzen tile servers.")
	nextzen_uri := flag.String("nextzen-uri", "", "A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.")

	go_cache := flag.Bool("go-cache", false, "Cache tiles with an in-memory (go-cache) cache.")
	fs_cache := flag.Bool("fs-cache", false, "Cache tiles with a filesystem-based cache.")
	fs_root := flag.String("fs-root", "", "The root of your filesystem cache. If empty rasterd will try to use the current working directory.")
	s3_cache := flag.Bool("s3-cache", false, "Cache tiles with a S3-based cache.")
	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	seed_rasterzen := flag.Bool("seed-rasterzen", false, "Seed Rasterzen tiles.")
	seed_svg := flag.Bool("seed-svg", false, "Seed SVG tiles.")
	seed_png := flag.Bool("seed-png", false, "Seed PNG tiles.")
	seed_all := flag.Bool("seed-all", false, "See all the tile formats")

	refresh_rasterzen := flag.Bool("refresh-rasterzen", false, "Force rasterzen tiles to be generated even if they are already cached.")
	refresh_svg := flag.Bool("refresh-svg", false, "Force SVG tiles to be generated even if they are already cached.")
	refresh_png := flag.Bool("refresh-png", false, "Force PNG tiles to be generated even if they are already cached.")
	refresh_all := flag.Bool("refresh-all", false, "Force all tiles to be generated even if they are already cached.")

	seed_worker := flag.String("seed-worker", "local", "The type of worker for seeding tiles. Valid workers are: lambda, local, sqs.")
	max_workers := flag.Int("seed-max-workers", 0, "The maximum number of concurrent workers to invoke when seeding tiles. The default is the value of `runtime.NumCPU()` * 2.")

	var lambda_dsn flags.DSNString
	flag.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	var sqs_dsn flags.DSNString
	flag.Var(&sqs_dsn, "sqs-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION' and 'queue=QUEUE'")

	var exclude flags.KeyValueArgs
	var include flags.KeyValueArgs

	flag.Var(&exclude, "exclude", "Exclude records not matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")
	flag.Var(&include, "include", "Include only those records matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")

	custom_rz_options := flag.String("rasterzen-options", "", "The path to a valid RasterzenOptions JSON file.")
	custom_svg_options := flag.String("svg-options", "", "The path to a valid RasterzenSVGOptions JSON file.")
	custom_png_options := flag.String("png-options", "", "The path to a valid RasterzenPNGOptions JSON file.")

	seed_tileset_catalog_dsn := flag.String("seed-tileset-catalog-dsn", "catalog=sync", "A valid tile.SeedCatalog DSN string. Required parameters are 'catalog=CATALOG'")

	lambda_function := flag.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	timings := flag.Bool("timings", false, "Display timings for tile seeding.")
	log_level := flag.String("log-level", "status", "Log level to use for logging")

	flag.Parse()

	if *seed_all {
		*seed_rasterzen = true
		*seed_svg = true
		*seed_png = true
	}

	if *refresh_all {
		*refresh_rasterzen = true
		*refresh_svg = true
		*refresh_png = true
	}

	logger := log.SimpleWOFLogger()

	writer := io.MultiWriter(os.Stdout)
	logger.AddLogger(writer, *log_level)

	nz_opts := &nextzen.Options{
		ApiKey: *nextzen_apikey,
		Origin: *nextzen_origin,
		Debug:  *nextzen_debug,
	}

	if *nextzen_uri != "" {

		template, err := uritemplates.Parse(*nextzen_uri)

		if err != nil {
			logger.Fatal(err)
		}

		nz_opts.URITemplate = template
	}

	caches := make([]cache.Cache, 0)

	if *go_cache {

		logger.Info("enable go-cache cache layer")

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			logger.Fatal(err)
		}

		c, err := cache.NewGoCache(opts)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *fs_cache {

		logger.Info("enable filesystem cache layer")

		if *fs_root == "" {

			cwd, err := os.Getwd()

			if err != nil {
				logger.Fatal(err)
			}

			*fs_root = cwd
		}

		c, err := cache.NewFSCache(*fs_root)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if *s3_cache {

		logger.Info("enable S3 cache layer")

		opts, err := s3.NewS3CacheOptionsFromString(*s3_opts)

		if err != nil {
			logger.Fatal(err)
		}

		c, err := s3.NewS3Cache(*s3_dsn, opts)

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	if len(caches) == 0 {

		// because we still need to pass a cache.Cache thingy
		// around (20180612/thisisaaronland)

		c, err := cache.NewNullCache()

		if err != nil {
			logger.Fatal(err)
		}

		caches = append(caches, c)
	}

	c, err := cache.NewMultiCache(caches)

	if err != nil {
		logger.Fatal(err)
	}

	var rz_opts *tile.RasterzenOptions
	var svg_opts *tile.RasterzenSVGOptions
	var png_opts *tile.RasterzenPNGOptions

	if *custom_rz_options != "" {

		var opts *tile.RasterzenOptions

		if strings.HasPrefix(*custom_png_options, "{") {
			opts, err = tile.RasterzenOptionsFromString(*custom_rz_options)
		} else {
			opts, err = tile.RasterzenOptionsFromFile(*custom_rz_options)
		}

		if err != nil {
			logger.Fatal(err)
		}

		rz_opts = opts

	} else {

		opts, err := tile.DefaultRasterzenOptions()

		if err != nil {
			logger.Fatal(err)
		}

		rz_opts = opts
	}

	if *seed_svg || *seed_png {

		if *custom_svg_options != "" {

			var opts *tile.RasterzenSVGOptions

			if strings.HasPrefix(*custom_svg_options, "{") {
				opts, err = tile.RasterzenSVGOptionsFromString(*custom_svg_options)
			} else {
				opts, err = tile.RasterzenSVGOptionsFromFile(*custom_svg_options)
			}

			if err != nil {
				logger.Fatal(err)
			}

			svg_opts = opts

		} else {

			opts, err := tile.DefaultRasterzenSVGOptions()

			if err != nil {
				logger.Fatal(err)
			}

			svg_opts = opts
		}

		if *custom_png_options != "" {

			var opts *tile.RasterzenPNGOptions

			if strings.HasPrefix(*custom_png_options, "{") {
				opts, err = tile.RasterzenPNGOptionsFromString(*custom_png_options)
			} else {
				opts, err = tile.RasterzenPNGOptionsFromFile(*custom_png_options)
			}

			if err != nil {
				logger.Fatal(err)
			}

			png_opts = opts

		} else {

			opts, err := tile.DefaultRasterzenPNGOptions()

			if err != nil {
				logger.Fatal(err)
			}

			png_opts = opts
		}

		svg_opts.Refresh = *refresh_svg
		png_opts.Refresh = *refresh_png
	}

	rz_opts.Refresh = *refresh_rasterzen

	var w worker.Worker
	var w_err error

	switch strings.ToUpper(*seed_worker) {

	case "LAMBDA":
		w, w_err = worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, c, nz_opts, rz_opts, svg_opts, png_opts)
	case "LOCAL":
		w, w_err = worker.NewLocalWorker(c, nz_opts, rz_opts, svg_opts, png_opts)
	case "SQS":
		w, w_err = worker.NewSQSWorker(sqs_dsn.Map())
	default:
		w_err = errors.New("Invalid worker")
	}

	if w_err != nil {
		logger.Fatal(w_err)
	}

	seeder, err := rz_seed.NewTileSeeder(w, c)

	if err != nil {
		logger.Fatal(err)
	}

	if *max_workers > 0 {
		seeder.MaxWorkers = *max_workers
	}

	seeder.Logger = logger
	seeder.Timings = *timings

	seeder.SeedRasterzen = *seed_rasterzen
	seeder.SeedSVG = *seed_svg
	seeder.SeedPNG = *seed_png

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

		tileset.Logger.Status("Gather tiles for %s", f.Name())

		gather_func, err := seed.NewGatherTilesFeatureFunc(f, *min_zoom, *max_zoom)

		if err != nil {
			return err
		}

		t1 := time.Now()

		defer func() {
			tileset.Logger.Status("Time to gather tiles for %s: %v", f.Name(), time.Since(t1))
		}()

		if tileset.Timings {

			ticker := time.NewTicker(time.Second * 10)
			ticker_ch := make(chan bool)

			defer func() {
				ticker_ch <- true
			}()

			go func() {

				for range ticker.C {

					select {
					case <-ctx.Done():
						return
					case <-ticker_ch:
						return
					default:
						count := tileset.Count()
						tileset.Logger.Status("%s (%s) %d tiles gathered so far (%v)", f.Name(), f.Id(), count, time.Since(t1))
					}
				}
			}()
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

	t1 := time.Now()

	for _, path := range flag.Args() {

		err := idx.IndexPath(path)

		if err != nil {
			logger.Fatal(err)
		}
	}

	count := tileset.Count()
	logger.Status("Time to gather %d tiles: %v", count, time.Since(t1))

	defer func() {
		logger.Status("Time to seed tileset: %v", time.Since(t1))
	}()

	ok, errors := seeder.SeedTileSet(context.Background(), tileset)

	// how best to handle this...

	if !ok {

		for _, e := range errors {
			tileset.Logger.Error("Failed to seed tiles: %s", e)
		}
	}

	os.Exit(0)
}
