module gallo

go 1.14

require (
	github.com/adlio/trello v1.7.0
	github.com/go-redis/cache/v8 v8.3.0
	github.com/go-redis/redis/v8 v8.5.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/jarcoal/httpmock v1.0.5
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	gotest.tools v2.2.0+incompatible
)

replace github.com/adlio/trello v1.7.0 => github.com/rhardih/trello v1.7.1
