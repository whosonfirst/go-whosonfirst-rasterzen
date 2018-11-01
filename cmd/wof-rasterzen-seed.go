package main

import (
	"context"
	"flag"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-whosonfirst-geojson/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"io"
	"io/ioutil"
	"log"
)

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	flag.Parse()

	seeder, err := seed.NewTileSeeder()

	if err != nil {
		log.Fatal(err)
	}

	ts, err := seed.NewTileSet()

	if err != nil {
		return nil
	}

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		b, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
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

		ex := geom.NewExtent(min, max), nil

		for z := *min_zoom; z < *max_zoom; z++ {

			for _, t := range slippy.FromBounds(ex, uint(z)) {
				ts.AddTile(t)
			}
		}

		return nil
	}

	idx, err := index.NewIndexer(*mode, cb)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := idx.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}
	}

	seeder.SeedTileSet(ts)
}
