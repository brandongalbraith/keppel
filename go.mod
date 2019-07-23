module github.com/sapcc/keppel

go 1.12

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d
	github.com/beorn7/perks v1.0.0
	github.com/bshuster-repo/logrus-logstash-hook v0.4.1
	github.com/bugsnag/bugsnag-go v0.0.0-20141110184014-b1d153021fcd
	github.com/bugsnag/osext v0.0.0-20130617224835-0dd3f918b21b
	github.com/bugsnag/panicwrap v0.0.0-20151223152923-e2c28503fcd0
	github.com/databus23/goslo.policy v0.0.0-20170317131957-3ae74dd07ebf
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/libtrust v0.0.0-20150114040149-fa567046d9b1
	github.com/garyburd/redigo v0.0.0-20150301180006-535138d7bcd7
	github.com/golang-migrate/migrate v3.4.0+incompatible
	github.com/golang/protobuf v1.3.2
	github.com/gophercloud/gophercloud v0.0.0-20180708220030-45c2d035713f
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v0.0.0-20150720190736-60c7bfde3e33
	github.com/gorilla/mux v0.0.0-20170228224354-599cba5e7b61
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/jpillora/longestcommon v0.0.0-20161227235612-adb9d91ee629
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/lib/pq v1.0.0
	github.com/majewsky/schwift v0.0.0-20180906125654-e1b3d5e2efc9
	github.com/mattes/migrate v0.0.0-20171024180000-69472d5f5cdc
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/dns v0.0.0-20161122061214-271c58e0c14f
	github.com/mitchellh/mapstructure v0.0.0-20150528213339-482a9fd5fa83
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.6.0
	github.com/prometheus/procfs v0.0.3
	github.com/sapcc/go-bits v0.0.0-20190522121402-dabab492e20b
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.0-20150605180824-312092086bed
	github.com/spf13/pflag v0.0.0-20150601220040-564482062245
	github.com/xenolf/lego v0.0.0-20160613233155-a9d8cec0e656
	github.com/yvasiyarov/go-metrics v0.0.0-20140926110328-57bccd1ccd43
	github.com/yvasiyarov/gorelic v0.0.0-20141212073537-a9bba5b9ab50
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20140908184405-b21fdbd4370f
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7
	golang.org/x/time v0.0.0-20160202183820-a4bde1265759
	gopkg.in/gorp.v2 v2.0.0-20180226155812-4df78490a9aa
	gopkg.in/square/go-jose.v1 v1.0.1
	gopkg.in/yaml.v2 v2.2.1
	rsc.io/letsencrypt v0.0.0-00010101000000-000000000000 // indirect
)

replace rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
