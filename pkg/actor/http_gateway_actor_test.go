package actor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPGatewayActor(t *testing.T) {
	router := NewRouter()
	defer router.StopAll()

	supervisor := NewSupervisorActor("test-supervisor", router)
	supervisor.Start()

	// The gateway actor now requires a supervisor name.
	gateway := NewHTTPGatewayActor("test-gateway", "test-supervisor", router)
	gateway.Start()

	// Create a test server from the gateway handler
	server := httptest.NewServer(gateway)
	defer server.Close()

	// The test now needs to use the /eval endpoint
	req, err := http.NewRequest("POST", server.URL+"/eval", strings.NewReader("1 + 1"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	gateway.ServeHTTP(rr, req)

	// Assert the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// We can check for a part of the error message
	expectedBody := "2"
	if !strings.Contains(rr.Body.String(), expectedBody) {
		t.Errorf("handler returned unexpected body: got %v wanted to contain %v",
			rr.Body.String(), expectedBody)
	}
}
