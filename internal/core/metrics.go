package core

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/aler9/rtsp-simple-server/internal/logger"
)

func formatMetric(key string, value int64) string {
	return key + " " + strconv.FormatInt(value, 10) + "\n"
}

type metricsPathManager interface {
	OnAPIPathsList(req apiPathsListReq1) apiPathsListRes1
}

type metricsRTSPServer interface {
	OnAPIRTSPSessionsList(req apiRTSPSessionsListReq) apiRTSPSessionsListRes
}

type metricsRTMPServer interface {
	OnAPIRTMPConnsList(req apiRTMPConnsListReq) apiRTMPConnsListRes
}

type metricsParent interface {
	Log(logger.Level, string, ...interface{})
}

type metrics struct {
	listener net.Listener
	mux      *http.ServeMux
	server   *http.Server

	mutex       sync.Mutex
	pathManager metricsPathManager
	rtspServer  metricsRTSPServer
	rtspsServer metricsRTSPServer
	rtmpServer  metricsRTMPServer
}

func newMetrics(
	address string,
	parent metricsParent,
) (*metrics, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	m := &metrics{
		listener: listener,
	}

	m.mux = http.NewServeMux()
	m.mux.HandleFunc("/metrics", m.onMetrics)

	m.server = &http.Server{
		Handler: m.mux,
	}

	parent.Log(logger.Info, "[metrics] opened on "+address)

	go m.run()

	return m, nil
}

func (m *metrics) close() {
	m.server.Shutdown(context.Background())
}

func (m *metrics) run() {
	err := m.server.Serve(m.listener)
	if err != http.ErrServerClosed {
		panic(err)
	}
}

func (m *metrics) onMetrics(w http.ResponseWriter, req *http.Request) {

	out := ""

	res := m.pathManager.OnAPIPathsList(apiPathsListReq1{})
	if res.Err == nil {

		for _, p := range res.Data.Items {
			if p.SourceReady {
				out += formatMetric("paths{name=\""+p.ConfName+"\", state=\"ready\"}", 1)
			} else {
				out += formatMetric("paths{name=\""+p.ConfName+"\", state=\"notReady\"}", 1)
			}
		}
	}

	if !interfaceIsEmpty(m.rtspServer) {
		res := m.rtspServer.OnAPIRTSPSessionsList(apiRTSPSessionsListReq{})
		if res.Err == nil {
			idleCount := int64(0)
			readCount := int64(0)
			publishCount := int64(0)

			for _, i := range res.Data.Items {
				switch i.State {
				case "idle":
					idleCount++
				case "read":
					readCount++
				case "publish":
					publishCount++
				}
			}

			out += formatMetric("rtsp_sessions{state=\"idle\"}",
				idleCount)
			out += formatMetric("rtsp_sessions{state=\"read\"}",
				readCount)
			out += formatMetric("rtsp_sessions{state=\"publish\"}",
				publishCount)
		}
	}

	if !interfaceIsEmpty(m.rtspsServer) {
		res := m.rtspsServer.OnAPIRTSPSessionsList(apiRTSPSessionsListReq{})
		if res.Err == nil {
			idleCount := int64(0)
			readCount := int64(0)
			publishCount := int64(0)

			for _, i := range res.Data.Items {
				switch i.State {
				case "idle":
					idleCount++
				case "read":
					readCount++
				case "publish":
					publishCount++
				}
			}

			out += formatMetric("rtsps_sessions{state=\"idle\"}",
				idleCount)
			out += formatMetric("rtsps_sessions{state=\"read\"}",
				readCount)
			out += formatMetric("rtsps_sessions{state=\"publish\"}",
				publishCount)
		}
	}

	if !interfaceIsEmpty(m.rtmpServer) {
		res := m.rtmpServer.OnAPIRTMPConnsList(apiRTMPConnsListReq{})
		if res.Err == nil {
			idleCount := int64(0)
			readCount := int64(0)
			publishCount := int64(0)

			for _, i := range res.Data.Items {
				switch i.State {
				case "idle":
					idleCount++
				case "read":
					readCount++
				case "publish":
					publishCount++
				}
			}

			out += formatMetric("rtmp_conns{state=\"idle\"}",
				idleCount)
			out += formatMetric("rtmp_conns{state=\"read\"}",
				readCount)
			out += formatMetric("rtmp_conns{state=\"publish\"}",
				publishCount)
		}
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, out)
}

// OnPathManagerSet is called by pathManager.
func (m *metrics) OnPathManagerSet(s metricsPathManager) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.pathManager = s
}

// OnRTSPServer is called by rtspServer (plain).
func (m *metrics) OnRTSPServerSet(s metricsRTSPServer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.rtspServer = s
}

// OnRTSPServer is called by rtspServer (plain).
func (m *metrics) OnRTSPSServerSet(s metricsRTSPServer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.rtspsServer = s
}

// OnRTMPServerSet is called by rtmpServer.
func (m *metrics) OnRTMPServerSet(s metricsRTMPServer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.rtmpServer = s
}
