package serve

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Invites lets a key holder pass an invite on. It is the one write surface
// authorized by a deploy key rather than the admin token: the people a node
// hosts can bring in the next person without going through the operator.
type Invites struct {
	ks *KeyStore
}

// NewInvites returns a handler that mints invites against ks.
func NewInvites(ks *KeyStore) *Invites {
	return &Invites{ks: ks}
}

// ServeHTTP handles POST /_relay/v1/invites with { "name": "…" }, carrying
// the inviter's own deploy token as the bearer token. Everything else is
// 404; an unknown or revoked inviter is 401 bad_token.
func (in *Invites) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	if r.Method != http.MethodPost || strings.TrimSuffix(r.URL.Path, "/") != "/_relay/v1/invites" {
		http.NotFound(w, r)
		return
	}
	_, token, ok := memberFromRequest(in.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or missing invitation")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	invited, key, err := in.ks.Invite(token, strings.TrimSpace(req.Name))
	switch {
	case errors.Is(err, ErrBadInviter):
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or missing invitation")
	case errors.Is(err, ErrBadUser):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	default:
		writeJSON(w, http.StatusCreated, map[string]any{"key": key, "token": invited})
	}
}
