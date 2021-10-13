module github.com/openziti/fabric

go 1.17

//replace github.com/openziti/dilithium => ../dilithium

//replace github.com/openziti/foundation => ../foundation

replace go.etcd.io/bbolt => github.com/openziti/bbolt v1.3.6-0.20210317142109-547da822475e

require (
	github.com/AppsFlyer/go-sundheit v0.4.0
	github.com/biogo/store v0.0.0-20200525035639-8c94ae1e7c9c // indirect
	github.com/ef-ds/deque v1.0.4
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/michaelquigley/pfxlog v0.6.1
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/openziti/foundation v0.15.83
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.etcd.io/bbolt v1.3.5-0.20200615073812-232d8fc87f50
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/antlr/antlr4 v0.0.0-20210114010855-d34d2e1c271a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d // indirect
	github.com/kataras/go-events v0.0.3-0.20201007151548-c411dc70c0a6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/speps/go-hashids v2.0.0+incompatible // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
