dp-table-renderer
================

Given json defining a table, capable of rendering a table in multiple formats

### Getting started


| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | :23100                                    | The host and port to bind to
| HOST                       | http://localhost:23100                    | The host name used to build URLs
| SHUTDOWN_TIMEOUT           | 5s                                        | The graceful shutdown timeout (`time.Duration` format)

### Endpoints

| url          | Method | Description                                                                      |
| ---          | ------ | -----------                                                                      |
| /render/html | POST   | Renders the (json) data provided in the post body as a self-contained html table |

### Healthchecking

Currently reported on endpoint `/healthcheck`. There are no other services consumed, so it will always return OK.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2018, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
