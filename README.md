# go-whosonfirst-rasterzen

Go package for working with Who's On First data and `go-rasterzen` tiles.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.11](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

## Important

This is work in progress. It works. Until it doesn't.

For background, you should read the following blog posts:

* [Maps (and map tiles) at SFO Museum](https://millsfield.sfomuseum.org/blog/2018/07/31/maps/)
* [Sweet spots between the extremes](https://millsfield.sfomuseum.org/blog/2018/11/07/rasterzen/)

## Tools

### wof-rasterzen-seed

Pre-seed [go-rasterzen](https://github.com/whosonfirst/go-rasterzen) tiles (and their vector/raster derivatives) for one or more [go-whosonfirst-index](https://github.com/whosonfirst/go-whosonfirst-index) compatible indices.

```
./bin/wof-rasterzen-seed -h
Usage of ./bin/wof-rasterzen-seed:
  -count
    	   Display the number of tiles to process and exit.
  -exclude value
    	   Exclude records not matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.    	
  -fs-cache
    	Cache tiles with a filesystem-based cache.
  -fs-root string
    	The root of your filesystem cache. If empty rasterd will try to use the current working directory.
  -go-cache
    	Cache tiles with an in-memory (go-cache) cache.
  -include value
    	   Include only those records matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the gjson package's 'dot' syntax.
  -max-zoom int
    	 (default 16)
  -min-zoom int
    	 (default 1)
  -mode string
    	 (default "repo")
  -nextzen-apikey string
    	A valid Nextzen API key.
  -nextzen-debug
    	Log requests (to STDOUT) to Nextzen tile servers.
  -nextzen-origin string
    	An optional HTTP 'Origin' host to pass along with your Nextzen requests.
  -nextzen-uri string
    	A valid URI template (RFC 6570) pointing to a custom Nextzen endpoint.
  -s3-cache
    	Cache tiles with a S3-based cache.
  -s3-dsn string
    	A valid go-whosonfirst-aws DSN string
  -s3-opts string
    	A valid go-whosonfirst-cache-s3 options string
  -seed-png
    	Seed PNG tiles.
  -seed-svg
    	Seed SVG tiles. (default true)
  -seed-timings
    	Display timings when seeding tiles.
  -seed-workers int
    	The maximum number of concurrent workers to invoke when seeding tiles (default 100)
```

For example:

```
./bin/wof-rasterzen-seed -fs-cache -fs-root cache -nextzen-apikey {APIKEY} -max-zoom 10 -mode repo /usr/local/data/sfomuseum-data-whosonfirst/
...time passes, your fan whirrrrrrrrrs...
```

Or if you just want to know how many tiles you'll end up fetching:

```
./bin/wof-rasterzen-seed -fs-cache -fs-root cache -nextzen-apikey {APIKEY} -max-zoom 10 -count -mode repo /usr/local/data/sfomuseum-data-whosonfirst/
164993
```

### Filtering records

You can filter records to process by passing one or more `-include` or `-exclude` flags. The `-exclude` flag will exclude records not matching one or path '{PATH}={VALUE}' statements. The `-include` flag will include only those records matching one or path '{PATH}={VALUE}' statements. Paths are evaluated using the `gjson` package's 'dot' syntax. For example:

```
./bin/wof-rasterzen-seed -count -max-zoom 10 -include 'properties.wof:placetype=country' /usr/local/data/sfomuseum-data-whosonfirst/
00:16:49.630981 [wof-rasterzen-seed] STATUS SKIP Dayton (101712161) because properties.wof:placetype != country (is locality)
00:16:49.727171 [wof-rasterzen-seed] STATUS SKIP Columbus (101712381) because properties.wof:placetype != country (is locality)
00:16:49.741976 [wof-rasterzen-seed] STATUS SKIP Middletown (101712653) because properties.wof:placetype != country (is locality)
00:16:49.749884 [wof-rasterzen-seed] STATUS SKIP Portland (101715829) because properties.wof:placetype != country (is locality)
00:16:49.765511 [wof-rasterzen-seed] STATUS SKIP McMinnville (101716117) because properties.wof:placetype != country (is locality)
...and so on
...time passes
00:18:37.793127 [wof-rasterzen-seed] STATUS SKIP Mills Field Municipal Airport of San Francisco (1360695653) because properties.wof:placetype != country (is campus)
00:18:37.804110 [wof-rasterzen-seed] STATUS SKIP San Francisco Airport (1360695655) because properties.wof:placetype != country (is campus)
00:18:37.815119 [wof-rasterzen-seed] STATUS SKIP Denver Stapleton International Airport (1360695657) because properties.wof:placetype != country (is campus)
152956
```

## See also

* https://github.com/whosonfirst/go-rasterzen
* https://github.com/whosonfirst/go-whosonfirst-index
* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://developers.nextzen.org/