package serve

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

// memberCookie is the browser credential after an invitation has been
// exchanged for a session. Its value is the invitation token; the server
// still hashes and revokes it on every request, so there is no session table
// and revocation is immediate.
const memberCookie = "relay_member"

// sessionLifetime is how long a browser keeps the member cookie. The key
// itself can live longer; the member can always re-open the invite link.
const sessionLifetime = 30 * 24 * 60 * 60 // 30 days, in seconds

// loginMinInterval throttles password login attempts per username. It is an
// in-memory guard, not a security boundary; the node is a small group.
const loginMinInterval = 2 * time.Second

// NodeView is the small, member-facing description of the node.
type NodeView struct {
	Name    string `json:"name"`
	Members int    `json:"members"`
}

// Session handles the member session, login, and profile surface:
//
//	POST   /_relay/v1/session   exchange an invitation token for a cookie
//	POST   /_relay/v1/login     sign in with username + password
//	DELETE /_relay/v1/session   clear the cookie
//	GET    /_relay/v1/me        current member + node view
//	PATCH  /_relay/v1/me        update display name and bio
type Session struct {
	node domain.Name
	ks   *KeyStore
	rate *rateGuard
}

// NewSession returns the session handler for a node.
func NewSession(node domain.Name, ks *KeyStore) *Session {
	return &Session{node: node, ks: ks, rate: newRateGuard(loginMinInterval)}
}

// ServeHTTP routes the member session surface. Everything else is 404.
func (s *Session) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case p == "/_relay/v1/session" && r.Method == http.MethodPost:
		s.create(w, r)
	case p == "/_relay/v1/login" && r.Method == http.MethodPost:
		s.login(w, r)
	case p == "/_relay/v1/session" && r.Method == http.MethodDelete:
		if !sameOrigin(r) {
			writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
			return
		}
		s.clear(w)
	case p == "/_relay/v1/me" && r.Method == http.MethodGet:
		s.me(w, r)
	case p == "/_relay/v1/me" && r.Method == http.MethodPatch:
		s.update(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Session) create(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var req struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "token is required")
		return
	}
	member, ok := s.ks.MemberByToken(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or revoked invitation")
		return
	}
	if req.Username != "" || req.Password != "" {
		if req.Username == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "username and password are required together")
			return
		}
		if _, found, err := s.ks.SetCredentials(member.ID, req.Username, req.Password); err != nil {
			switch {
			case errors.Is(err, ErrBadUser), errors.Is(err, ErrBadPassword):
				writeError(w, http.StatusBadRequest, "bad_request", err.Error())
			case errors.Is(err, ErrUsernameTaken):
				writeError(w, http.StatusConflict, "username_taken", err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "internal", err.Error())
			}
			return
		} else if !found {
			// The member was resolved from the token above, so a missing key
			// here means it was revoked mid-request.
			writeError(w, http.StatusUnauthorized, "bad_token", "invalid or revoked invitation")
			return
		}
	}
	s.setCookie(w, r, token)
	writeJSON(w, http.StatusOK, s.memberReply(member))
}

func (s *Session) clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     memberCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// login authenticates a member by username and password and sets the same
// member cookie the invitation exchange sets. Unknown usernames and wrong
// passwords return the identical bad_credentials error.
func (s *Session) login(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		writeError(w, http.StatusUnauthorized, "bad_credentials", "invalid username or password")
		return
	}
	if !s.rate.allow(strings.ToLower(username)) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "wait a moment before trying again")
		return
	}
	token, ok := s.ks.AuthenticatePassword(username, req.Password)
	if !ok {
		writeError(w, http.StatusUnauthorized, "bad_credentials", "invalid username or password")
		return
	}
	member, ok := s.ks.MemberByToken(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "bad_credentials", "invalid username or password")
		return
	}
	s.setCookie(w, r, token)
	writeJSON(w, http.StatusOK, s.memberReply(member))
}

func (s *Session) me(w http.ResponseWriter, r *http.Request) {
	member, _, ok := memberFromRequest(s.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	writeJSON(w, http.StatusOK, s.memberReply(member))
}

func (s *Session) update(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	_, token, ok := memberFromRequest(s.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var req struct {
		DisplayName string `json:"displayName"`
		Bio         string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	member, found, err := s.ks.UpdateProfile(token, req.DisplayName, req.Bio)
	switch {
	case errors.Is(err, ErrBadProfile):
		writeError(w, http.StatusBadRequest, "bad_profile", err.Error())
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	case !found:
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
	default:
		writeJSON(w, http.StatusOK, s.memberReply(member))
	}
}

// memberFromRequest resolves the member cookie first, then a bearer token.
// The bearer path keeps the CLI, uploads, and assistants working with the same
// credential. Both paths are the same lookup, so revocation is immediate.
func memberFromRequest(ks *KeyStore, r *http.Request) (Member, string, bool) {
	if c, err := r.Cookie(memberCookie); err == nil && c.Value != "" {
		if m, ok := ks.MemberByToken(c.Value); ok {
			return m, c.Value, true
		}
	}
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok && token != "" {
		if m, ok := ks.MemberByToken(token); ok {
			return m, token, true
		}
	}
	return Member{}, "", false
}

func (s *Session) setCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     memberCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionLifetime,
		HttpOnly: true,
		Secure:   requestIsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Session) nodeView() NodeView {
	return NodeView{Name: string(s.node), Members: s.ks.LiveCount()}
}

// memberReply is the session surface's response: the member view, the node
// summary, and the member's own username (empty until credentials are set).
func (s *Session) memberReply(m Member) map[string]any {
	return map[string]any{"member": m, "node": s.nodeView(), "username": s.ks.UsernameByID(m.ID)}
}

func requestIsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
