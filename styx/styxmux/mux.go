package styxmux

import (
	"aqwari.net/net/styx"
	"regexp"
	"sync"
)

type TMatcher interface {
	MatchT(styx.Request) bool
}

type THandler interface {
	HandleT(styx.Request)
}

type THandlerFunc func(styx.Request)

func (f THandlerFunc) HandleT(t styx.Request) { f(t) }

type TMatchHandler interface {
	TMatcher
	THandler
}

type regexmh struct {
	THandler
	r *regexp.Regexp
}

func (h *regexmh) MatchT(t styx.Request) bool {
	return h.r.MatchString(t.Path())
}

func WithMatchT(h THandler, r *regexp.Regexp) TMatchHandler {
	return &regexmh{THandler: h, r: r}
}

type Mux struct {
	sync.RWMutex
	handlers []TMatchHandler
}

func (m *Mux) AppendHandlers(h ...TMatchHandler) {
	m.Lock()
	defer m.Unlock()

	if m.handlers == nil {
		m.handlers = make([]TMatchHandler, len(h))
	}
	m.handlers = append(m.handlers, h...)
}

func (m *Mux) HandleT(t styx.Request) {
	m.RLock()
	defer m.RUnlock()
	for _, h := range m.handlers {
		if h.Match(t.Path()) {
			h.HandleT(t)
			return
		}
	}
}

func (m *Mux) Serve9P(s *styx.Session) {
	for s.Next() {
		m.HandleT(s.Request())
	}
}

// NOTE: order matters.
func New(h ...TMatchHandler) *Mux {
	m := new(Mux)
	m.AppendHandlers(h...)
	return m
}
