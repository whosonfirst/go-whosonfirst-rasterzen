# go-whosonfirst-rasterzen

Go package for working with Who's On First data and `go-rasterzen` tiles.

## Install

You will need to have both `Go` (specifically version [1.12](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This is work in progress. It works. Until it doesn't.

For background, you should read the following blog posts:

* [Maps (and map tiles) at SFO Museum](https://millsfield.sfomuseum.org/blog/2018/07/31/maps/)
* [Sweet spots between the extremes](https://millsfield.sfomuseum.org/blog/2018/11/07/rasterzen/)

## Tools

### wof-rasterzen-seed

Pre-seed [go-rasterzen](https://github.com/whosonfirst/go-rasterzen) tiles (and their vector/raster derivatives) for one or more [go-whosonfirst-index](https://github.com/whosonfirst/go-whosonfirst-index) compatible indices.

```
> go run cmd/wof-rasterzen-seed/main.go -h
Usage of /var/folders/_k/h7ndzcyx3dq027gsrg1q45xm0000gn/T/go-build375480020/b001/exe/main:
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
  -lambda-dsn value
    	A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION'
  -lambda-function string
    	A valid AWS Lambda function name. (default "Rasterzen")
  -log-level string
    	Log level to use for logging (default "status")
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
  -png-options string
    	The path to a valid RasterzenPNGOptions JSON file.
  -rasterzen-options string
    	The path to a valid RasterzenOptions JSON file.
  -refresh-all
    	Force all tiles to be generated even if they are already cached.
  -refresh-png
    	Force PNG tiles to be generated even if they are already cached.
  -refresh-rasterzen
    	Force rasterzen tiles to be generated even if they are already cached.
  -refresh-svg
    	Force SVG tiles to be generated even if they are already cached.
  -s3-cache
    	Cache tiles with a S3-based cache.
  -s3-dsn string
    	A valid go-whosonfirst-aws DSN string
  -s3-opts string
    	A valid go-whosonfirst-cache-s3 options string
  -seed-all
    	See all the tile formats
  -seed-max-workers runtime.NumCPU()
    	The maximum number of concurrent workers to invoke when seeding tiles. The default is the value of runtime.NumCPU() * 2.
  -seed-png
    	Seed PNG tiles.
  -seed-rasterzen
    	Seed Rasterzen tiles.
  -seed-svg
    	Seed SVG tiles.
  -seed-tileset-catalog-dsn string
    	A valid tile.SeedCatalog DSN string. Required parameters are 'catalog=CATALOG' (default "catalog=sync")
  -seed-worker string
    	The type of worker for seeding tiles. Valid workers are: lambda, local, sqs. (default "local")
  -sqs-dsn value
    	A valid go-whosonfirst-aws DSN string. Required paremeters are 'credentials=CREDENTIALS' and 'region=REGION' and 'queue=QUEUE'
  -svg-options string
    	The path to a valid RasterzenSVGOptions JSON file.
  -timings
    	Display timings for tile seeding.
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

## Docker

Yes. There is a [Dockerfile](Dockerfile) for the `wof-rasterzen-seed` tool.

### ECS (AWS)

The detailed details of using the above mentioned Dockerfile with the AWS ECS service are outside the scope of this document but your container will need an IAM role with the following (minimum) built-in policies:

* `AmazonEC2ContainerServiceforEC2Role`

Additionally, if you plan to use the SQS `seed-worker` you'll need to add a policy for sending messages to the SQS queue in question. For example:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:SendMessage",
                "sqs:GetQueueUrl",
                "sqs:GetQueueAttributes"
            ],
            "Resource": [
                "arn:aws:sqs:{AWS_REGION}:{AWS_ACCOUNT_ID}:{SQS_QUEUE}"
            ]
        }
    ]
}
```

If you are going to use the Lambda `seed-worker` you'll need to add a policy for invoking the Lambda function in question. For example:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowLambdaRasterzenSeeder",
            "Effect": "Allow",
            "Action": "lambda:InvokeFunction",
            "Resource": "arn:aws:lambda:{AWS_REGION}:{AWS_ACCOUNT_ID}:function:{LAMBDA_FUNCTION}"
        }
    ]
}
```

Your role should have the following trust relationship:

```
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Finally, assuming you haven't baked a specific `wof-rasterzen-seed` command in to your container image you would invoke it as a (ECS) task with a "container override" like this:

```
/usr/local/bin/wof-rasterzen-seed,-seed-max-workers,25,-nextzen-apikey,{NEXTZEN_APIKEY},-seed-worker,sqs,-sqs-dsn,credentials=iam: region={AWS_REGION} queue={SQS_QUEUE},-seed-all,-min-zoom,14,-max-zoom,16,-timings,-mode,git,https://github.com/sfomuseum-data/sfomuseum-data-architecture.git
```

_See the commas between everything and the lack of quotes around the `-sqs-dsn` flag? I love that about ECS..._

The command above will fetch a copy of the `sfomuseum-data-architecture` repo, calculate all the tiles between zoom levels 14 and 16 (for all the WOF records in the repo) and then "fetch" each one of those tiles using the SQS `seed-worker` which just means an entry for the tile will be created in an SQS queue and assumes you've configured the `go-rasterzen` [rasterzen-seed-sqs](https://github.com/whosonfirst/go-rasterzen/blob/master/cmd/rasterzen-seed-sqs/main.go) tool as a Lambda trigger for new queue items.

For a pretty picture of everything just described, see [go-rasterzen/docs/rasterzen-seed-sqs-arch.jpg ](https://github.com/whosonfirst/go-rasterzen/blob/master/docs/rasterzen-seed-sqs-arch.jpg). (Note that the image assumes the `go-rasterzen` [rasterzen-seed](https://github.com/whosonfirst/go-rasterzen#rasterzen-seed) tool which doesn't know how to seed WOF repos, but the principle is the same.)

Because this (`wof-rasterzen-seed`) is being run from an ECS instance you could also invoke the Lambda `seed-worker` directly rather than queueing everything up in SQS. That's your business. The point is you can do either. In both cases though it's assumed that there is a Lambda `seed-worker` that is fetching data from a copy of the `go-rasterzen` [rasterd](https://github.com/whosonfirst/go-rasterzen#rasterd) application.

See also: [go-rasterzen workers](https://github.com/whosonfirst/go-rasterzen/tree/master/worker).

## See also

* https://github.com/whosonfirst/go-rasterzen
* https://github.com/whosonfirst/go-whosonfirst-index
* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://developers.nextzen.org/