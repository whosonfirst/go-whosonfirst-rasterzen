package seed

import (
	"fmt"
	rz_seed "github.com/whosonfirst/go-rasterzen/seed"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

func NewGatherTilesFeatureFunc(f geojson.Feature, min_zoom int, max_zoom int) (rz_seed.GatherTilesFunc, error) {

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return nil, err
	}

	str_bounds := make([]string, 0)

	for _, bounds := range bboxes.Bounds() {

		str_extent := fmt.Sprintf("%0.6f,%0.6f,%0.6f,%0.6f", bounds.Min.X, bounds.Min.Y, bounds.Max.Y, bounds.Max.Y)
		str_bounds = append(str_bounds, str_extent)
	}

	return rz_seed.NewGatherTilesExtentFunc(str_bounds, ",", min_zoom, max_zoom)
}
