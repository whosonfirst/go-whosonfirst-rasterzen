# tile-cover

Keep it simple stupid. Given a geojson feature, and a zoom return all the tiles that have an intersection. The old one tried to determine the zoom for you this one skips that. 

# Usage 
```go
package main

import (
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"fmt"
	"github.com/murphy214/tile-cover"
	m "github.com/murphy214/mercantile"

)

func main() {
	bytevals,_ := ioutil.ReadFile("states.geojson")
	fc, _ := geojson.UnmarshalFeatureCollection(bytevals)
	feat := fc.Features[20]

	tileids := tile_cover.TileCover(feat,10)
	for _,i := range tileids {
		fmt.Println(m.Tilestr(i))
	}
}
```

# Output 
![](https://user-images.githubusercontent.com/10904982/35519140-6876e2a0-04e1-11e8-9348-d87eb4614dcd.png)
