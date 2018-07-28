package tilecover

import (
	"fmt"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	"math"
)

//
func RunPolygon(polygon [][][]float64, tile m.TileID) bool {
	for _, cont := range polygon {
		for _, pt := range cont {
			tileid := m.Tile(pt[0], pt[1], int(tile.Z))
			if tileid.X == tile.X && tileid.Y == tile.Y {
				return true
			}
		}
	}
	return false
}

// checks a single tile
func TilePolygon(polygon [][][]float64, polygonbds m.Extrema, tileid m.TileID) bool {
	bds := m.Bounds(tileid)

	if polygonbds.N < bds.N && polygonbds.S > bds.S && polygonbds.E < bds.E && polygonbds.W > bds.W {
		return true
	}

	wn := []float64{bds.W, bds.N}
	ws := []float64{bds.W, bds.S}
	en := []float64{bds.E, bds.N}
	es := []float64{bds.E, bds.S}

	return Pip(polygon, wn) || Pip(polygon, ws) || Pip(polygon, en) || Pip(polygon, es) || RunPolygon(polygon, tileid)
}

// point in polygon
func Pip(cont [][][]float64, p []float64) bool {
	// Cast ray from p.x towards the right
	intersections := 0
	for _, c := range cont {
		for i := range c {
			curr := c[i]
			ii := i + 1
			if ii == len(c) {
				ii = 0
			}
			next := c[ii]

			// Is the point out of the edge's bounding box?
			// bottom vertex is inclusive (belongs to edge), top vertex is
			// exclusive (not part of edge) -- i.e. p lies "slightly above
			// the ray"
			bottom, top := curr, next
			if bottom[1] >= top[1] {
				bottom, top = top, bottom
			}
			if p[1] < bottom[1] || p[1] >= top[1] {
				continue
			}
			// Edge is from curr to next.

			if p[0] >= math.Max(curr[0], next[0]) ||
				next[1] == curr[1] {
				continue
			}

			// Find where the line intersects...
			xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
			if curr[0] != next[0] && p[0] >= xint {
				continue
			}

			intersections++
		}
	}

	return intersections%2 != 0
}

// gets the tiles on a line
func GetTilesLine(line [][]float64, zoom int) []m.TileID {
	firsttile := m.Tile(line[0][0], line[0][1], zoom)
	bds := m.Bounds(firsttile)

	deltax := bds.E - bds.W
	deltay := bds.N - bds.S
	oldpt := line[0]
	tiles := []m.TileID{firsttile}
	tilesmap := map[m.TileID]string{firsttile: ""}
	for _, pt := range line[1:] {
		if math.Abs(pt[0]-oldpt[0]) > deltax {
			fmt.Println("here")
		}
		if math.Abs(pt[1]-oldpt[1]) > deltay {
			fmt.Println("here")
		}
		currenttile := m.Tile(pt[0], pt[1], zoom)
		_, boolval := tilesmap[currenttile]
		if !boolval {
			tilesmap[currenttile] = ""
			tiles = append(tiles, currenttile)
		}

		oldpt = pt
	}

	return tiles
}

// BoundingBox implementation as per https://tools.ietf.org/html/rfc7946
// BoundingBox syntax: "bbox": [west, south, east, north]
// BoundingBox defaults "bbox": [-180.0, -90.0, 180.0, 90.0]
func BoundingBoxPoints(pts [][]float64) m.Extrema {
	// setting opposite default values
	west, south, east, north := 180.0, 90.0, -180.0, -90.0

	for _, pt := range pts {
		x, y := pt[0], pt[1]
		// can only be one condition
		// using else if reduces one comparison
		if x < west {
			west = x
		} else if x > east {
			east = x
		}

		if y < south {
			south = y
		} else if y > north {
			north = y
		}
	}
	return m.Extrema{N: north, S: south, E: east, W: west}
}

//
func GetDif(x1, x2 int64) []int64 {
	current := x1
	newlist := []int64{current}
	for current < x2 {
		current++
		newlist = append(newlist, current)
	}
	return newlist
}

// gets the tiles of a single polygon
func GetTilesPolygon(polygon [][][]float64, zoom int) []m.TileID {
	bds := BoundingBoxPoints(polygon[0])
	wn := []float64{bds.W, bds.N}
	es := []float64{bds.E, bds.S}

	tilemin := m.Tile(wn[0], wn[1], zoom)
	tilemax := m.Tile(es[0], es[1], zoom)

	xs := GetDif(tilemin.X, tilemax.X)
	ys := GetDif(tilemin.Y, tilemax.Y)
	pottiles := []m.TileID{}
	for _, x := range xs {
		for _, y := range ys {
			pottiles = append(pottiles, m.TileID{x, y, uint64(zoom)})
		}
	}

	// checks the potential tiles
	tiles := []m.TileID{}
	for _, i := range pottiles {
		if TilePolygon(polygon, bds, i) {
			tiles = append(tiles, i)
		}
	}

	return tiles
}

// creates a map
func CreateMap(tiles []m.TileID) map[m.TileID]string {
	mymap := map[m.TileID]string{}
	for _, i := range tiles {
		mymap[i] = ""
	}
	return mymap
}

func TileCover(feature *geojson.Feature, zoom int) []m.TileID {
	mymap := map[m.TileID]string{}
	total := []m.TileID{}
	switch feature.Geometry.Type {
	case "Point":
		return []m.TileID{m.Tile(feature.Geometry.Point[0], feature.Geometry.Point[1], zoom)}
	case "LineString":
		return GetTilesLine(feature.Geometry.LineString, zoom)
	case "Polygon":
		return GetTilesPolygon(feature.Geometry.Polygon, zoom)
	case "MultiPoint":
		return GetTilesLine(feature.Geometry.LineString, zoom)
	case "MultiLineString":
		for pos, line := range feature.Geometry.MultiLineString {
			tmp := GetTilesLine(line, zoom)
			if pos == 0 {
				mymap = CreateMap(tmp)
				total = tmp
			} else {
				for _, tile := range tmp {
					_, boolval := mymap[tile]
					if !boolval {
						mymap[tile] = ""
						total = append(total, tile)
					}
				}
			}
		}
		return total
	case "MultiPolygon":
		/*
			for pos, polygon := range feature.Geometry.MultiPolygon {

				tmp := GetTilesPolygon(polygon, zoom)
				fmt.Println(tmp)
				if pos == 0 {
					fmt.Println("here1")
					mymap = CreateMap(tmp)
					total = tmp
				} else {
					fmt.Println("here2")
					for _, tile := range tmp {
						_, boolval := mymap[tile]
						if !boolval {
							mymap[tile] = ""
							total = append(total, tile)
							fmt.Println(len(total))
						}
					}
				}
			}
		*/
		return total
	}
	return total
}
