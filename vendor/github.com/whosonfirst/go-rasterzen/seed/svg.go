package seed

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-rasterzen/nextzen"
	"github.com/whosonfirst/go-rasterzen/tile"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	_ "path/filepath"
)

func SeedSVG(t slippy.Tile, c cache.Cache, nz_opts *nextzen.Options) (io.ReadCloser, error) {

	z := int(t.Z)
	x := int(t.X)
	y := int(t.Y)

	svg_key := fmt.Sprintf("svg/%d/%d/%d.svg", z, x, y)

	var svg_data io.ReadCloser
	var err error

	svg_data, err = c.Get(svg_key)

	if err == nil {
		return svg_data, nil
	}

	geojson_fh, err := SeedRasterzen(t, c, nz_opts)

	if err != nil {
		return nil, err
	}

	defer geojson_fh.Close()

	var buf bytes.Buffer
	svg_wr := bufio.NewWriter(&buf)

	err = tile.ToSVG(geojson_fh, svg_wr)

	if err != nil {
		return nil, err
	}

	svg_wr.Flush()

	r := bytes.NewReader(buf.Bytes())
	svg_fh := ioutil.NopCloser(r)

	return c.Set(svg_key, svg_fh)
}
