package testvault

import (
	"fmt"
	"net/http"
	"time"

	"github.com/common-fate/apikit/apio"
	"github.com/common-fate/apikit/logger"
	"github.com/common-fate/apikit/openapi"
	"github.com/common-fate/ddb"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=.api-codegen.yaml openapi.yml

// API holds the HTTP routes for the testvault service.
type API struct {
	db      ddb.Storage
	log     *zap.SugaredLogger
	swagger *openapi3.T
}

var _ ServerInterface = &API{}

type APIOpts struct {
	DB  ddb.Storage
	Log *zap.SugaredLogger
}

// NewAPI creates a new API.
func NewAPI(opts APIOpts) (*API, error) {
	swagger, err := GetSwagger()
	if err != nil {
		return nil, err
	}
	// strip the servers from the swagger spec, as we don't know the hostname until runtime.
	swagger.Servers = nil
	a := API{
		db:      opts.DB,
		log:     opts.Log,
		swagger: swagger,
	}
	return &a, nil
}

// Add member to vault
// (POST /vaults/{vaultId}/members)
func (a *API) AddMemberToVault(w http.ResponseWriter, r *http.Request, vaultId string) {
	ctx := r.Context()

	var b AddMemberToVaultJSONRequestBody
	err := apio.DecodeJSONBody(w, r, &b)
	if err != nil {
		apio.Error(ctx, w, err)
		return
	}

	m := Membership{
		Vault:  vaultId,
		User:   b.User,
		Active: true,
	}
	err = a.db.Put(ctx, &m)
	if err != nil {
		apio.Error(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Check vault membership
// (GET /vaults/{vaultId}/members/{memberId})
func (a *API) CheckVaultMembership(w http.ResponseWriter, r *http.Request, vaultId string, memberId string) {
	ctx := r.Context()
	q := &GetMembership{
		Vault: vaultId,
		User:  memberId,
	}
	_, err := a.db.Query(ctx, q)
	if err == ddb.ErrNoItems || !q.Result.Active {
		apio.ErrorString(ctx, w, "user is not a member of this vault", http.StatusNotFound)
		return
	}

	if err != nil {
		// we don't know how to handle other errors, so fail with a HTTP 500
		apio.Error(ctx, w, err)
		return
	}

	// if we get here, the query succeeded, so the user is a member of the vault.
	res := MembershipResponse{
		Message: fmt.Sprintf("success! user %s is a member of vault %s", memberId, vaultId),
	}

	apio.JSON(ctx, w, res, http.StatusOK)
}

// Remove a member from a vault
// (POST /vaults/{vaultId}/members/{memberId}/remove)
func (a *API) RemoveMemberFromVault(w http.ResponseWriter, r *http.Request, vaultId string, memberId string) {
	ctx := r.Context()
	q := &GetMembership{
		Vault: vaultId,
		User:  memberId,
	}
	_, err := a.db.Query(ctx, q)
	if err == ddb.ErrNoItems {
		apio.ErrorString(ctx, w, "cannot remove user: user is not a member of this vault", http.StatusNotFound)
		return
	}
	if err != nil {
		// we don't know how to handle other errors, so fail with a HTTP 500
		apio.Error(ctx, w, err)
		return
	}
	m := q.Result
	m.Active = false
	err = a.db.Put(ctx, m)
	if err != nil {
		apio.Error(ctx, w, err)
		return
	}
	// if we get here, the delete worked, so return HTTP200 OK.
	w.WriteHeader(http.StatusOK)
}

func (a *API) Server() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(logger.Middleware(a.log.Desugar()))
	r.Use(cors.AllowAll().Handler)
	r.Use(openapi.Validator(a.swagger))

	// register the API handlers based on the OpenAPI spec
	// and return the resulting http.Handler.
	return HandlerFromMux(a, r)
}
