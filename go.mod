module gallo

go 1.14

require (
	github.com/adlio/trello v1.7.0
	github.com/go-redis/cache v6.4.0+incompatible
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/go-redis/redis/v8 v8.4.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/jarcoal/httpmock v1.0.5
	github.com/kr/pretty v0.1.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/adlio/trello v1.7.0 => github.com/rhardih/trello v1.7.1
