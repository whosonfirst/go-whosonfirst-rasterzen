package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"os"
)

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	var count = flag.Bool("count", false, "")

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

	seed_svg := flag.Bool("seed-svg", true, "Seed SVG tiles.")
	seed_png := flag.Bool("seed-png", false, "Seed PNG tiles.")
	seed_timings := flag.Bool("seed-timings", "Display timings when seeding tiles.")
	seed_workers := flag.Int("seed-workers", 100, "The maximum number of concurrent workers to invoke when seeding tiles")

	flag.Parse()

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

	seeder, err := seed.NewTileSeeder(c, nz_opts)

	if err != nil {
		logger.Fatal(err)
	}

	seeder.SeedSVG = *seed_svg
	seeder.SeedPNG = *seed_png
	seeder.MaxWorkers = *seed_workers
	seeder.Timings = *seed_timings
	seeder.Logger = logger

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

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		bboxes, err := f.BoundingBoxes()

		if err != nil {
			return err
		}

		mbr := bboxes.MBR()

		min := [2]float64{mbr.Min.X, mbr.Min.Y}
		max := [2]float64{mbr.Max.Y, mbr.Max.Y}

		ex := geom.NewExtent(min, max)

		for z := *min_zoom; z < *max_zoom; z++ {

			for _, t := range slippy.FromBounds(ex, uint(z)) {
				ts.AddTile(t)
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

	ok, errors := seeder.SeedTileSet(ts)

	if !ok {

		for _, e := range errors {
			logger.Error(e)
		}

		logger.Fatal("tile seeding completed unsuccessfully")
	}

	os.Exit(0)
}
