/*
	(c) Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package router

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/fabric/controller/xctrl"
	"github.com/openziti/fabric/pb/ctrl_pb"
	"github.com/openziti/fabric/router/forwarder"
	"github.com/openziti/fabric/router/handler_ctrl"
	"github.com/openziti/fabric/router/handler_link"
	"github.com/openziti/fabric/router/handler_xgress"
	"github.com/openziti/fabric/router/xgress"
	"github.com/openziti/fabric/router/xgress_proxy"
	"github.com/openziti/fabric/router/xgress_proxy_udp"
	"github.com/openziti/fabric/router/xgress_transport"
	"github.com/openziti/fabric/router/xgress_transport_udp"
	"github.com/openziti/fabric/router/xlink"
	"github.com/openziti/fabric/router/xlink_transport"
	"github.com/openziti/foundation/channel2"
	"github.com/openziti/foundation/common"
	"github.com/openziti/foundation/event"
	"github.com/openziti/foundation/events"
	"github.com/openziti/foundation/metrics"
	"github.com/openziti/foundation/profiler"
	"github.com/openziti/foundation/util/concurrenz"
	"github.com/openziti/foundation/util/info"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"time"
)

type Router struct {
	config          *Config
	ctrl            channel2.Channel
	ctrlOptions     *channel2.Options
	linkOptions     *channel2.Options
	linkListener    channel2.UnderlayListener
	faulter         *forwarder.Faulter
	forwarder       *forwarder.Forwarder
	xctrls          []xctrl.Xctrl
	xctrlDone       chan struct{}
	xlinkFactories  map[string]xlink.Factory
	xlinkListeners  []xlink.Listener
	xlinkDialers    []xlink.Dialer
	xgressListeners []xgress.Listener
	metricsRegistry metrics.UsageRegistry
	shutdownC       chan struct{}
	isShutdown      concurrenz.AtomicBoolean
	eventDispatcher event.Dispatcher
	metricsReporter metrics.Handler
	versionProvider common.VersionProvider
}

func (self *Router) MetricsRegistry() metrics.UsageRegistry {
	return self.metricsRegistry
}

func (self *Router) Channel() channel2.Channel {
	return self.ctrl
}

func (self *Router) DefaultRequestTimeout() time.Duration {
	return self.config.Ctrl.DefaultRequestTimeout
}

func Create(config *Config, versionProvider common.VersionProvider) *Router {
	closeNotify := make(chan struct{})

	eventDispatcher := event.NewDispatcher(closeNotify)
	metricsRegistry := metrics.NewUsageRegistry(config.Id.Token, map[string]string{}, closeNotify)
	xgress.InitMetrics(metricsRegistry)

	faulter := forwarder.NewFaulter(config.Forwarder.FaultTxInterval, closeNotify)
	fwd := forwarder.NewForwarder(metricsRegistry, faulter, config.Forwarder, closeNotify)

	xgress.InitPayloadIngester(closeNotify)
	xgress.InitAcker(fwd, metricsRegistry, closeNotify)
	xgress.InitRetransmitter(fwd, fwd, metricsRegistry, closeNotify)

	return &Router{
		config:          config,
		faulter:         faulter,
		forwarder:       fwd,
		metricsRegistry: metricsRegistry,
		shutdownC:       closeNotify,
		eventDispatcher: eventDispatcher,
		versionProvider: versionProvider,
	}
}

func (self *Router) RegisterXctrl(x xctrl.Xctrl) error {
	if err := self.config.Configure(x); err != nil {
		return err
	}
	if x.Enabled() {
		self.xctrls = append(self.xctrls, x)
	}
	return nil
}

func (self *Router) Start() error {
	rand.Seed(info.NowInMilliseconds())

	self.showOptions()

	self.startProfiling()

	if err := self.registerComponents(); err != nil {
		return err
	}

	self.startXlinkDialers()
	self.startXlinkListeners()
	self.startXgressListeners()

	if err := self.startControlPlane(); err != nil {
		return err
	}
	return nil
}

func (self *Router) Shutdown() error {
	var errors []error
	if self.isShutdown.CompareAndSwap(false, true) {
		if self.metricsReporter != nil {
			events.RemoveMetricsEventHandler(self.metricsReporter)
		}

		if err := self.ctrl.Close(); err != nil {
			errors = append(errors, err)
		}

		close(self.shutdownC)

		for _, xlinkListener := range self.xlinkListeners {
			if err := xlinkListener.Close(); err != nil {
				errors = append(errors, err)
			}
		}

		for _, xgressListener := range self.xgressListeners {
			if err := xgressListener.Close(); err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}
	return network.MultipleErrors(errors)
}

func (self *Router) Run() error {
	if err := self.Start(); err != nil {
		return err
	}
	for {
		time.Sleep(1 * time.Hour)
	}
}

func (self *Router) showOptions() {
	if output, err := json.Marshal(self.config.Ctrl.Options); err == nil {
		pfxlog.Logger().Infof("ctrl = %s", string(output))
	} else {
		logrus.Fatalf("unable to display options (%v)", err)
	}

	if output, err := json.Marshal(self.config.Metrics); err == nil {
		pfxlog.Logger().Infof("metrics = %s", string(output))
	} else {
		logrus.Fatalf("unable to display options (%v)", err)
	}
}

func (self *Router) startProfiling() {
	if self.config.Profile.Memory.Path != "" {
		go profiler.NewMemoryWithShutdown(self.config.Profile.Memory.Path, self.config.Profile.Memory.Interval, self.shutdownC).Run()
	}
	if self.config.Profile.CPU.Path != "" {
		if cpu, err := profiler.NewCPUWithShutdown(self.config.Profile.CPU.Path, self.shutdownC); err == nil {
			go cpu.Run()
		} else {
			logrus.Errorf("unexpected error launching cpu profiling (%v)", err)
		}
	}
	go newRouterMonitor(self.forwarder, self.shutdownC).Monitor()
}

func (self *Router) registerComponents() error {
	self.xlinkFactories = make(map[string]xlink.Factory)
	xlinkAccepter := newXlinkAccepter(self.forwarder)
	xlinkChAccepter := handler_link.NewChannelAccepter(self,
		self.forwarder,
		self.config.Forwarder,
		self.metricsRegistry,
	)
	self.xlinkFactories["transport"] = xlink_transport.NewFactory(xlinkAccepter, xlinkChAccepter, self.config.Transport)

	xgress.GlobalRegistry().Register("proxy", xgress_proxy.NewFactory(self.config.Id, self, self.config.Transport))
	xgress.GlobalRegistry().Register("proxy_udp", xgress_proxy_udp.NewFactory(self))
	xgress.GlobalRegistry().Register("transport", xgress_transport.NewFactory(self.config.Id, self, self.config.Transport))
	xgress.GlobalRegistry().Register("transport_udp", xgress_transport_udp.NewFactory(self.config.Id, self))

	return nil
}

func (self *Router) startXlinkDialers() {
	for _, lmap := range self.config.Link.Dialers {
		binding := lmap["binding"].(string)
		if factory, found := self.xlinkFactories[binding]; found {
			dialer, err := factory.CreateDialer(self.config.Id, self.forwarder, lmap)
			if err != nil {
				logrus.Fatalf("error creating Xlink dialer (%v)", err)
			}
			self.xlinkDialers = append(self.xlinkDialers, dialer)
			logrus.Infof("started Xlink dialer with binding [%s]", binding)
		}
	}
}

func (self *Router) startXlinkListeners() {
	for _, lmap := range self.config.Link.Listeners {
		binding := lmap["binding"].(string)
		if factory, found := self.xlinkFactories[binding]; found {
			listener, err := factory.CreateListener(self.config.Id, self.forwarder, lmap)
			if err != nil {
				logrus.Fatalf("error creating Xlink listener (%v)", err)
			}
			if err := listener.Listen(); err != nil {
				logrus.Fatalf("error listening on Xlink (%v)", err)
			}
			self.xlinkListeners = append(self.xlinkListeners, listener)
			logrus.Infof("started Xlink listener with binding [%s] advertising [%s]", binding, listener.GetAdvertisement())
		}
	}
}

func (self *Router) startXgressListeners() {
	for _, binding := range self.config.Listeners {
		factory, err := xgress.GlobalRegistry().Factory(binding.name)
		if err != nil {
			logrus.Fatalf("error getting xgress factory [%s] (%v)", binding.name, err)
		}
		listener, err := factory.CreateListener(binding.options)
		if err != nil {
			logrus.Fatalf("error creating xgress listener [%s] (%v)", binding.name, err)
		}
		self.xgressListeners = append(self.xgressListeners, listener)

		var address string
		if addressVal, found := binding.options["address"]; found {
			address = addressVal.(string)
		}

		err = listener.Listen(address,
			handler_xgress.NewBindHandler(
				handler_xgress.NewReceiveHandler(self.forwarder),
				handler_xgress.NewCloseHandler(self, self.forwarder),
				self.forwarder,
			),
		)
		if err != nil {
			logrus.Fatalf("error listening [%s] (%v)", binding.name, err)
		}
		logrus.Infof("created xgress listener [%s] at [%s]", binding.name, address)
	}
}

func (self *Router) startControlPlane() error {
	attributes := map[int32][]byte{}

	version, err := self.versionProvider.EncoderDecoder().Encode(self.versionProvider.AsVersionInfo())

	if err != nil {
		return fmt.Errorf("error with version header information value: %v", err)
	}

	attributes[channel2.HelloVersionHeader] = version

	if len(self.xlinkListeners) == 1 {
		attributes[channel2.HelloRouterAdvertisementsHeader] = []byte(self.xlinkListeners[0].GetAdvertisement())
	}

	reconnectHandler := func() {
		for _, x := range self.xctrls {
			x.NotifyOfReconnect()
		}
	}

	dialer := channel2.NewReconnectingDialerWithHandler(self.config.Id, self.config.Ctrl.Endpoint, attributes, reconnectHandler)

	bindHandler := handler_ctrl.NewBindHandler(
		self.config.Id,
		self.config.Dialers,
		self.xlinkDialers,
		self,
		self.forwarder,
		self.xctrls,
		self.shutdownC,
	)

	self.config.Ctrl.Options.BindHandlers = append(self.config.Ctrl.Options.BindHandlers, bindHandler)

	ch, err := channel2.NewChannel("ctrl", dialer, self.config.Ctrl.Options)
	if err != nil {
		return fmt.Errorf("error connecting ctrl (%v)", err)
	}

	self.ctrl = ch
	self.faulter.SetCtrl(ch)

	self.xctrlDone = make(chan struct{})
	for _, x := range self.xctrls {
		if err := x.Run(self.ctrl, nil, self.xctrlDone); err != nil {
			return err
		}
	}

	self.metricsReporter = metrics.NewChannelReporter(self.ctrl)
	self.metricsRegistry.StartReporting(self.metricsReporter, self.config.Metrics.ReportInterval, self.config.Metrics.MessageQueueSize)

	return nil
}

const (
	DumpForwarderTables byte = 1
	UpdateRoute         byte = 2
	CloseControlChannel byte = 3
	OpenControlChannel  byte = 4
)

func (router *Router) HandleDebug(conn io.ReadWriter) error {
	bconn := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	op, err := bconn.ReadByte()
	if err != nil {
		return err
	}

	switch op {
	case DumpForwarderTables:
		tables := router.forwarder.Debug()
		_, err := conn.Write([]byte(tables))
		return err
	case UpdateRoute:
		logrus.Error("received debug operation to update routes")
		sizeBuf := make([]byte, 4)
		if _, err := bconn.Read(sizeBuf); err != nil {
			return err
		}
		size := binary.LittleEndian.Uint32(sizeBuf)
		messageBuf := make([]byte, size)

		if _, err := bconn.Read(messageBuf); err != nil {
			return err
		}

		route := &ctrl_pb.Route{}
		if err := proto.Unmarshal(messageBuf, route); err != nil {
			return err
		}

		logrus.Errorf("updating with route: %+v", route)
		logrus.Errorf("updating with route: %v", route)

		router.forwarder.Route(route)
	case CloseControlChannel:
		logrus.Warn("control channel: closing")
		_, _ = bconn.WriteString("control channel: closing\n")
		if togglable, ok := router.ctrl.Underlay().(connectionToggle); ok {
			if err := togglable.Disconnect(); err != nil {
				logrus.WithError(err).Error("control channel: failed to close")
				_, _ = bconn.WriteString(fmt.Sprintf("control channel: failed to close (%v)\n", err))
			} else {
				logrus.Warn("control channel: closed")
				_, _ = bconn.WriteString("control channel: closed")
			}
		} else {
			logrus.Warn("control channel: error not toggleable")
			_, _ = bconn.WriteString("control channel: error not toggleable")
		}
	case OpenControlChannel:
		logrus.Warn("control channel: reconnecting")
		if togglable, ok := router.ctrl.Underlay().(connectionToggle); ok {
			if err := togglable.Reconnect(); err != nil {
				logrus.WithError(err).Error("control channel: failed to reconnect")
				_, _ = bconn.WriteString(fmt.Sprintf("control channel: failed to reconnect (%v)\n", err))
			} else {
				logrus.Warn("control channel: reconnected")
				_, _ = bconn.WriteString("control channel: reconnected")
			}
		} else {
			logrus.Warn("control channel: error not toggleable")
			_, _ = bconn.WriteString("control channel: error not toggleable")
		}
	}

	return nil
}

type connectionToggle interface {
	Disconnect() error
	Reconnect() error
}
