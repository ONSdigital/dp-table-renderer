dp-table-renderer
================

Given json defining a table, capable of rendering a table in multiple formats

### Getting started


| Environment variable       | Default                  | Description                                            |
| -------------------------- | ------------------------ | -----------                                            |
| BIND_ADDR                  | :23300                   | The host and port to bind to                           |
| HEALTH_CHECK_INTERVAL           | Interval between health checks                                                            |    30 seconds |
| HEALTH_CHECK_CRITICAL_TIMEOUT    | Amount of time to pass since last healthy health check to be deemed a critical failure    |    90 seconds |
| CORS_ALLOWED_ORIGINS       | *                        | The allowed origins for CORS requests                  |
| SHUTDOWN_TIMEOUT           | 5s                       | The graceful shutdown timeout ([`time.Duration`](https://golang.org/pkg/time/#Duration) format) |
| OTEL_EXPORTER_OTLP_ENDPOINT      | localhost:4317                            | Host and port for the OpenTelemetry endpoint                                             |
| OTEL_SERVICE_NAME                | dp-table-renderer                         | Service name to report to telemetry tools                                                |
| OTEL_BATCH_TIMEOUT               | 5s                                        | Interval between pushes to OT Collector                                                  |
| OTEL_ENABLED                     | false                                     | Feature flag to enable OpenTelemetry

### Endpoints

| url                   | Method | Parameter values                       | Description                                                                                   |
| ---                   | ------ | ----------------                       | -----------                                                                                   |
| /render/{render_type} | POST   | render_type = `html`, `csv`, or `xlsx` | Renders the (json) data provided in the post body as a table in the requested format          |
| /parse/html           | POST   |                                        | Parses an html table and returns the json format suitable for sending to the /render endpoint |

See the [swagger.yaml](swagger.yaml) file for a full definition (use http://editor.swagger.io to make it easy to read),
and see the json files in the testdata directory for example requests.

#### /render/{render_type}

Merged cells can be specified using `colspan` and `rowspan` properties of `cell_format` elements.
Please note that the `data` array should include *all* cells (i.e. each row should contain the same number of cells), even if some of them have been merged. This is the same approach/format used by some javascript spreadsheet components such as [Handsontable](https://handsontable.com/).

#### /parse/html

Please note that the is assumed to include *all* cells (i.e. each row should contain the same number of cells), even if some of them have been hidden by merged cells. This is the same approach/format used by some javascript spreadsheet components such as [Handsontable](https://handsontable.com/).
The response contains the html generated by /render/html as well as the json required to call that endpoint.

### Healthchecking

Currently, reported on endpoint `/healthcheck`. There are no other services consumed, so it will always return OK.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2018-2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
