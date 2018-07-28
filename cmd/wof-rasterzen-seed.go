package main

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/murphy214/tile-cover"
	"github.com/paulmach/go.geojson"
	"io"
	"io/ioutil"
	"log"
	"sync"
)

func seed(ctx context.Context, f *geojson.Feature, i int) {

	tileids := tilecover.TileCover(f, i)

	for _, t := range tileids {

		select {
		case <-ctx.Done():
			break
		default:
			z := t.Z
			x := t.X
			y := t.Y

			log.Println("CALL FETCH TILES WITH", z, x, y)
			// rasterzen.FetchTileWithCache(c, z, x, y)
		}
	}

}

func main() {

	var mode = flag.String("mode", "repo", "")
	var min_zoom = flag.Int("min-zoom", 1, "")
	var max_zoom = flag.Int("max-zoom", 16, "")

	flag.Parse()

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

		f, err := geojson.UnmarshalFeature(b)

		if err != nil {
			return err
		}

		wg := new(sync.WaitGroup)

		for i := *min_zoom; i < *max_zoom; i++ {

			wg.Add(1)

			go func(ctx context.Context, f *geojson.Feature, i int) {

				defer wg.Done()
				seed(ctx, f, i)

			}(ctx, f, i)
		}

		wg.Wait()

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

}
