package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/jtacoma/uritemplates"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-rasterzen/worker"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/warning"
	"io"
	"os"
	"strings"
)

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	var count = flag.Bool("count", false, "Display the number of tiles to process and exit.")

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
	// seed_geojson := flag.Bool("seed-geojson", true, "Seed GeoJSON tiles.")
	seed_svg := flag.Bool("seed-svg", false, "Seed SVG tiles.")
	seed_png := flag.Bool("seed-png", false, "Seed PNG tiles.")
	seed_all := flag.Bool("seed-all", false, "See all the tile formats")

	seed_worker := flag.String("seed-worker", "local", "The type of worker for seeding tiles. Valid workers are: lambda, local.")
	max_workers := flag.Int("seed-max-workers", 100, "The maximum number of concurrent workers to invoke when seeding tiles")

	var lambda_dsn flags.DSNString
	flag.Var(&lambda_dsn, "lambda-dsn", "A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'")

	var exclude flags.KeyValueArgs
	var include flags.KeyValueArgs

	flag.Var(&exclude, "exclude", "Exclude records not matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")
	flag.Var(&include, "include", "Include only those records matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.")

	lambda_function := flag.String("lambda-function", "Rasterzen", "A valid AWS Lambda function name.")

	timings := flag.Bool("timings", false, "Display timings for tile seeding.")

	flag.Parse()

	if *seed_all {
		*seed_rasterzen = true
		// *seed_geojson = true
		*seed_svg = true
		*seed_png = true
	}

	logger := log.SimpleWOFLogger()

	writer := io.MultiWriter(os.Stdout)
	logger.AddLogger(writer, "status")

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

	svg_opts, err := tile.DefaultRasterzenSVGOptions()

	if err != nil {
		logger.Fatal(err)
	}

	var w worker.Worker
	var w_err error

	switch strings.ToUpper(*seed_worker) {

	case "LAMBDA":
		w, w_err = worker.NewLambdaWorker(lambda_dsn.Map(), *lambda_function, c, nz_opts, svg_opts)
	case "LOCAL":
		w, w_err = worker.NewLocalWorker(c, nz_opts, svg_opts)
	default:
		w_err = errors.New("Invalid worker")

	}

	if w_err != nil {
		logger.Fatal(w_err)
	}

	seeder, err := seed.NewTileSeeder(w)

	if err != nil {
		logger.Fatal(err)
	}

	seeder.MaxWorkers = *max_workers
	seeder.Logger = logger
	seeder.Timings = *timings

	seeder.SeedRasterzen = *seed_rasterzen
	seeder.SeedSVG = *seed_svg
	seeder.SeedPNG = *seed_png

	ts, err := seed.NewTileSet()

	if err != nil {
		logger.Fatal(err)
	}

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

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

				path, _ := index.PathForContext(ctx)

				logger.Warning("%s triggered a critical error (%s)", path, err)
				return err
			}

			logger.Warning(err)
		}

		for _, e := range exclude {

			path := e.Key
			test := e.Value

			rsp := gjson.GetBytes(f.Bytes(), path)

			if rsp.Exists() && rsp.String() == test {
				logger.Status("%s (%s) is being exluded because it matches the %s=%s -exclude test", f.Name(), f.Id(), path, test)
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
					logger.Status("%s (%s) fails -include test because %s does not exist", f.Name(), f.Id(), path)
					continue
				}

				if rsp.String() != test {
					logger.Status("%s (%s) fails -include test because %s != %s (is %s)", f.Name(), f.Id(), path, test, rsp.String())
					continue
				}

				include_ok = true
				break
			}

			if !include_ok {
				logger.Status("%s (%s) is being excluded because all -include tests failed", f.Name(), f.Id())
				return nil
			}
		}

		bboxes, err := f.BoundingBoxes()

		if err != nil {
			return err
		}

		for _, bounds := range bboxes.Bounds() {

			min := [2]float64{bounds.Min.X, bounds.Min.Y}
			max := [2]float64{bounds.Max.Y, bounds.Max.Y}

			ex := geom.NewExtent(min, max)

			for z := *min_zoom; z < *max_zoom; z++ {

				for _, t := range slippy.FromBounds(ex, uint(z)) {
					ts.AddTile(t)
				}
			}
		}

		return nil
	}

	idx, err := index.NewIndexer(*mode, cb)

	if err != nil {
		logger.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := idx.IndexPath(path)

		if err != nil {
			logger.Fatal(err)
		}
	}

	if *count {
		fmt.Println(ts.Count())
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ok, errors := seeder.SeedTileSet(ctx, ts)

	if !ok {

		for _, e := range errors {
			logger.Error(e)
		}

		logger.Fatal("tile seeding completed unsuccessfully")
	}

	os.Exit(0)
}