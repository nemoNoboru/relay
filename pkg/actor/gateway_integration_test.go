package actor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPGatewaySpawnsAndEvaluates(t *testing.T) {
	// 1. Setup our actor system
	router := NewRouter()
	defer router.StopAll()

	supervisor := NewSupervisorActor("supervisor", router)
	supervisor.Start()

	// The gateway needs to know its supervisor
	httpGateway := NewHTTPGatewayActor("http-gateway", "", "supervisor", router)
	httpGateway.Start()

	// Give the actors a moment to start up
	time.Sleep(10 * time.Millisecond)

	// 2. Create a test HTTP server from our gateway's handler
	// The gateway itself is an http.Handler.
	testServer := httptest.NewServer(httpGateway)
	defer testServer.Close()

	// 3. Make a POST request to the /eval endpoint
	evalEndpoint := testServer.URL + "/eval"
	reqBody := "10 + 5"
	resp, err := http.Post(evalEndpoint, "text/plain", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to send request to test server: %v", err)
	}
	defer resp.Body.Close()

	// 4. Assert the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// The RelayServerActor will return a Number, which has a float64 representation.
	// The runtime.Value's String() method will format it.
	expectedBody := "15"
	if string(bodyBytes) != expectedBody {
		t.Errorf("Expected response body '%s'; got '%s'", expectedBody, string(bodyBytes))
	}
}
