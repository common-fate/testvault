package testvault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/common-fate/ddb"
	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testClient struct {
	Handler http.Handler
}

func (c *testClient) Do(req *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	c.Handler.ServeHTTP(rr, req)
	return rr.Result(), nil
}

// TestIntegration runs through the create/check/delete workflow
// using the live database.
func TestIntegration(t *testing.T) {
	_ = godotenv.Load("../.env")
	table := os.Getenv("TESTVAULT_TABLE_NAME")
	if table == "" {
		t.Skip("TESTVAULT_TABLE_NAME is not set")
	}

	ctx := context.Background()

	db, err := ddb.New(ctx, table)
	if err != nil {
		t.Fatal(err)
	}

	a, err := NewAPI(APIOpts{
		Log: zaptest.NewLogger(t).Sugar(),
		DB:  db,
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewClientWithResponses("http://test.internal", WithHTTPClient(&testClient{
		Handler: a.Server(),
	}))
	if err != nil {
		t.Fatal(err)
	}

	vaultID := ksuid.New().String()
	userID := ksuid.New().String()

	addRes, err := c.AddMemberToVault(ctx, vaultID, AddMemberToVaultJSONRequestBody{
		User: userID,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, addRes.StatusCode)

	checkRes, err := c.CheckVaultMembership(ctx, vaultID, userID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, checkRes.StatusCode)

	removeRes, err := c.RemoveMemberFromVault(ctx, vaultID, userID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, removeRes.StatusCode)

	checkRes, err = c.CheckVaultMembership(ctx, vaultID, userID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, checkRes.StatusCode)
}
