module github.com/openziti/fabric

go 1.15

//replace github.com/openziti/foundation => ../foundation

require (
	github.com/biogo/store v0.0.0-20200525035639-8c94ae1e7c9c // indirect
	github.com/ef-ds/deque v1.0.4
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/michaelquigley/pfxlog v0.3.7
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/openziti/foundation v0.15.30
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.etcd.io/bbolt v1.3.5-0.20200615073812-232d8fc87f50
	google.golang.org/protobuf v1.23.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
