package actor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPGatewayActor(t *testing.T) {
	router := NewRouter()
	defer router.StopAll()

	// The gateway actor now requires a supervisor name, even for this test.
	gateway := NewHTTPGatewayActor("test-gateway", "", "test-supervisor", router)
	// We don't call gateway.Start() because httptest.NewServer will do the listening.
	// gateway.Start()

	// Create a test server from the gateway handler
	server := httptest.NewServer(gateway)
	defer server.Close()

	// Give the server a moment to start
	time.Sleep(10 * time.Millisecond)

	// Create a mock receiver actor
	receiver := NewActor("test-receiver", router, func(msg ActorMsg) {
		// In a real test, we would assert the content of msg
		t.Logf("Test receiver got message: %v", msg)
	})
	receiver.Start()
	defer receiver.Stop()

	// The test now needs to use the /eval endpoint
	// In the test, we don't have a real supervisor, so the gateway will time out
	// waiting for a reply. We expect a 500 Internal Server Error.
	req, err := http.NewRequest("POST", server.URL+"/eval", strings.NewReader("1+1"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	gateway.ServeHTTP(rr, req)

	// Assert the response
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	// We can check for a part of the error message
	expectedBodyPart := "Timeout waiting for eval result"
	if !strings.Contains(rr.Body.String(), expectedBodyPart) {
		t.Errorf("handler returned unexpected body: got %v wanted to contain %v",
			rr.Body.String(), expectedBodyPart)
	}
}
