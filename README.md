# go-whosonfirst-rasterzen

Go package for working with Who's On First data and `go-rasterzen` tiles.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

## Important

This is work in progress. It works. Until it doesn't.

## Tools

### wof-rasterzen-seed

Pre-seed [go-rasterzen](https://github.com/whosonfirst/go-rasterzen) tiles (and their vector/raster derivatives) for one or more [go-whosonfirst-index](https://github.com/whosonfirst/go-whosonfirst-index) compatible indices.

```
./bin/wof-rasterzen-seed -h
Usage of ./bin/wof-rasterzen-seed:
  -count
    	
  -fs-cache
    	Cache tiles with a filesystem-based cache.
  -fs-root string
    	The root of your filesystem cache. If empty rasterd will try to use the current working directory.
  -go-cache
    	Cache tiles with an in-memory (go-cache) cache.
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
...time passes...
```

Or if you just want to know how many tiles you'll end up fetching:

```
./bin/wof-rasterzen-seed -fs-cache -fs-root cache -nextzen-apikey {APIKEY} -max-zoom 10 -count -mode repo /usr/local/data/sfomuseum-data-whosonfirst/
164993
```

## See also

* https://github.com/whosonfirst/go-rasterzen
* https://github.com/whosonfirst/go-whosonfirst-index
* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://developers.nextzen.org/