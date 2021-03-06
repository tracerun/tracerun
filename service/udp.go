package service

import (
	"net"

	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

// UDPServer to define a UDP server
type UDPServer struct {
	port   uint16
	router map[uint8]RouteFunc
}

// NewUDPServer to create a server instance
func NewUDPServer(port uint16, router map[uint8]RouteFunc) *UDPServer {
	return &UDPServer{
		port:   port,
		router: router,
	}
}

// Start the UDP server
func (s *UDPServer) Start() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: int(s.port)})
	if err != nil {
		lg.L.Error("error accept connection", zap.Error(err))
		panic(err)
	}
	lg.L.Info("started to listen udp connections", zap.Uint16("port", s.port))

	s.handleUDPConn(conn)
}

func (s *UDPServer) handleUDPConn(c *net.UDPConn) {
	defer func() {
		if r := recover(); r != nil {
			lg.L.Warn("recovered", zap.Any("error", r), zap.Stack("info"))
		}
	}()

	for {
		// read one request
		data, route, err := ReadOne(c)
		if err != nil {
			lg.L.Error("error to read data", zap.Error(err))
			break
		}
		lg.L.Debug("data", zap.Uint8("route", route), zap.Binary("data", data))

		// get routed function
		fn, ok := s.router[route]
		if !ok {
			lg.L.Warn("not found")
		} else {
			fn(data, c)
		}
	}

	if err := c.Close(); err != nil {
		lg.L.Error("error close", zap.Error(err))
	}
	lg.L.Debug("connection closed")
}
