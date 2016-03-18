package govh

import (
	"fmt"
	"net/http"
	"testing"
)

// Common helpers are in govh_test.go

func TestNewCkReqest(t *testing.T) {
	const expectedRequest = `{"accessRules":[{"method":"GET","path":"/me"},{"method":"GET","path":"/xdsl/*"}]}`

	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, 200, `{
		"validationUrl":"https://validation.url",
		"ConsumerKey":"`+MockConsumerKey+`",
		"state":"pendingValidation"
	}`, &InputRequestBody)
	client.ConsumerKey = ""
	defer ts.Close()

	// Test
	ckRequest := client.NewCkRequest()
	ckRequest.AddRule("GET", "/me")
	ckRequest.AddRule("GET", "/xdsl/*")

	got, err := ckRequest.Do()

	// Validate
	if err != nil {
		t.Fatalf("CkRequest.Do() should not return an error. Got: %q", err)
	}
	if client.ConsumerKey != MockConsumerKey {
		t.Fatalf("CkRequest.Do() should set client.ConsumerKey to %s. Got %s", MockConsumerKey, client.ConsumerKey)
	}
	if got.ConsumerKey != MockConsumerKey {
		t.Fatalf("CkRequest.Do() should set CkValidationState.ConsumerKey to %s. Got %s", MockConsumerKey, got.ConsumerKey)
	}
	if got.ValidationURL == "" {
		t.Fatalf("CkRequest.Do() should set CkValidationState.ValidationURL")
	}
	if InputRequestBody != expectedRequest {
		t.Fatalf("CkRequest.Do() should issue '%s' request. Got %s", expectedRequest, InputRequestBody)
	}
	ensureHeaderPresent(t, InputRequest, "Accept", "application/json")
	ensureHeaderPresent(t, InputRequest, "X-Ovh-Application", MockApplicationKey)
}

func TestInvalidCkReqest(t *testing.T) {
	// Init test
	var InputRequest *http.Request
	var InputRequestBody string
	ts, client := initMockServer(&InputRequest, http.StatusForbidden, `{"message":"Invalid application key"}`, &InputRequestBody)
	client.ConsumerKey = ""
	defer ts.Close()

	// Test
	ckRequest := client.NewCkRequest()
	ckRequest.AddRule("GET", "/me")
	ckRequest.AddRule("GET", "/xdsl/*")

	_, err := ckRequest.Do()
	apiError, ok := err.(*APIError)

	// Validate
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
	if !ok {
		t.Fatal("Expected error of type APIError")
	}
	if apiError.Code != http.StatusForbidden {
		t.Fatalf("Expected HTTP error 403. Got %d", apiError.Code)
	}
	if apiError.Message != "Invalid application key" {
		t.Fatalf("Expected API error message 'Invalid application key'. Got '%s'", apiError.Message)
	}
}

func TestCkReqestString(t *testing.T) {
	ckValidationState := &CkValidationState{
		ConsumerKey:   "ck",
		State:         "pending",
		ValidationURL: "fakeURL",
	}

	expected := fmt.Sprintf("CK: \"ck\"\nStatus: \"pending\"\nValidation URL: \"fakeURL\"\n")
	got := fmt.Sprintf("%s", ckValidationState)

	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
