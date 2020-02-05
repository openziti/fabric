module github.com/netfoundry/ziti-fabric

go 1.13

// replace github.com/netfoundry/ziti-foundation => ../ziti-foundation

require (
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/influxdata/influxdb1-client v0.0.0-20190809212627-fc22c7df067e
	github.com/michaelquigley/pfxlog v0.0.0-20190813191113-2be43bd0dccc
	github.com/netfoundry/ziti-foundation v0.6.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.2.2
	go.etcd.io/bbolt v1.3.3
	gopkg.in/yaml.v2 v2.2.7
)
