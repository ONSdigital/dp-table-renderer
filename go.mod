module github.com/ONSdigital/dp-table-renderer

go 1.24

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/ONSdigital/dp-healthcheck v1.6.4
	github.com/ONSdigital/dp-net/v3 v3.3.0
	github.com/ONSdigital/dp-otel-go v0.0.7
	github.com/ONSdigital/log.go/v2 v2.4.5
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/smartystreets/goconvey v1.8.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0
	go.opentelemetry.io/otel v1.35.0
	golang.org/x/net v0.39.0
)

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.266.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.22.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.22.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.22.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.22.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.22.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

// Override for vulnerable indirect go-jose/v4@4.0.4 CVE-2025-27144
tool github.com/go-jose/go-jose/v4
