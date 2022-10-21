package router

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/openziti/channel/v2"
	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/tls"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/stretchr/testify/assert"
)

func Test_initializeCtrlEndpoints_ErrorsWithoutDataDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	r := Router{
		config: &Config{},
	}
	err = r.initializeCtrlEndpoints()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "ctrl DataDir not configured")
}

func Test_initializeCtrlEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	transport.AddAddressParser(tls.AddressParser{})
	addr, err := transport.ParseAddress("tls:localhost:6565")
	if err != nil {
		t.Fatal(err)
	}
	r := Router{
		config: &Config{
			Ctrl: struct {
				InitialEndpoints      []*UpdatableAddress
				LocalBinding          string
				DefaultRequestTimeout time.Duration
				Options               *channel.Options
				DataDir               string
			}{
				DataDir:          tmpDir,
				InitialEndpoints: []*UpdatableAddress{NewUpdatableAddress(addr)},
			},
		},
		ctrlEndpoints: cmap.New[*UpdatableAddress](),
	}

	assert.NoError(t, r.initializeCtrlEndpoints())
	assert.FileExists(t, path.Join(tmpDir, "endpoints"))

	b, err := os.ReadFile(path.Join(tmpDir, "endpoints"))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	found := []*UpdatableAddress{}
	eps := strings.Split(string(b), "\n")
	for _, ep := range eps {
		parsed, err := transport.ParseAddress(ep)
		assert.NoError(t, err)
		found = append(found, NewUpdatableAddress(parsed))
	}

	assert.Equal(t, r.config.Ctrl.InitialEndpoints, found)
}

func Test_UpdateCtrlEndpoints(t *testing.T) {}
