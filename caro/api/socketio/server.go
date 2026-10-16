package socketio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"htqrcode/caro/api"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	"htqrcode/common"
	"htqrcode/common/log"
)

const (
	Path       = "/api/v1/caro/socket.io/"
	maxPayload = 64 * 1024
)

type (
	member struct {
		socket *socket.Socket
		code   string
	}
	Server struct {
		service         *app.Service
		io              *socket.Server
		handler         http.Handler
		mu              sync.Mutex
		members         map[string]member
		ctx             context.Context
		ready, stopping bool
		requests        sync.WaitGroup
	}
)

func NewServer(service *app.Service) *Server {
	opts := socket.DefaultServerOptions()
	opts.SetPath(Path)
	opts.SetMaxHttpBufferSize(maxPayload)
	server := socket.NewServer(nil, opts)
	s := &Server{service: service, io: server, members: map[string]member{}}
	s.handler = server.ServeHandler(nil)
	_ = server.On("connection", func(args ...any) { s.connected(args[0].(*socket.Socket)) })
	return s
}

func (s *Server) Register(e common.EchoRouter) {
	h := func(c *echo.Context) error {
		req := c.Request()
		if origin := req.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			scheme := c.Scheme()
			normalize := func(host string) string {
				port := ":80"
				if scheme == "https" {
					port = ":443"
				}
				return strings.TrimSuffix(strings.ToLower(host), port)
			}
			if err != nil || u.Scheme != scheme || normalize(u.Host) != normalize(req.Host) || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
				return common.NewForbiddenError("origin_forbidden", "cross-origin requests are not allowed")
			}
		}
		s.mu.Lock()
		if !s.ready || s.stopping {
			s.mu.Unlock()
			return echo.NewHTTPError(http.StatusServiceUnavailable, "signaling unavailable")
		}
		s.requests.Add(1)
		s.mu.Unlock()
		defer s.requests.Done()
		// Polling and upgraded transports own their deadlines, including on embedded servers.
		_ = http.NewResponseController(c.Response()).SetWriteDeadline(time.Time{})
		req.Body = http.MaxBytesReader(c.Response(), req.Body, maxPayload)
		ip, err := domain.NormalizeIP(c.RealIP())
		if err != nil {
			return api.PublicError(err)
		}
		// Preserve Echo's trusted IP resolution in the Engine.IO handshake request.
		engineRequest := req.Clone(req.Context())
		engineRequest.RemoteAddr = net.JoinHostPort(ip, "0")
		s.handler.ServeHTTP(c.Response(), engineRequest)
		return nil
	}
	e.GET(Path, h)
	e.POST(Path, h)
	e.GET(Path+"*", h)
	e.POST(Path+"*", h)
}

func (s *Server) connected(client *socket.Socket) {
	s.mu.Lock()
	if !s.ready || s.stopping {
		s.mu.Unlock()
		client.Disconnect(true)
		return
	}
	s.mu.Unlock()
	for _, event := range []string{"join_room", "leave_room", "signal_offer", "signal_answer", "signal_ice"} {
		_ = client.On(event, func(args ...any) { s.event(client, event, args) })
	}
	_ = client.On("disconnect", func(...any) { s.disconnect(client) })
}

type payload struct {
	Code        string          `json:"code"`
	DisplayName string          `json:"displayName,omitempty"`
	HostToken   string          `json:"hostToken,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

func decode(args []any) (payload, error) {
	if len(args) != 1 {
		return payload{}, domain.ErrInvalid
	}
	b, err := json.Marshal(args[0])
	if err != nil || len(b) > maxPayload {
		return payload{}, domain.ErrInvalid
	}
	var p payload
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(&p); err != nil {
		return p, domain.ErrInvalid
	}
	p.Code, err = domain.NormalizeCode(p.Code)
	return p, err
}

func (s *Server) event(client *socket.Socket, event string, args []any) {
	var ack socket.Ack
	if len(args) > 0 {
		if fn, ok := args[len(args)-1].(socket.Ack); ok {
			ack = fn
			args = args[:len(args)-1]
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopping || !s.ready {
		return
	}
	p, err := decode(args)
	if err == nil {
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		defer cancel()
		switch event {
		case "join_room":
			err = s.join(ctx, client, p)
		case "leave_room":
			if p.HostToken != "" || p.DisplayName != "" || len(p.Data) > 0 {
				err = domain.ErrInvalid
				break
			}
			m, ok := s.members[string(client.Id())]
			if ok && m.code != p.Code {
				err = domain.ErrUnauthorized
				break
			}
			err = s.leave(ctx, client)
		default:
			err = s.relay(ctx, client, event, p)
		}
	}
	if err != nil {
		public, _ := common.PublicHTTPError(api.PublicError(err))
		_ = client.Emit("room_error", public)
		if ack != nil {
			ack([]any{map[string]any{"ok": false, "error": public}}, nil)
		}
		return
	}
	if ack != nil {
		ack([]any{map[string]any{"ok": true}}, nil)
	}
}

func (s *Server) join(ctx context.Context, client *socket.Socket, p payload) error {
	if len(p.Data) > 0 || (p.HostToken != "" && (len(p.HostToken) != 64 || p.DisplayName != "")) {
		return domain.ErrInvalid
	}
	id := string(client.Id())
	if old, ok := s.members[id]; ok && old.code != p.Code {
		return app.ErrMembershipConflict
	}
	ip, _, err := net.SplitHostPort(client.Request().Request().RemoteAddr)
	if err != nil {
		return domain.ErrInvalid
	}
	result, err := s.service.Join(ctx, p.Code, p.DisplayName, id, p.HostToken, ip)
	if err != nil {
		return err
	}
	if !result.Changed {
		return nil
	}
	room := socket.Room("room:" + p.Code)
	s.members[id] = member{client, p.Code}
	client.Join(room)
	if client.Disconnected() || !client.Rooms().Has(room) {
		compensation, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if leaveErr := s.leave(compensation, client); leaveErr != nil {
			return leaveErr
		}
		return domain.ErrUnavailable
	}
	if err := client.Emit("room_joined", map[string]string{"code": p.Code, "role": result.Player.Role()}); err != nil {
		_ = s.leave(ctx, client)
		return err
	}
	if result.Room.Status() == domain.Playing {
		for _, player := range result.Room.Players() {
			if player.Role() == domain.Host {
				if host, ok := s.members[player.Socket()]; ok {
					_ = host.socket.Emit("peer_joined", map[string]string{"code": p.Code})
				}
			}
		}
	}
	return nil
}

func (s *Server) relay(ctx context.Context, client *socket.Socket, event string, p payload) error {
	if p.HostToken != "" || p.DisplayName != "" {
		return domain.ErrInvalid
	}
	m, ok := s.members[string(client.Id())]
	if !ok || m.code != p.Code {
		return domain.ErrUnauthorized
	}
	if !validSignal(event, p.Data) {
		return domain.ErrInvalid
	}
	peer, err := s.service.Peer(ctx, p.Code, string(client.Id()))
	if err != nil {
		return err
	}
	target, ok := s.members[peer]
	if !ok || target.code != p.Code {
		return domain.ErrUnavailable
	}
	var data map[string]any
	if err := json.Unmarshal(p.Data, &data); err != nil {
		return domain.ErrInvalid
	}
	return target.socket.Emit(event, map[string]any{"code": p.Code, "data": data})
}

func validSignal(event string, data json.RawMessage) bool {
	if len(data) == 0 || len(data) > maxPayload {
		return false
	}
	if event == "signal_ice" {
		var ice struct {
			Candidate     *string `json:"candidate"`
			SDPMid        *string `json:"sdpMid"`
			SDPMLineIndex *int    `json:"sdpMLineIndex"`
		}
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		return d.Decode(&ice) == nil && ice.Candidate != nil && len(*ice.Candidate) <= 8192 && (ice.SDPMLineIndex == nil || *ice.SDPMLineIndex >= 0)
	}
	var sdp struct {
		Type string `json:"type"`
		SDP  string `json:"sdp"`
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(&sdp) == nil && sdp.Type == strings.TrimPrefix(event, "signal_") && sdp.SDP != ""
}

func (s *Server) leave(ctx context.Context, client *socket.Socket) error {
	id := string(client.Id())
	m, ok := s.members[id]
	if !ok {
		return nil
	}
	result, err := s.service.Leave(ctx, m.code, id)
	if err != nil && !errors.Is(err, app.ErrNotFound) && !errors.Is(err, domain.ErrUnavailable) {
		return err
	}
	delete(s.members, id)
	client.Leave(socket.Room("room:" + m.code))
	if err == nil && result.Changed {
		for _, p := range result.Room.Players() {
			if peer, ok := s.members[p.Socket()]; ok {
				_ = peer.socket.Emit("peer_disconnected", map[string]string{"code": m.code})
			}
		}
	}
	return nil
}

func (s *Server) disconnect(client *socket.Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopping {
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	if err := s.leave(ctx, client); err != nil {
		log.FromContext(ctx).Error("Caro disconnect failed", "error", err)
	}
}

// Reconcile checks every registry entry, including retries after deletion succeeded.
func (s *Server) Reconcile(ctx context.Context) error {
	s.mu.Lock()
	var closeSockets []*socket.Socket
	var result error
	for id, m := range s.members {
		if m.socket.Disconnected() || !m.socket.Rooms().Has(socket.Room("room:"+m.code)) {
			if err := s.leave(ctx, m.socket); err != nil {
				result = errors.Join(result, err)
			}
			continue
		}
		r, err := s.service.Get(ctx, m.code)
		_, present := r.Member(id)
		if errors.Is(err, app.ErrNotFound) || (err == nil && (!present || r.Status() == domain.Finished)) {
			delete(s.members, id)
			closeSockets = append(closeSockets, m.socket)
		} else if err != nil {
			result = errors.Join(result, err)
		}
	}
	s.mu.Unlock()
	for _, client := range closeSockets {
		client.Disconnect(true)
	}
	return result
}

func (s *Server) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	if err := s.service.ResetConnections(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	s.mu.Lock()
	s.ctx = ctx
	s.ready = true
	s.mu.Unlock()
	<-ctx.Done()
	s.mu.Lock()
	s.stopping = true
	s.ready = false
	s.mu.Unlock()
	done := make(chan struct{})
	s.io.Close(func(error) { close(done) })
	<-done
	s.requests.Wait()
	s.mu.Lock()
	clear(s.members)
	s.mu.Unlock()
	return nil
}
