/*
HashiCorp Vault API

HTTP API that gives you full access to Vault. All API routes are prefixed with `/v1/`.

API version: 1.12.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package vault

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Client manages communication with the HashiCorp Vault API v1.12.0
// In most cases there should be only one, shared, Client.
type Client struct {
	configuration Configuration

	parsedBaseAddress url.URL

	client            *http.Client
	clientWithRetries *retryablehttp.Client

	// headers, callbacks, etc. that will be added to each request
	requestModifiers     requestModifiers
	requestModifiersLock sync.RWMutex

	// replication state cache used to ensure read-after-write semantics
	replicationStates replicationStateCache

	// API wrappers
	Auth     Auth
	Identity Identity
	Secrets  Secrets
	System   System
}

type (
	RequestCallback  func(*http.Request)
	ResponseCallback func(*http.Request, *http.Response)
)

// requestModifiers contains headers, callbacks, etc. that will be added to
// each request
type requestModifiers struct {
	headers requestHeaders

	requestCallbacks  []RequestCallback
	responseCallbacks []ResponseCallback

	// This error is set in client.WithX methods and checked in client.newRequest.
	// Since client.WithX methods are used for method chaining, they cannot
	// return errors.
	validationError error
}

// requestHeaders contains headers that will be added to each request
type requestHeaders struct {
	token                     string                    // 'X-Vault-Token'
	namespace                 string                    // 'X-Vault-Namespace'
	mfaCredentials            []string                  // 'X-Vault-MFA'
	responseWrappingTTL       time.Duration             // 'X-Vault-Wrap-TTL'
	replicationForwardingMode ReplicationForwardingMode // 'X-Vault-Forward' or 'X-Vault-Inconsistent'
	customHeaders             http.Header
}

// NewClient returns a new Vault client with a copy of the given configuration
func NewClient(configuration Configuration) (*Client, error) {
	// Ensure that the configuration fields are initialized
	configuration.SetDefaultsForUninitialized()

	c := Client{
		configuration: configuration,

		// configured or default HTTP client
		client: configuration.BaseClient,

		// retryablehttp wrapper around the HTTP client
		clientWithRetries: &retryablehttp.Client{
			HTTPClient:   configuration.BaseClient,
			Logger:       configuration.Retry.Logger,
			RetryWaitMin: configuration.Retry.RetryWaitMin,
			RetryWaitMax: configuration.Retry.RetryWaitMax,
			RetryMax:     configuration.Retry.RetryMax,
			CheckRetry:   configuration.Retry.CheckRetry,
			Backoff:      configuration.Retry.Backoff,
			ErrorHandler: configuration.Retry.ErrorHandler,
		},

		requestModifiers: requestModifiers{
			headers: requestHeaders{
				token:                     configuration.InitialToken,
				namespace:                 configuration.InitialNamespace,
				replicationForwardingMode: ReplicationForwardNone,
			},
			validationError: nil,
		},
		requestModifiersLock: sync.RWMutex{},
	}

	address, err := parseAddress(configuration.BaseAddress)
	if err != nil {
		return nil, err
	}
	c.parsedBaseAddress = *address

	// Internet draft https://datatracker.ietf.org/doc/html/draft-andrews-http-srv-02
	// specifies that the port must be empty
	if configuration.EnableSRVLookup && address.Port() != "" {
		return nil, fmt.Errorf("cannot enable service record (SRV) lookup since the base address port (%q) is not empty", address.Port())
	}

	transport, ok := c.client.Transport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("the configured base client's transport (%T) is not of type *http.Transport", c.client.Transport)
	}

	// Adjust the dial contex for unix domain socket addresses
	if strings.HasPrefix(configuration.BaseAddress, "unix://") {
		transport.DialContext = func(context.Context, string, string) (net.Conn, error) {
			socket := strings.TrimPrefix(configuration.BaseAddress, "unix://")
			return net.Dial("unix", socket)
		}
	}

	if err := configuration.TLS.applyTo(transport.TLSClientConfig); err != nil {
		return nil, err
	}

	c.Auth = Auth{
		client: &c,
	}
	c.Identity = Identity{
		client: &c,
	}
	c.Secrets = Secrets{
		client: &c,
	}
	c.System = System{
		client: &c,
	}

	return &c, nil
}

// parseAddress parses the given address string with special handling for unix
// domain sockets
func parseAddress(address string) (*url.URL, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(address, "unix://") {
		// The address in the client is expected to be pointing to the protocol
		// used in the application layer and not to the transport layer. Hence,
		// setting the fields accordingly.
		parsed.Scheme = "http"
		parsed.Host = strings.TrimPrefix(address, "unix://") // socket
		parsed.Path = ""
	}

	return parsed, nil
}

// Clone creates a new Vault client with the same configuration as the original
// client. Note that the cloned Vault client will point to the same base
// http.Client and retryablehttp.Client objects.
func (c *Client) Clone() *Client {
	clone := Client{
		configuration:     c.configuration,
		parsedBaseAddress: c.parsedBaseAddress,
		client:            c.client,
		clientWithRetries: c.clientWithRetries,
	}

	if c.configuration.EnforceReadYourWritesConsistency {
		clone.replicationStates = c.replicationStates.clone()
	}

	clone.requestModifiers = c.cloneRequestModifiers()

	clone.Auth = Auth{
		client: &clone,
	}
	clone.Identity = Identity{
		client: &clone,
	}
	clone.Secrets = Secrets{
		client: &clone,
	}
	clone.System = System{
		client: &clone,
	}

	return &clone
}

// cloneRequestModifiers returns a copy of the request modifiers behind a mutex;
// the replication states will point to the same cache
func (c *Client) cloneRequestModifiers() requestModifiers {
	/* */ c.requestModifiersLock.RLock()
	defer c.requestModifiersLock.RUnlock()

	var clone requestModifiers

	copy(clone.requestCallbacks, c.requestModifiers.requestCallbacks)
	copy(clone.responseCallbacks, c.requestModifiers.responseCallbacks)

	clone.headers = requestHeaders{
		token:                     c.requestModifiers.headers.token,
		namespace:                 c.requestModifiers.headers.namespace,
		responseWrappingTTL:       c.requestModifiers.headers.responseWrappingTTL,
		replicationForwardingMode: c.requestModifiers.headers.replicationForwardingMode,
		customHeaders:             c.requestModifiers.headers.customHeaders.Clone(),
	}

	copy(clone.headers.mfaCredentials, c.requestModifiers.headers.mfaCredentials)

	return clone
}

// Configuration returns a copy of the configuration object used to initialize
// this client
func (c *Client) Configuration() Configuration {
	return c.configuration
}
