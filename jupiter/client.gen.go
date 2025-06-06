// Package jupiter provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package jupiter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// Defines values for SwapMode.
const (
	ExactIn  SwapMode = "ExactIn"
	ExactOut SwapMode = "ExactOut"
)

// AccountMeta defines model for AccountMeta.
type AccountMeta struct {
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
	Pubkey     string `json:"pubkey"`
}

// Instruction defines model for Instruction.
type Instruction struct {
	Accounts  []AccountMeta `json:"accounts"`
	Data      string        `json:"data"`
	ProgramId string        `json:"programId"`
}

// PlatformFee defines model for PlatformFee.
type PlatformFee struct {
	Amount *string `json:"amount,omitempty"`
	FeeBps *int32  `json:"feeBps,omitempty"`
}

// QuoteResponse defines model for QuoteResponse.
type QuoteResponse struct {
	ContextSlot *float32 `json:"contextSlot,omitempty"`
	InAmount    string   `json:"inAmount"`
	InputMint   string   `json:"inputMint"`

	// OtherAmountThreshold - Calculated minimum output amount after accounting for `slippageBps` and `platformFeeBps`
	// - Not used by build transaction
	OtherAmountThreshold string `json:"otherAmountThreshold"`

	// OutAmount - Calculated output amount from routing algorithm
	// - Exlcuding network fees, slippage or platform fees
	OutAmount      string          `json:"outAmount"`
	OutputMint     string          `json:"outputMint"`
	PlatformFee    *PlatformFee    `json:"platformFee,omitempty"`
	PriceImpactPct string          `json:"priceImpactPct"`
	RoutePlan      []RoutePlanStep `json:"routePlan"`
	SlippageBps    int32           `json:"slippageBps"`
	SwapMode       SwapMode        `json:"swapMode"`

	// TimeTaken Time taken to determine quote
	TimeTaken *float32 `json:"timeTaken,omitempty"`
}

// RoutePlanStep defines model for RoutePlanStep.
type RoutePlanStep struct {
	Percent  int32    `json:"percent"`
	SwapInfo SwapInfo `json:"swapInfo"`
}

// SwapInfo defines model for SwapInfo.
type SwapInfo struct {
	AmmKey     string  `json:"ammKey"`
	FeeAmount  string  `json:"feeAmount"`
	FeeMint    string  `json:"feeMint"`
	InAmount   string  `json:"inAmount"`
	InputMint  string  `json:"inputMint"`
	Label      *string `json:"label,omitempty"`
	OutAmount  string  `json:"outAmount"`
	OutputMint string  `json:"outputMint"`
}

// SwapInstructionsResponse defines model for SwapInstructionsResponse.
type SwapInstructionsResponse struct {
	// AddressLookupTableAddresses The lookup table addresses that you can use if you are using versioned transaction.
	AddressLookupTableAddresses []string     `json:"addressLookupTableAddresses"`
	CleanupInstruction          *Instruction `json:"cleanupInstruction,omitempty"`

	// ComputeBudgetInstructions The necessary instructions to setup the compute budget.
	ComputeBudgetInstructions []Instruction `json:"computeBudgetInstructions"`

	// OtherInstructions If you set `{"prioritizationFeeLamports": {"jitoTipLamports": 5000}}`, you will see a custom tip instruction to Jito here.
	OtherInstructions []Instruction `json:"otherInstructions"`

	// SetupInstructions Setup missing ATA for the users.
	SetupInstructions []Instruction `json:"setupInstructions"`
	SwapInstruction   Instruction   `json:"swapInstruction"`
}

// SwapMode defines model for SwapMode.
type SwapMode string

// SwapRequest defines model for SwapRequest.
type SwapRequest struct {
	// AsLegacyTransaction Default: false
	// - Request a legacy transaction rather than the default versioned transaction
	// - Used together with `asLegacyTransaction` in /quote, otherwise the transaction might be too large
	AsLegacyTransaction *bool `json:"asLegacyTransaction,omitempty"`

	// BlockhashSlotsToExpiry Pass in the number of slots we want the transaction to be valid for
	// - Example: If you pass in 10 slots, the transaction will be valid for ~400ms * 10 = approximately 4 seconds before it expires
	BlockhashSlotsToExpiry *int `json:"blockhashSlotsToExpiry,omitempty"`

	// ComputeUnitPriceMicroLamports - To specify a compute unit price to calculate priority fee
	// - `computeUnitLimit (1400000) * computeUnitPriceMicroLamports`
	// - **We recommend using `prioritizationFeeLamports` and `dynamicComputeUnitLimit` instead of passing in a compute unit price**
	ComputeUnitPriceMicroLamports *int `json:"computeUnitPriceMicroLamports,omitempty"`

	// DestinationTokenAccount - Public key of a token account that will be used to receive the token out of the swap
	// - If not provided, the signer's ATA will be used
	// - If provided, we assume that the token account is already initialized
	DestinationTokenAccount *string `json:"destinationTokenAccount,omitempty"`

	// DynamicComputeUnitLimit Default: false
	// - When enabled, it will do a swap simulation to get the compute unit used and set it in ComputeBudget's compute unit limit
	// - This will increase latency slightly since there will be one extra RPC call to simulate this
	// - This can be useful to estimate compute unit correctly and reduce priority fees needed or have higher chance to be included in a block
	DynamicComputeUnitLimit *bool `json:"dynamicComputeUnitLimit,omitempty"`

	// DynamicSlippage Default: false
	// - When enabled, it estimate slippage and apply it in the swap transaction directly, overwriting the `slippageBps` parameter in the quote response.
	// - [See notes for more information](/docs/swap-api/send-swap-transaction#how-jupiter-estimates-slippage)
	DynamicSlippage *bool `json:"dynamicSlippage,omitempty"`

	// FeeAccount - An Associated Token Address (ATA) of specific mints depending on `SwapMode` to collect fees
	// - You no longer need the Referral Program
	// - See [Add Fees](/docs/swap-api/add-fees-to-swap) guide for more details
	FeeAccount *string `json:"feeAccount,omitempty"`

	// PrioritizationFeeLamports - To specify a level or amount of additional fees to prioritize the transaction
	// - It can be used for EITHER priority fee OR Jito tip
	// - If you want to include both, you will need to use `/swap-instructions` to add both at the same time
	PrioritizationFeeLamports *struct {
		// JitoTipLamports - Exact amount of tip to use in a tip instruction
		// - Estimate how much to set using Jito tip percentiles endpoint
		// - It has to be used together with a connection to a Jito RPC
		// - [See their docs](https://docs.jito.wtf/)
		JitoTipLamports              *int `json:"jitoTipLamports,omitempty"`
		PriorityLevelWithMaxLamports *struct {
			// MaxLamports Maximum lamports to cap the priority fee estimation, to prevent overpaying
			MaxLamports *int `json:"maxLamports,omitempty"`

			// PriorityLevel Either `medium`, `high` or `veryHigh`
			PriorityLevel *string `json:"priorityLevel,omitempty"`
		} `json:"priorityLevelWithMaxLamports,omitempty"`
	} `json:"prioritizationFeeLamports,omitempty"`
	QuoteResponse QuoteResponse `json:"quoteResponse"`

	// SkipUserAccountsRpcCalls Default: false
	// - When enabled, it will not do any additional RPC calls to check on user's accounts
	// - Enable it only when you already setup all the accounts needed for the trasaction, like wrapping or unwrapping sol, or destination account is already created
	SkipUserAccountsRpcCalls *bool `json:"skipUserAccountsRpcCalls,omitempty"`

	// TrackingAccount - Specify any public key that belongs to you to track the transactions
	// - Useful for integrators to get all the swap transactions from this public key
	// - Query the data using a block explorer like Solscan/SolanaFM or query like Dune/Flipside
	TrackingAccount *string `json:"trackingAccount,omitempty"`

	// UseSharedAccounts Default: true
	// - This enables the usage of shared program accounts, this is essential as complex routing will require multiple intermediate token accounts which the user might not have
	// - If true, you do not need to handle the creation of intermediate token accounts for the user.
	UseSharedAccounts *bool `json:"useSharedAccounts,omitempty"`

	// UserPublicKey The user public key
	UserPublicKey string `json:"userPublicKey"`

	// WrapAndUnwrapSol Default: true
	// - To automatically wrap/unwrap SOL in the transaction
	// - If false, it will use wSOL token account
	// - Parameter will be ignored if `destinationTokenAccount` is set because the `destinationTokenAccount` may belong to a different user that we have no authority to close
	WrapAndUnwrapSol *bool `json:"wrapAndUnwrapSol,omitempty"`
}

// SwapResponse defines model for SwapResponse.
type SwapResponse struct {
	LastValidBlockHeight      int    `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports *int   `json:"prioritizationFeeLamports,omitempty"`
	SwapTransaction           string `json:"swapTransaction"`
}

// GetQuoteParams defines parameters for GetQuote.
type GetQuoteParams struct {
	// InputMint Input token mint address
	InputMint string `form:"inputMint" json:"inputMint"`

	// OutputMint Output token mint address
	OutputMint string `form:"outputMint" json:"outputMint"`

	// Amount Amount of input token
	Amount float32 `form:"amount" json:"amount"`

	// SlippageBps Slippage in basis points
	SlippageBps *float32 `form:"slippageBps,omitempty" json:"slippageBps,omitempty"`

	// SwapMode Swap mode (ExactIn or ExactOut)
	SwapMode *SwapMode `form:"swapMode,omitempty" json:"swapMode,omitempty"`

	// Dexes List of DEXes to include
	Dexes *[]string `form:"dexes,omitempty" json:"dexes,omitempty"`

	// ExcludeDexes List of DEXes to exclude
	ExcludeDexes *[]string `form:"excludeDexes,omitempty" json:"excludeDexes,omitempty"`

	// RestrictIntermediateTokens Whether to restrict intermediate tokens
	RestrictIntermediateTokens *bool `form:"restrictIntermediateTokens,omitempty" json:"restrictIntermediateTokens,omitempty"`

	// OnlyDirectRoutes Whether to only use direct routes
	OnlyDirectRoutes *bool `form:"onlyDirectRoutes,omitempty" json:"onlyDirectRoutes,omitempty"`

	// AsLegacyTransaction Whether to use legacy transaction
	AsLegacyTransaction *bool `form:"asLegacyTransaction,omitempty" json:"asLegacyTransaction,omitempty"`

	// PlatformFeeBps Platform fee in basis points
	PlatformFeeBps *float32 `form:"platformFeeBps,omitempty" json:"platformFeeBps,omitempty"`

	// MaxAccounts Maximum number of accounts
	MaxAccounts *float32 `form:"maxAccounts,omitempty" json:"maxAccounts,omitempty"`
}

// PostSwapJSONRequestBody defines body for PostSwap for application/json ContentType.
type PostSwapJSONRequestBody = SwapRequest

// PostSwapInstructionsJSONRequestBody defines body for PostSwapInstructions for application/json ContentType.
type PostSwapInstructionsJSONRequestBody = SwapRequest

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetProgramIdToLabel request
	GetProgramIdToLabel(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetQuote request
	GetQuote(ctx context.Context, params *GetQuoteParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostSwapWithBody request with any body
	PostSwapWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostSwap(ctx context.Context, body PostSwapJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostSwapInstructionsWithBody request with any body
	PostSwapInstructionsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostSwapInstructions(ctx context.Context, body PostSwapInstructionsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetProgramIdToLabel(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetProgramIdToLabelRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetQuote(ctx context.Context, params *GetQuoteParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetQuoteRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSwapWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSwapRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSwap(ctx context.Context, body PostSwapJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSwapRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSwapInstructionsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSwapInstructionsRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSwapInstructions(ctx context.Context, body PostSwapInstructionsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSwapInstructionsRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetProgramIdToLabelRequest generates requests for GetProgramIdToLabel
func NewGetProgramIdToLabelRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/program-id-to-label")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetQuoteRequest generates requests for GetQuote
func NewGetQuoteRequest(server string, params *GetQuoteParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/quote")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "inputMint", runtime.ParamLocationQuery, params.InputMint); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "outputMint", runtime.ParamLocationQuery, params.OutputMint); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "amount", runtime.ParamLocationQuery, params.Amount); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

		if params.SlippageBps != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "slippageBps", runtime.ParamLocationQuery, *params.SlippageBps); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.SwapMode != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "swapMode", runtime.ParamLocationQuery, *params.SwapMode); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.Dexes != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "dexes", runtime.ParamLocationQuery, *params.Dexes); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.ExcludeDexes != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "excludeDexes", runtime.ParamLocationQuery, *params.ExcludeDexes); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.RestrictIntermediateTokens != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "restrictIntermediateTokens", runtime.ParamLocationQuery, *params.RestrictIntermediateTokens); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.OnlyDirectRoutes != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "onlyDirectRoutes", runtime.ParamLocationQuery, *params.OnlyDirectRoutes); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.AsLegacyTransaction != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "asLegacyTransaction", runtime.ParamLocationQuery, *params.AsLegacyTransaction); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.PlatformFeeBps != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "platformFeeBps", runtime.ParamLocationQuery, *params.PlatformFeeBps); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.MaxAccounts != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "maxAccounts", runtime.ParamLocationQuery, *params.MaxAccounts); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostSwapRequest calls the generic PostSwap builder with application/json body
func NewPostSwapRequest(server string, body PostSwapJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostSwapRequestWithBody(server, "application/json", bodyReader)
}

// NewPostSwapRequestWithBody generates requests for PostSwap with any type of body
func NewPostSwapRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/swap")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewPostSwapInstructionsRequest calls the generic PostSwapInstructions builder with application/json body
func NewPostSwapInstructionsRequest(server string, body PostSwapInstructionsJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostSwapInstructionsRequestWithBody(server, "application/json", bodyReader)
}

// NewPostSwapInstructionsRequestWithBody generates requests for PostSwapInstructions with any type of body
func NewPostSwapInstructionsRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/swap-instructions")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetProgramIdToLabelWithResponse request
	GetProgramIdToLabelWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetProgramIdToLabelResponse, error)

	// GetQuoteWithResponse request
	GetQuoteWithResponse(ctx context.Context, params *GetQuoteParams, reqEditors ...RequestEditorFn) (*GetQuoteResponse, error)

	// PostSwapWithBodyWithResponse request with any body
	PostSwapWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSwapResponse, error)

	PostSwapWithResponse(ctx context.Context, body PostSwapJSONRequestBody, reqEditors ...RequestEditorFn) (*PostSwapResponse, error)

	// PostSwapInstructionsWithBodyWithResponse request with any body
	PostSwapInstructionsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSwapInstructionsResponse, error)

	PostSwapInstructionsWithResponse(ctx context.Context, body PostSwapInstructionsJSONRequestBody, reqEditors ...RequestEditorFn) (*PostSwapInstructionsResponse, error)
}

type GetProgramIdToLabelResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *map[string]string
}

// Status returns HTTPResponse.Status
func (r GetProgramIdToLabelResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetProgramIdToLabelResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetQuoteResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *QuoteResponse
}

// Status returns HTTPResponse.Status
func (r GetQuoteResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetQuoteResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostSwapResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SwapResponse
}

// Status returns HTTPResponse.Status
func (r PostSwapResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostSwapResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostSwapInstructionsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SwapInstructionsResponse
}

// Status returns HTTPResponse.Status
func (r PostSwapInstructionsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostSwapInstructionsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetProgramIdToLabelWithResponse request returning *GetProgramIdToLabelResponse
func (c *ClientWithResponses) GetProgramIdToLabelWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetProgramIdToLabelResponse, error) {
	rsp, err := c.GetProgramIdToLabel(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetProgramIdToLabelResponse(rsp)
}

// GetQuoteWithResponse request returning *GetQuoteResponse
func (c *ClientWithResponses) GetQuoteWithResponse(ctx context.Context, params *GetQuoteParams, reqEditors ...RequestEditorFn) (*GetQuoteResponse, error) {
	rsp, err := c.GetQuote(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetQuoteResponse(rsp)
}

// PostSwapWithBodyWithResponse request with arbitrary body returning *PostSwapResponse
func (c *ClientWithResponses) PostSwapWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSwapResponse, error) {
	rsp, err := c.PostSwapWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSwapResponse(rsp)
}

func (c *ClientWithResponses) PostSwapWithResponse(ctx context.Context, body PostSwapJSONRequestBody, reqEditors ...RequestEditorFn) (*PostSwapResponse, error) {
	rsp, err := c.PostSwap(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSwapResponse(rsp)
}

// PostSwapInstructionsWithBodyWithResponse request with arbitrary body returning *PostSwapInstructionsResponse
func (c *ClientWithResponses) PostSwapInstructionsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSwapInstructionsResponse, error) {
	rsp, err := c.PostSwapInstructionsWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSwapInstructionsResponse(rsp)
}

func (c *ClientWithResponses) PostSwapInstructionsWithResponse(ctx context.Context, body PostSwapInstructionsJSONRequestBody, reqEditors ...RequestEditorFn) (*PostSwapInstructionsResponse, error) {
	rsp, err := c.PostSwapInstructions(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSwapInstructionsResponse(rsp)
}

// ParseGetProgramIdToLabelResponse parses an HTTP response from a GetProgramIdToLabelWithResponse call
func ParseGetProgramIdToLabelResponse(rsp *http.Response) (*GetProgramIdToLabelResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetProgramIdToLabelResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest map[string]string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetQuoteResponse parses an HTTP response from a GetQuoteWithResponse call
func ParseGetQuoteResponse(rsp *http.Response) (*GetQuoteResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetQuoteResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest QuoteResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostSwapResponse parses an HTTP response from a PostSwapWithResponse call
func ParsePostSwapResponse(rsp *http.Response) (*PostSwapResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostSwapResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SwapResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostSwapInstructionsResponse parses an HTTP response from a PostSwapInstructionsWithResponse call
func ParsePostSwapInstructionsResponse(rsp *http.Response) (*PostSwapInstructionsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostSwapInstructionsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SwapInstructionsResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
