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

	// 1a. Ask the supervisor to create a RelayServerActor
	createReplyChan := make(chan ActorMsg)
	createMsg := ActorMsg{
		To:        "test-supervisor",
		From:      "test",
		Type:      "create_child:RelayServerActor",
		Data:      "", // No gateway in this test
		ReplyChan: createReplyChan,
	}
	router.Send(createMsg)

	var relayServerName string
	select {
	case reply := <-createReplyChan:
		relayServerName = reply.Data.(string)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for worker actor to be created")
	}

	// 1b. The gateway now gets the worker's name.
	httpGateway := NewHTTPGatewayActor("http-gateway", "test-supervisor", relayServerName, router)
	httpGateway.Start()

	// Allow the server to start
	time.Sleep(50 * time.Millisecond)

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
