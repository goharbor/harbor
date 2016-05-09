// Package adexchangebuyer provides access to the Ad Exchange Buyer API.
//
// See https://developers.google.com/ad-exchange/buyer-rest
//
// Usage example:
//
//   import "google.golang.org/api/adexchangebuyer/v1.4"
//   ...
//   adexchangebuyerService, err := adexchangebuyer.New(oauthHttpClient)
package adexchangebuyer // import "google.golang.org/api/adexchangebuyer/v1.4"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/internal"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = internal.MarshalJSON
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "adexchangebuyer:v1.4"
const apiName = "adexchangebuyer"
const apiVersion = "v1.4"
const basePath = "https://www.googleapis.com/adexchangebuyer/v1.4/"

// OAuth2 scopes used by this API.
const (
	// Manage your Ad Exchange buyer account configuration
	AdexchangeBuyerScope = "https://www.googleapis.com/auth/adexchange.buyer"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Accounts = NewAccountsService(s)
	s.BillingInfo = NewBillingInfoService(s)
	s.Budget = NewBudgetService(s)
	s.Clientaccess = NewClientaccessService(s)
	s.Creatives = NewCreativesService(s)
	s.Deals = NewDealsService(s)
	s.Marketplacedeals = NewMarketplacedealsService(s)
	s.Marketplacenotes = NewMarketplacenotesService(s)
	s.Marketplaceoffers = NewMarketplaceoffersService(s)
	s.Marketplaceorders = NewMarketplaceordersService(s)
	s.Negotiationrounds = NewNegotiationroundsService(s)
	s.Negotiations = NewNegotiationsService(s)
	s.Offers = NewOffersService(s)
	s.PerformanceReport = NewPerformanceReportService(s)
	s.PretargetingConfig = NewPretargetingConfigService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Accounts *AccountsService

	BillingInfo *BillingInfoService

	Budget *BudgetService

	Clientaccess *ClientaccessService

	Creatives *CreativesService

	Deals *DealsService

	Marketplacedeals *MarketplacedealsService

	Marketplacenotes *MarketplacenotesService

	Marketplaceoffers *MarketplaceoffersService

	Marketplaceorders *MarketplaceordersService

	Negotiationrounds *NegotiationroundsService

	Negotiations *NegotiationsService

	Offers *OffersService

	PerformanceReport *PerformanceReportService

	PretargetingConfig *PretargetingConfigService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewAccountsService(s *Service) *AccountsService {
	rs := &AccountsService{s: s}
	return rs
}

type AccountsService struct {
	s *Service
}

func NewBillingInfoService(s *Service) *BillingInfoService {
	rs := &BillingInfoService{s: s}
	return rs
}

type BillingInfoService struct {
	s *Service
}

func NewBudgetService(s *Service) *BudgetService {
	rs := &BudgetService{s: s}
	return rs
}

type BudgetService struct {
	s *Service
}

func NewClientaccessService(s *Service) *ClientaccessService {
	rs := &ClientaccessService{s: s}
	return rs
}

type ClientaccessService struct {
	s *Service
}

func NewCreativesService(s *Service) *CreativesService {
	rs := &CreativesService{s: s}
	return rs
}

type CreativesService struct {
	s *Service
}

func NewDealsService(s *Service) *DealsService {
	rs := &DealsService{s: s}
	return rs
}

type DealsService struct {
	s *Service
}

func NewMarketplacedealsService(s *Service) *MarketplacedealsService {
	rs := &MarketplacedealsService{s: s}
	return rs
}

type MarketplacedealsService struct {
	s *Service
}

func NewMarketplacenotesService(s *Service) *MarketplacenotesService {
	rs := &MarketplacenotesService{s: s}
	return rs
}

type MarketplacenotesService struct {
	s *Service
}

func NewMarketplaceoffersService(s *Service) *MarketplaceoffersService {
	rs := &MarketplaceoffersService{s: s}
	return rs
}

type MarketplaceoffersService struct {
	s *Service
}

func NewMarketplaceordersService(s *Service) *MarketplaceordersService {
	rs := &MarketplaceordersService{s: s}
	return rs
}

type MarketplaceordersService struct {
	s *Service
}

func NewNegotiationroundsService(s *Service) *NegotiationroundsService {
	rs := &NegotiationroundsService{s: s}
	return rs
}

type NegotiationroundsService struct {
	s *Service
}

func NewNegotiationsService(s *Service) *NegotiationsService {
	rs := &NegotiationsService{s: s}
	return rs
}

type NegotiationsService struct {
	s *Service
}

func NewOffersService(s *Service) *OffersService {
	rs := &OffersService{s: s}
	return rs
}

type OffersService struct {
	s *Service
}

func NewPerformanceReportService(s *Service) *PerformanceReportService {
	rs := &PerformanceReportService{s: s}
	return rs
}

type PerformanceReportService struct {
	s *Service
}

func NewPretargetingConfigService(s *Service) *PretargetingConfigService {
	rs := &PretargetingConfigService{s: s}
	return rs
}

type PretargetingConfigService struct {
	s *Service
}

// Account: Configuration data for an Ad Exchange buyer account.
type Account struct {
	// BidderLocation: Your bidder locations that have distinct URLs.
	BidderLocation []*AccountBidderLocation `json:"bidderLocation,omitempty"`

	// CookieMatchingNid: The nid parameter value used in cookie match
	// requests. Please contact your technical account manager if you need
	// to change this.
	CookieMatchingNid string `json:"cookieMatchingNid,omitempty"`

	// CookieMatchingUrl: The base URL used in cookie match requests.
	CookieMatchingUrl string `json:"cookieMatchingUrl,omitempty"`

	// Id: Account id.
	Id int64 `json:"id,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// MaximumActiveCreatives: The maximum number of active creatives that
	// an account can have, where a creative is active if it was inserted or
	// bid with in the last 30 days. Please contact your technical account
	// manager if you need to change this.
	MaximumActiveCreatives int64 `json:"maximumActiveCreatives,omitempty"`

	// MaximumTotalQps: The sum of all bidderLocation.maximumQps values
	// cannot exceed this. Please contact your technical account manager if
	// you need to change this.
	MaximumTotalQps int64 `json:"maximumTotalQps,omitempty"`

	// NumberActiveCreatives: The number of creatives that this account
	// inserted or bid with in the last 30 days.
	NumberActiveCreatives int64 `json:"numberActiveCreatives,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "BidderLocation") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Account) MarshalJSON() ([]byte, error) {
	type noMethod Account
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AccountBidderLocation struct {
	// MaximumQps: The maximum queries per second the Ad Exchange will send.
	MaximumQps int64 `json:"maximumQps,omitempty"`

	// Region: The geographical region the Ad Exchange should send requests
	// from. Only used by some quota systems, but always setting the value
	// is recommended. Allowed values:
	// - ASIA
	// - EUROPE
	// - US_EAST
	// - US_WEST
	Region string `json:"region,omitempty"`

	// Url: The URL to which the Ad Exchange will send bid requests.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaximumQps") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AccountBidderLocation) MarshalJSON() ([]byte, error) {
	type noMethod AccountBidderLocation
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// AccountsList: An account feed lists Ad Exchange buyer accounts that
// the user has access to. Each entry in the feed corresponds to a
// single buyer account.
type AccountsList struct {
	// Items: A list of accounts.
	Items []*Account `json:"items,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Items") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AccountsList) MarshalJSON() ([]byte, error) {
	type noMethod AccountsList
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AdSize struct {
	Height int64 `json:"height,omitempty"`

	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AdSize) MarshalJSON() ([]byte, error) {
	type noMethod AdSize
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AdSlotDto struct {
	ChannelCode string `json:"channelCode,omitempty"`

	ChannelId int64 `json:"channelId,omitempty"`

	Description string `json:"description,omitempty"`

	Name string `json:"name,omitempty"`

	Size string `json:"size,omitempty"`

	WebPropertyId int64 `json:"webPropertyId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ChannelCode") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AdSlotDto) MarshalJSON() ([]byte, error) {
	type noMethod AdSlotDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AddOrderDealsRequest struct {
	// Deals: The list of deals to add
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// OrderRevisionNumber: The last known order revision number.
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	// UpdateAction: Indicates an optional action to take on the order
	UpdateAction string `json:"updateAction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AddOrderDealsRequest) MarshalJSON() ([]byte, error) {
	type noMethod AddOrderDealsRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AddOrderDealsResponse struct {
	// Deals: List of deals added (in the same order as passed in the
	// request)
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// OrderRevisionNumber: The updated revision number for the order.
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AddOrderDealsResponse) MarshalJSON() ([]byte, error) {
	type noMethod AddOrderDealsResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AddOrderNotesRequest struct {
	// Notes: The list of notes to add.
	Notes []*MarketplaceNote `json:"notes,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Notes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AddOrderNotesRequest) MarshalJSON() ([]byte, error) {
	type noMethod AddOrderNotesRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AddOrderNotesResponse struct {
	Notes []*MarketplaceNote `json:"notes,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Notes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AddOrderNotesResponse) MarshalJSON() ([]byte, error) {
	type noMethod AddOrderNotesResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AdvertiserDto struct {
	Brands []*BrandDto `json:"brands,omitempty"`

	Id int64 `json:"id,omitempty,string"`

	Name string `json:"name,omitempty"`

	Status string `json:"status,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Brands") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AdvertiserDto) MarshalJSON() ([]byte, error) {
	type noMethod AdvertiserDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type AudienceSegment struct {
	Description string `json:"description,omitempty"`

	Id int64 `json:"id,omitempty,string"`

	Name string `json:"name,omitempty"`

	NumCookies int64 `json:"numCookies,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AudienceSegment) MarshalJSON() ([]byte, error) {
	type noMethod AudienceSegment
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// BillingInfo: The configuration data for an Ad Exchange billing info.
type BillingInfo struct {
	// AccountId: Account id.
	AccountId int64 `json:"accountId,omitempty"`

	// AccountName: Account name.
	AccountName string `json:"accountName,omitempty"`

	// BillingId: A list of adgroup IDs associated with this particular
	// account. These IDs may show up as part of a realtime bidding
	// BidRequest, which indicates a bid request for this account.
	BillingId []string `json:"billingId,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BillingInfo) MarshalJSON() ([]byte, error) {
	type noMethod BillingInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// BillingInfoList: A billing info feed lists Billing Info the Ad
// Exchange buyer account has access to. Each entry in the feed
// corresponds to a single billing info.
type BillingInfoList struct {
	// Items: A list of billing info relevant for your account.
	Items []*BillingInfo `json:"items,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Items") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BillingInfoList) MarshalJSON() ([]byte, error) {
	type noMethod BillingInfoList
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type BrandDto struct {
	AdvertiserId int64 `json:"advertiserId,omitempty,string"`

	Id int64 `json:"id,omitempty,string"`

	Name string `json:"name,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AdvertiserId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BrandDto) MarshalJSON() ([]byte, error) {
	type noMethod BrandDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// Budget: The configuration data for Ad Exchange RTB - Budget API.
type Budget struct {
	// AccountId: The id of the account. This is required for get and update
	// requests.
	AccountId int64 `json:"accountId,omitempty,string"`

	// BillingId: The billing id to determine which adgroup to provide
	// budget information for. This is required for get and update requests.
	BillingId int64 `json:"billingId,omitempty,string"`

	// BudgetAmount: The budget amount to apply for the billingId provided.
	// This is required for update requests.
	BudgetAmount int64 `json:"budgetAmount,omitempty,string"`

	// CurrencyCode: The currency code for the buyer. This cannot be altered
	// here.
	CurrencyCode string `json:"currencyCode,omitempty"`

	// Id: The unique id that describes this item.
	Id string `json:"id,omitempty"`

	// Kind: The kind of the resource, i.e. "adexchangebuyer#budget".
	Kind string `json:"kind,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Budget) MarshalJSON() ([]byte, error) {
	type noMethod Budget
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Buyer struct {
	// AccountId: Adx account id of the buyer.
	AccountId string `json:"accountId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Buyer) MarshalJSON() ([]byte, error) {
	type noMethod Buyer
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type BuyerDto struct {
	AccountId int64 `json:"accountId,omitempty"`

	CustomerId int64 `json:"customerId,omitempty"`

	DisplayName string `json:"displayName,omitempty"`

	EnabledForInterestTargetingDeals bool `json:"enabledForInterestTargetingDeals,omitempty"`

	EnabledForPreferredDeals bool `json:"enabledForPreferredDeals,omitempty"`

	Id int64 `json:"id,omitempty"`

	SponsorAccountId int64 `json:"sponsorAccountId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BuyerDto) MarshalJSON() ([]byte, error) {
	type noMethod BuyerDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ClientAccessCapabilities struct {
	Capabilities []int64 `json:"capabilities,omitempty"`

	ClientAccountId int64 `json:"clientAccountId,omitempty,string"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Capabilities") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClientAccessCapabilities) MarshalJSON() ([]byte, error) {
	type noMethod ClientAccessCapabilities
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ContactInformation struct {
	// Email: Email address of the contact.
	Email string `json:"email,omitempty"`

	// Name: The name of the contact.
	Name string `json:"name,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Email") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ContactInformation) MarshalJSON() ([]byte, error) {
	type noMethod ContactInformation
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreateOrdersRequest struct {
	// Orders: The list of orders to create.
	Orders []*MarketplaceOrder `json:"orders,omitempty"`

	WebPropertyCode string `json:"webPropertyCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Orders") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreateOrdersRequest) MarshalJSON() ([]byte, error) {
	type noMethod CreateOrdersRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreateOrdersResponse struct {
	// Orders: The list of orders successfully created.
	Orders []*MarketplaceOrder `json:"orders,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Orders") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreateOrdersResponse) MarshalJSON() ([]byte, error) {
	type noMethod CreateOrdersResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// Creative: A creative and its classification data.
type Creative struct {
	// HTMLSnippet: The HTML snippet that displays the ad when inserted in
	// the web page. If set, videoURL should not be set.
	HTMLSnippet string `json:"HTMLSnippet,omitempty"`

	// AccountId: Account id.
	AccountId int64 `json:"accountId,omitempty"`

	// AdvertiserId: Detected advertiser id, if any. Read-only. This field
	// should not be set in requests.
	AdvertiserId googleapi.Int64s `json:"advertiserId,omitempty"`

	// AdvertiserName: The name of the company being advertised in the
	// creative.
	AdvertiserName string `json:"advertiserName,omitempty"`

	// AgencyId: The agency id for this creative.
	AgencyId int64 `json:"agencyId,omitempty,string"`

	// ApiUploadTimestamp: The last upload timestamp of this creative if it
	// was uploaded via API. Read-only. The value of this field is
	// generated, and will be ignored for uploads. (formatted RFC 3339
	// timestamp).
	ApiUploadTimestamp string `json:"api_upload_timestamp,omitempty"`

	// Attribute: All attributes for the ads that may be shown from this
	// snippet.
	Attribute []int64 `json:"attribute,omitempty"`

	// BuyerCreativeId: A buyer-specific id identifying the creative in this
	// ad.
	BuyerCreativeId string `json:"buyerCreativeId,omitempty"`

	// ClickThroughUrl: The set of destination urls for the snippet.
	ClickThroughUrl []string `json:"clickThroughUrl,omitempty"`

	// Corrections: Shows any corrections that were applied to this
	// creative. Read-only. This field should not be set in requests.
	Corrections []*CreativeCorrections `json:"corrections,omitempty"`

	// DealsStatus: Top-level deals status. Read-only. This field should not
	// be set in requests. If disapproved, an entry for
	// auctionType=DIRECT_DEALS (or ALL) in servingRestrictions will also
	// exist. Note that this may be nuanced with other contextual
	// restrictions, in which case it may be preferable to read from
	// servingRestrictions directly.
	DealsStatus string `json:"dealsStatus,omitempty"`

	// FilteringReasons: The filtering reasons for the creative. Read-only.
	// This field should not be set in requests.
	FilteringReasons *CreativeFilteringReasons `json:"filteringReasons,omitempty"`

	// Height: Ad height.
	Height int64 `json:"height,omitempty"`

	// ImpressionTrackingUrl: The set of urls to be called to record an
	// impression.
	ImpressionTrackingUrl []string `json:"impressionTrackingUrl,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// NativeAd: If nativeAd is set, HTMLSnippet and videoURL should not be
	// set.
	NativeAd *CreativeNativeAd `json:"nativeAd,omitempty"`

	// OpenAuctionStatus: Top-level open auction status. Read-only. This
	// field should not be set in requests. If disapproved, an entry for
	// auctionType=OPEN_AUCTION (or ALL) in servingRestrictions will also
	// exist. Note that this may be nuanced with other contextual
	// restrictions, in which case it may be preferable to read from
	// ServingRestrictions directly.
	OpenAuctionStatus string `json:"openAuctionStatus,omitempty"`

	// ProductCategories: Detected product categories, if any. Read-only.
	// This field should not be set in requests.
	ProductCategories []int64 `json:"productCategories,omitempty"`

	// RestrictedCategories: All restricted categories for the ads that may
	// be shown from this snippet.
	RestrictedCategories []int64 `json:"restrictedCategories,omitempty"`

	// SensitiveCategories: Detected sensitive categories, if any.
	// Read-only. This field should not be set in requests.
	SensitiveCategories []int64 `json:"sensitiveCategories,omitempty"`

	// ServingRestrictions: The granular status of this ad in specific
	// contexts. A context here relates to where something ultimately serves
	// (for example, a physical location, a platform, an HTTPS vs HTTP
	// request, or the type of auction). Read-only. This field should not be
	// set in requests.
	ServingRestrictions []*CreativeServingRestrictions `json:"servingRestrictions,omitempty"`

	// VendorType: All vendor types for the ads that may be shown from this
	// snippet.
	VendorType []int64 `json:"vendorType,omitempty"`

	// Version: The version for this creative. Read-only. This field should
	// not be set in requests.
	Version int64 `json:"version,omitempty"`

	// VideoURL: The url to fetch a video ad. If set, HTMLSnippet should not
	// be set.
	VideoURL string `json:"videoURL,omitempty"`

	// Width: Ad width.
	Width int64 `json:"width,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "HTMLSnippet") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Creative) MarshalJSON() ([]byte, error) {
	type noMethod Creative
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreativeCorrections struct {
	// Details: Additional details about the correction.
	Details []string `json:"details,omitempty"`

	// Reason: The type of correction that was applied to the creative.
	Reason string `json:"reason,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Details") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeCorrections) MarshalJSON() ([]byte, error) {
	type noMethod CreativeCorrections
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativeFilteringReasons: The filtering reasons for the creative.
// Read-only. This field should not be set in requests.
type CreativeFilteringReasons struct {
	// Date: The date in ISO 8601 format for the data. The data is collected
	// from 00:00:00 to 23:59:59 in PST.
	Date string `json:"date,omitempty"`

	// Reasons: The filtering reasons.
	Reasons []*CreativeFilteringReasonsReasons `json:"reasons,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Date") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeFilteringReasons) MarshalJSON() ([]byte, error) {
	type noMethod CreativeFilteringReasons
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreativeFilteringReasonsReasons struct {
	// FilteringCount: The number of times the creative was filtered for the
	// status. The count is aggregated across all publishers on the
	// exchange.
	FilteringCount int64 `json:"filteringCount,omitempty,string"`

	// FilteringStatus: The filtering status code. Please refer to the
	// creative-status-codes.txt file for different statuses.
	FilteringStatus int64 `json:"filteringStatus,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FilteringCount") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeFilteringReasonsReasons) MarshalJSON() ([]byte, error) {
	type noMethod CreativeFilteringReasonsReasons
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativeNativeAd: If nativeAd is set, HTMLSnippet and videoURL should
// not be set.
type CreativeNativeAd struct {
	Advertiser string `json:"advertiser,omitempty"`

	// AppIcon: The app icon, for app download ads.
	AppIcon *CreativeNativeAdAppIcon `json:"appIcon,omitempty"`

	// Body: A long description of the ad.
	Body string `json:"body,omitempty"`

	// CallToAction: A label for the button that the user is supposed to
	// click.
	CallToAction string `json:"callToAction,omitempty"`

	// ClickTrackingUrl: The URL to use for click tracking.
	ClickTrackingUrl string `json:"clickTrackingUrl,omitempty"`

	// Headline: A short title for the ad.
	Headline string `json:"headline,omitempty"`

	// Image: A large image.
	Image *CreativeNativeAdImage `json:"image,omitempty"`

	// ImpressionTrackingUrl: The URLs are called when the impression is
	// rendered.
	ImpressionTrackingUrl []string `json:"impressionTrackingUrl,omitempty"`

	// Logo: A smaller image, for the advertiser logo.
	Logo *CreativeNativeAdLogo `json:"logo,omitempty"`

	// Price: The price of the promoted app including the currency info.
	Price string `json:"price,omitempty"`

	// StarRating: The app rating in the app store. Must be in the range
	// [0-5].
	StarRating float64 `json:"starRating,omitempty"`

	// Store: The URL to the app store to purchase/download the promoted
	// app.
	Store string `json:"store,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Advertiser") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeNativeAd) MarshalJSON() ([]byte, error) {
	type noMethod CreativeNativeAd
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativeNativeAdAppIcon: The app icon, for app download ads.
type CreativeNativeAdAppIcon struct {
	Height int64 `json:"height,omitempty"`

	Url string `json:"url,omitempty"`

	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeNativeAdAppIcon) MarshalJSON() ([]byte, error) {
	type noMethod CreativeNativeAdAppIcon
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativeNativeAdImage: A large image.
type CreativeNativeAdImage struct {
	Height int64 `json:"height,omitempty"`

	Url string `json:"url,omitempty"`

	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeNativeAdImage) MarshalJSON() ([]byte, error) {
	type noMethod CreativeNativeAdImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativeNativeAdLogo: A smaller image, for the advertiser logo.
type CreativeNativeAdLogo struct {
	Height int64 `json:"height,omitempty"`

	Url string `json:"url,omitempty"`

	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeNativeAdLogo) MarshalJSON() ([]byte, error) {
	type noMethod CreativeNativeAdLogo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreativeServingRestrictions struct {
	// Contexts: All known contexts/restrictions.
	Contexts []*CreativeServingRestrictionsContexts `json:"contexts,omitempty"`

	// DisapprovalReasons: The reasons for disapproval within this
	// restriction, if any. Note that not all disapproval reasons may be
	// categorized, so it is possible for the creative to have a status of
	// DISAPPROVED or CONDITIONALLY_APPROVED with an empty list for
	// disapproval_reasons. In this case, please reach out to your TAM to
	// help debug the issue.
	DisapprovalReasons []*CreativeServingRestrictionsDisapprovalReasons `json:"disapprovalReasons,omitempty"`

	// Reason: Why the creative is ineligible to serve in this context
	// (e.g., it has been explicitly disapproved or is pending review).
	Reason string `json:"reason,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Contexts") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeServingRestrictions) MarshalJSON() ([]byte, error) {
	type noMethod CreativeServingRestrictions
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreativeServingRestrictionsContexts struct {
	// AuctionType: Only set when contextType=AUCTION_TYPE. Represents the
	// auction types this restriction applies to.
	AuctionType []string `json:"auctionType,omitempty"`

	// ContextType: The type of context (e.g., location, platform, auction
	// type, SSL-ness).
	ContextType string `json:"contextType,omitempty"`

	// GeoCriteriaId: Only set when contextType=LOCATION. Represents the geo
	// criterias this restriction applies to.
	GeoCriteriaId []int64 `json:"geoCriteriaId,omitempty"`

	// Platform: Only set when contextType=PLATFORM. Represents the
	// platforms this restriction applies to.
	Platform []string `json:"platform,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AuctionType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeServingRestrictionsContexts) MarshalJSON() ([]byte, error) {
	type noMethod CreativeServingRestrictionsContexts
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CreativeServingRestrictionsDisapprovalReasons struct {
	// Details: Additional details about the reason for disapproval.
	Details []string `json:"details,omitempty"`

	// Reason: The categorized reason for disapproval.
	Reason string `json:"reason,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Details") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativeServingRestrictionsDisapprovalReasons) MarshalJSON() ([]byte, error) {
	type noMethod CreativeServingRestrictionsDisapprovalReasons
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CreativesList: The creatives feed lists the active creatives for the
// Ad Exchange buyer accounts that the user has access to. Each entry in
// the feed corresponds to a single creative.
type CreativesList struct {
	// Items: A list of creatives.
	Items []*Creative `json:"items,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// NextPageToken: Continuation token used to page through creatives. To
	// retrieve the next page of results, set the next request's "pageToken"
	// value to this.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Items") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreativesList) MarshalJSON() ([]byte, error) {
	type noMethod CreativesList
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DateTime struct {
	Day int64 `json:"day,omitempty"`

	Hour int64 `json:"hour,omitempty"`

	Minute int64 `json:"minute,omitempty"`

	Month int64 `json:"month,omitempty"`

	Second int64 `json:"second,omitempty"`

	TimeZoneId string `json:"timeZoneId,omitempty"`

	Year int64 `json:"year,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Day") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DateTime) MarshalJSON() ([]byte, error) {
	type noMethod DateTime
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DealPartyDto struct {
	Buyer *BuyerDto `json:"buyer,omitempty"`

	BuyerSellerRole string `json:"buyerSellerRole,omitempty"`

	CustomerId int64 `json:"customerId,omitempty"`

	Name string `json:"name,omitempty"`

	WebProperty *WebPropertyDto `json:"webProperty,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Buyer") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DealPartyDto) MarshalJSON() ([]byte, error) {
	type noMethod DealPartyDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DealTerms struct {
	// Description: Description for the proposed terms of the deal.
	Description string `json:"description,omitempty"`

	// GuaranteedFixedPriceTerms: The terms for guaranteed fixed price
	// deals.
	GuaranteedFixedPriceTerms *DealTermsGuaranteedFixedPriceTerms `json:"guaranteedFixedPriceTerms,omitempty"`

	// NonGuaranteedAuctionTerms: The terms for non-guaranteed auction
	// deals.
	NonGuaranteedAuctionTerms *DealTermsNonGuaranteedAuctionTerms `json:"nonGuaranteedAuctionTerms,omitempty"`

	// NonGuaranteedFixedPriceTerms: The terms for non-guaranteed fixed
	// price deals.
	NonGuaranteedFixedPriceTerms *DealTermsNonGuaranteedFixedPriceTerms `json:"nonGuaranteedFixedPriceTerms,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DealTerms) MarshalJSON() ([]byte, error) {
	type noMethod DealTerms
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DealTermsGuaranteedFixedPriceTerms struct {
	// FixedPrices: Fixed price for the specified buyer.
	FixedPrices []*PricePerBuyer `json:"fixedPrices,omitempty"`

	// GuaranteedImpressions: Guaranteed impressions as a percentage. This
	// is the percentage of guaranteed looks that the buyer is guaranteeing
	// to buy.
	GuaranteedImpressions int64 `json:"guaranteedImpressions,omitempty,string"`

	// GuaranteedLooks: Count of guaranteed looks. Required for deal,
	// optional for offer.
	GuaranteedLooks int64 `json:"guaranteedLooks,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "FixedPrices") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DealTermsGuaranteedFixedPriceTerms) MarshalJSON() ([]byte, error) {
	type noMethod DealTermsGuaranteedFixedPriceTerms
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DealTermsNonGuaranteedAuctionTerms struct {
	// PrivateAuctionId: Id of the corresponding private auction.
	PrivateAuctionId string `json:"privateAuctionId,omitempty"`

	// ReservePricePerBuyers: Reserve price for the specified buyer.
	ReservePricePerBuyers []*PricePerBuyer `json:"reservePricePerBuyers,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PrivateAuctionId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DealTermsNonGuaranteedAuctionTerms) MarshalJSON() ([]byte, error) {
	type noMethod DealTermsNonGuaranteedAuctionTerms
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DealTermsNonGuaranteedFixedPriceTerms struct {
	// FixedPrices: Fixed price for the specified buyer.
	FixedPrices []*PricePerBuyer `json:"fixedPrices,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FixedPrices") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DealTermsNonGuaranteedFixedPriceTerms) MarshalJSON() ([]byte, error) {
	type noMethod DealTermsNonGuaranteedFixedPriceTerms
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DeleteOrderDealsRequest struct {
	// DealIds: List of deals to delete for a given order
	DealIds []string `json:"dealIds,omitempty"`

	// OrderRevisionNumber: The last known order revision number.
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	UpdateAction string `json:"updateAction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DealIds") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DeleteOrderDealsRequest) MarshalJSON() ([]byte, error) {
	type noMethod DeleteOrderDealsRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DeleteOrderDealsResponse struct {
	// Deals: List of deals deleted (in the same order as passed in the
	// request)
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// OrderRevisionNumber: The updated revision number for the order.
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DeleteOrderDealsResponse) MarshalJSON() ([]byte, error) {
	type noMethod DeleteOrderDealsResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DeliveryControl struct {
	DeliveryRateType string `json:"deliveryRateType,omitempty"`

	FrequencyCaps []*DeliveryControlFrequencyCap `json:"frequencyCaps,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeliveryRateType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DeliveryControl) MarshalJSON() ([]byte, error) {
	type noMethod DeliveryControl
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type DeliveryControlFrequencyCap struct {
	MaxImpressions int64 `json:"maxImpressions,omitempty"`

	NumTimeUnits int64 `json:"numTimeUnits,omitempty"`

	TimeUnitType string `json:"timeUnitType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaxImpressions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DeliveryControlFrequencyCap) MarshalJSON() ([]byte, error) {
	type noMethod DeliveryControlFrequencyCap
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type EditAllOrderDealsRequest struct {
	// Deals: List of deals to edit. Service may perform 3 different
	// operations based on comparison of deals in this list vs deals already
	// persisted in database: 1. Add new deal to order If a deal in this
	// list does not exist in the order, the service will create a new deal
	// and add it to the order. Validation will follow AddOrderDealsRequest.
	// 2. Update existing deal in the order If a deal in this list already
	// exist in the order, the service will update that existing deal to
	// this new deal in the request. Validation will follow
	// UpdateOrderDealsRequest. 3. Delete deals from the order (just need
	// the id) If a existing deal in the order is not present in this list,
	// the service will delete that deal from the order. Validation will
	// follow DeleteOrderDealsRequest.
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// Order: If specified, also updates the order in the batch transaction.
	// This is useful when the order and the deals need to be updated in one
	// transaction.
	Order *MarketplaceOrder `json:"order,omitempty"`

	// OrderRevisionNumber: The last known revision number for the order.
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	// UpdateAction: Indicates an optional action to take on the order
	UpdateAction string `json:"updateAction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *EditAllOrderDealsRequest) MarshalJSON() ([]byte, error) {
	type noMethod EditAllOrderDealsRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type EditAllOrderDealsResponse struct {
	// Deals: List of all deals in the order after edit.
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *EditAllOrderDealsResponse) MarshalJSON() ([]byte, error) {
	type noMethod EditAllOrderDealsResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type EditHistoryDto struct {
	CreatedByLoginName string `json:"createdByLoginName,omitempty"`

	CreatedTimeStamp int64 `json:"createdTimeStamp,omitempty,string"`

	LastUpdateTimeStamp int64 `json:"lastUpdateTimeStamp,omitempty,string"`

	LastUpdatedByLoginName string `json:"lastUpdatedByLoginName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CreatedByLoginName")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *EditHistoryDto) MarshalJSON() ([]byte, error) {
	type noMethod EditHistoryDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetFinalizedNegotiationByExternalDealIdRequest struct {
	IncludePrivateAuctions bool `json:"includePrivateAuctions,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "IncludePrivateAuctions") to unconditionally include in API requests.
	// By default, fields with empty values are omitted from API requests.
	// However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetFinalizedNegotiationByExternalDealIdRequest) MarshalJSON() ([]byte, error) {
	type noMethod GetFinalizedNegotiationByExternalDealIdRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetNegotiationByIdRequest struct {
	IncludePrivateAuctions bool `json:"includePrivateAuctions,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "IncludePrivateAuctions") to unconditionally include in API requests.
	// By default, fields with empty values are omitted from API requests.
	// However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetNegotiationByIdRequest) MarshalJSON() ([]byte, error) {
	type noMethod GetNegotiationByIdRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetNegotiationsRequest struct {
	Finalized bool `json:"finalized,omitempty"`

	IncludePrivateAuctions bool `json:"includePrivateAuctions,omitempty"`

	SinceTimestampMillis int64 `json:"sinceTimestampMillis,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "Finalized") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetNegotiationsRequest) MarshalJSON() ([]byte, error) {
	type noMethod GetNegotiationsRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetNegotiationsResponse struct {
	Kind string `json:"kind,omitempty"`

	Negotiations []*NegotiationDto `json:"negotiations,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Kind") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetNegotiationsResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetNegotiationsResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetOffersResponse struct {
	// Offers: The returned list of offers.
	Offers []*MarketplaceOffer `json:"offers,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Offers") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetOffersResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetOffersResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetOrderDealsResponse struct {
	// Deals: List of deals for the order
	Deals []*MarketplaceDeal `json:"deals,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Deals") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetOrderDealsResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetOrderDealsResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetOrderNotesResponse struct {
	// Notes: The list of matching notes.
	Notes []*MarketplaceNote `json:"notes,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Notes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetOrderNotesResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetOrderNotesResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type GetOrdersResponse struct {
	// Orders: The list of matching orders.
	Orders []*MarketplaceOrder `json:"orders,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Orders") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetOrdersResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetOrdersResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type InventorySegmentTargeting struct {
	NegativeAdSizes []*AdSize `json:"negativeAdSizes,omitempty"`

	NegativeAdTypeSegments []string `json:"negativeAdTypeSegments,omitempty"`

	NegativeAudienceSegments googleapi.Int64s `json:"negativeAudienceSegments,omitempty"`

	NegativeDeviceCategories googleapi.Int64s `json:"negativeDeviceCategories,omitempty"`

	NegativeIcmBrands googleapi.Int64s `json:"negativeIcmBrands,omitempty"`

	NegativeIcmInterests googleapi.Int64s `json:"negativeIcmInterests,omitempty"`

	NegativeInventorySlots []string `json:"negativeInventorySlots,omitempty"`

	NegativeKeyValues []*RuleKeyValuePair `json:"negativeKeyValues,omitempty"`

	NegativeLocations googleapi.Int64s `json:"negativeLocations,omitempty"`

	NegativeMobileApps []string `json:"negativeMobileApps,omitempty"`

	NegativeOperatingSystemVersions googleapi.Int64s `json:"negativeOperatingSystemVersions,omitempty"`

	NegativeOperatingSystems googleapi.Int64s `json:"negativeOperatingSystems,omitempty"`

	NegativeSiteUrls []string `json:"negativeSiteUrls,omitempty"`

	NegativeSizes googleapi.Int64s `json:"negativeSizes,omitempty"`

	NegativeVideoAdPositionSegments []string `json:"negativeVideoAdPositionSegments,omitempty"`

	NegativeVideoDurationSegments googleapi.Int64s `json:"negativeVideoDurationSegments,omitempty"`

	NegativeXfpAdSlots googleapi.Int64s `json:"negativeXfpAdSlots,omitempty"`

	NegativeXfpPlacements googleapi.Int64s `json:"negativeXfpPlacements,omitempty"`

	PositiveAdSizes []*AdSize `json:"positiveAdSizes,omitempty"`

	PositiveAdTypeSegments []string `json:"positiveAdTypeSegments,omitempty"`

	PositiveAudienceSegments googleapi.Int64s `json:"positiveAudienceSegments,omitempty"`

	PositiveDeviceCategories googleapi.Int64s `json:"positiveDeviceCategories,omitempty"`

	PositiveIcmBrands googleapi.Int64s `json:"positiveIcmBrands,omitempty"`

	PositiveIcmInterests googleapi.Int64s `json:"positiveIcmInterests,omitempty"`

	PositiveInventorySlots []string `json:"positiveInventorySlots,omitempty"`

	PositiveKeyValues []*RuleKeyValuePair `json:"positiveKeyValues,omitempty"`

	PositiveLocations googleapi.Int64s `json:"positiveLocations,omitempty"`

	PositiveMobileApps []string `json:"positiveMobileApps,omitempty"`

	PositiveOperatingSystemVersions googleapi.Int64s `json:"positiveOperatingSystemVersions,omitempty"`

	PositiveOperatingSystems googleapi.Int64s `json:"positiveOperatingSystems,omitempty"`

	PositiveSiteUrls []string `json:"positiveSiteUrls,omitempty"`

	PositiveSizes googleapi.Int64s `json:"positiveSizes,omitempty"`

	PositiveVideoAdPositionSegments []string `json:"positiveVideoAdPositionSegments,omitempty"`

	PositiveVideoDurationSegments googleapi.Int64s `json:"positiveVideoDurationSegments,omitempty"`

	PositiveXfpAdSlots googleapi.Int64s `json:"positiveXfpAdSlots,omitempty"`

	PositiveXfpPlacements googleapi.Int64s `json:"positiveXfpPlacements,omitempty"`

	// ForceSendFields is a list of field names (e.g. "NegativeAdSizes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *InventorySegmentTargeting) MarshalJSON() ([]byte, error) {
	type noMethod InventorySegmentTargeting
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ListClientAccessCapabilitiesRequest struct {
	SponsorAccountId int64 `json:"sponsorAccountId,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "SponsorAccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClientAccessCapabilitiesRequest) MarshalJSON() ([]byte, error) {
	type noMethod ListClientAccessCapabilitiesRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ListClientAccessCapabilitiesResponse struct {
	ClientAccessPermissions []*ClientAccessCapabilities `json:"clientAccessPermissions,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g.
	// "ClientAccessPermissions") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClientAccessCapabilitiesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListClientAccessCapabilitiesResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ListOffersRequest struct {
	SinceTimestampMillis int64 `json:"sinceTimestampMillis,omitempty,string"`

	// ForceSendFields is a list of field names (e.g.
	// "SinceTimestampMillis") to unconditionally include in API requests.
	// By default, fields with empty values are omitted from API requests.
	// However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListOffersRequest) MarshalJSON() ([]byte, error) {
	type noMethod ListOffersRequest
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ListOffersResponse struct {
	Kind string `json:"kind,omitempty"`

	Offers []*OfferDto `json:"offers,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Kind") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListOffersResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListOffersResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// MarketplaceDeal: An order can contain multiple deals. A deal contains
// the terms and targeting information that is used for serving.
type MarketplaceDeal struct {
	// BuyerPrivateData: Buyer private data (hidden from seller).
	BuyerPrivateData *PrivateData `json:"buyerPrivateData,omitempty"`

	// CreationTimeMs: The time (ms since epoch) of the deal creation.
	// (readonly)
	CreationTimeMs int64 `json:"creationTimeMs,omitempty,string"`

	// DealId: A unique deal=id for the deal (readonly).
	DealId string `json:"dealId,omitempty"`

	// DeliveryControl: The set of fields around delivery control that are
	// interesting for a buyer to see but are non-negotiable. These are set
	// by the publisher. This message is assigned an id of 100 since some
	// day we would want to model this as a protobuf extension.
	DeliveryControl *DeliveryControl `json:"deliveryControl,omitempty"`

	// ExternalDealId: The external deal id assigned to this deal once the
	// deal is finalized. This is the deal-id that shows up in
	// serving/reporting etc. (readonly)
	ExternalDealId string `json:"externalDealId,omitempty"`

	// FlightEndTimeMs: Proposed flight end time of the deal (ms since
	// epoch) This will generally be stored in a granularity of a second.
	// (updatable)
	FlightEndTimeMs int64 `json:"flightEndTimeMs,omitempty,string"`

	// FlightStartTimeMs: Proposed flight start time of the deal (ms since
	// epoch) This will generally be stored in a granularity of a second.
	// (updatable)
	FlightStartTimeMs int64 `json:"flightStartTimeMs,omitempty,string"`

	// InventoryDescription: Description for the deal terms. (updatable)
	InventoryDescription string `json:"inventoryDescription,omitempty"`

	// Kind: Identifies what kind of resource this is. Value: the fixed
	// string "adexchangebuyer#marketplaceDeal".
	Kind string `json:"kind,omitempty"`

	// LastUpdateTimeMs: The time (ms since epoch) when the deal was last
	// updated. (readonly)
	LastUpdateTimeMs int64 `json:"lastUpdateTimeMs,omitempty,string"`

	// Name: The name of the deal. (updatable)
	Name string `json:"name,omitempty"`

	// OfferId: The offer-id from which this deal was created. (readonly,
	// except on create)
	OfferId string `json:"offerId,omitempty"`

	// OfferRevisionNumber: The revision number of the offer that the deal
	// was created from (readonly, except on create)
	OfferRevisionNumber int64 `json:"offerRevisionNumber,omitempty,string"`

	OrderId string `json:"orderId,omitempty"`

	// SellerContacts: Optional Seller contact information for the deal
	// (buyer-readonly)
	SellerContacts []*ContactInformation `json:"sellerContacts,omitempty"`

	// SharedTargetings: The shared targeting visible to buyers and sellers.
	// (updatable)
	SharedTargetings []*SharedTargeting `json:"sharedTargetings,omitempty"`

	// SyndicationProduct: The syndication product associated with the deal.
	// (readonly, except on create)
	SyndicationProduct string `json:"syndicationProduct,omitempty"`

	// Terms: The negotiable terms of the deal. (updatable)
	Terms *DealTerms `json:"terms,omitempty"`

	WebPropertyCode string `json:"webPropertyCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BuyerPrivateData") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceDeal) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceDeal
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type MarketplaceDealParty struct {
	// Buyer: The buyer/seller associated with the deal. One of buyer/seller
	// is specified for a deal-party.
	Buyer *Buyer `json:"buyer,omitempty"`

	// Seller: The buyer/seller associated with the deal. One of
	// buyer/seller is specified for a deal party.
	Seller *Seller `json:"seller,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Buyer") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceDealParty) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceDealParty
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type MarketplaceLabel struct {
	// AccountId: The accountId of the party that created the label.
	AccountId string `json:"accountId,omitempty"`

	// CreateTimeMs: The creation time (in ms since epoch) for the label.
	CreateTimeMs int64 `json:"createTimeMs,omitempty,string"`

	// DeprecatedMarketplaceDealParty: Information about the party that
	// created the label.
	DeprecatedMarketplaceDealParty *MarketplaceDealParty `json:"deprecatedMarketplaceDealParty,omitempty"`

	// Label: The label to use.
	Label string `json:"label,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceLabel) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceLabel
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// MarketplaceNote: An order is associated with a bunch of notes which
// may optionally be associated with a deal and/or revision number.
type MarketplaceNote struct {
	// CreatorRole: The role of the person (buyer/seller) creating the note.
	// (readonly)
	CreatorRole string `json:"creatorRole,omitempty"`

	// DealId: Notes can optionally be associated with a deal. (readonly,
	// except on create)
	DealId string `json:"dealId,omitempty"`

	// Kind: Identifies what kind of resource this is. Value: the fixed
	// string "adexchangebuyer#marketplaceNote".
	Kind string `json:"kind,omitempty"`

	// Note: The actual note to attach. (readonly, except on create)
	Note string `json:"note,omitempty"`

	// NoteId: The unique id for the note. (readonly)
	NoteId string `json:"noteId,omitempty"`

	// OrderId: The order_id that a note is attached to. (readonly)
	OrderId string `json:"orderId,omitempty"`

	// OrderRevisionNumber: If the note is associated with an order revision
	// number, then store that here. (readonly, except on create)
	OrderRevisionNumber int64 `json:"orderRevisionNumber,omitempty,string"`

	// TimestampMs: The timestamp (ms since epoch) that this note was
	// created. (readonly)
	TimestampMs int64 `json:"timestampMs,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "CreatorRole") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceNote) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceNote
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// MarketplaceOffer: An offer is segment of inventory that a seller
// wishes to sell. It is associated with certain terms and targeting
// information which helps buyer know more about the inventory. Each
// field in an order can have one of the following setting:
//
// (readonly) - It is an error to try and set this field.
// (buyer-readonly) - Only the seller can set this field.
// (seller-readonly) - Only the buyer can set this field. (updatable) -
// The field is updatable at all times by either buyer or the seller.
type MarketplaceOffer struct {
	// CreationTimeMs: Creation time in ms. since epoch (readonly)
	CreationTimeMs int64 `json:"creationTimeMs,omitempty,string"`

	// CreatorContacts: Optional contact information for the creator of this
	// offer. (buyer-readonly)
	CreatorContacts []*ContactInformation `json:"creatorContacts,omitempty"`

	// FlightEndTimeMs: The proposed end time for the deal (ms since epoch)
	// (buyer-readonly)
	FlightEndTimeMs int64 `json:"flightEndTimeMs,omitempty,string"`

	// FlightStartTimeMs: Inventory availability dates. (times are in ms
	// since epoch) The granularity is generally in the order of seconds.
	// (buyer-readonly)
	FlightStartTimeMs int64 `json:"flightStartTimeMs,omitempty,string"`

	// HasCreatorSignedOff: If the creator has already signed off on the
	// offer, then the buyer can finalize the deal by accepting the offer as
	// is. When copying to an order, if any of the terms are changed, then
	// auto_finalize is automatically set to false.
	HasCreatorSignedOff bool `json:"hasCreatorSignedOff,omitempty"`

	// Kind: Identifies what kind of resource this is. Value: the fixed
	// string "adexchangebuyer#marketplaceOffer".
	Kind string `json:"kind,omitempty"`

	// Labels: Optional List of labels for the offer (optional,
	// buyer-readonly).
	Labels []*MarketplaceLabel `json:"labels,omitempty"`

	// LastUpdateTimeMs: Time of last update in ms. since epoch (readonly)
	LastUpdateTimeMs int64 `json:"lastUpdateTimeMs,omitempty,string"`

	// Name: The name for this offer as set by the seller. (buyer-readonly)
	Name string `json:"name,omitempty"`

	// OfferId: The unique id for the offer (readonly)
	OfferId string `json:"offerId,omitempty"`

	// RevisionNumber: The revision number of the offer. (readonly)
	RevisionNumber int64 `json:"revisionNumber,omitempty,string"`

	// Seller: Information about the seller that created this offer
	// (readonly, except on create)
	Seller *Seller `json:"seller,omitempty"`

	// SharedTargetings: Targeting that is shared between the buyer and the
	// seller. Each targeting criteria has a specified key and for each key
	// there is a list of inclusion value or exclusion values.
	// (buyer-readonly)
	SharedTargetings []*SharedTargeting `json:"sharedTargetings,omitempty"`

	// State: The state of the offer. (buyer-readonly)
	State string `json:"state,omitempty"`

	// SyndicationProduct: The syndication product associated with the deal.
	// (readonly, except on create)
	SyndicationProduct string `json:"syndicationProduct,omitempty"`

	// Terms: The negotiable terms of the deal (buyer-readonly)
	Terms *DealTerms `json:"terms,omitempty"`

	WebPropertyCode string `json:"webPropertyCode,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "CreationTimeMs") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceOffer) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceOffer
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// MarketplaceOrder: Represents an order in the marketplace. An order is
// the unit of negotiation between a seller and a buyer and contains
// deals which are served. Each field in an order can have one of the
// following setting:
//
// (readonly) - It is an error to try and set this field.
// (buyer-readonly) - Only the seller can set this field.
// (seller-readonly) - Only the buyer can set this field. (updatable) -
// The field is updatable at all times by either buyer or the seller.
type MarketplaceOrder struct {
	// BilledBuyer: Reference to the buyer that will get billed for this
	// order. (readonly)
	BilledBuyer *Buyer `json:"billedBuyer,omitempty"`

	// Buyer: Reference to the buyer on the order. (readonly, except on
	// create)
	Buyer *Buyer `json:"buyer,omitempty"`

	// BuyerContacts: Optional contact information fort the buyer.
	// (seller-readonly)
	BuyerContacts []*ContactInformation `json:"buyerContacts,omitempty"`

	// BuyerPrivateData: Private data for buyer. (hidden from seller).
	BuyerPrivateData *PrivateData `json:"buyerPrivateData,omitempty"`

	// HasBuyerSignedOff: When an order is in an accepted state, indicates
	// whether the buyer has signed off Once both sides have signed off on a
	// deal, the order can be finalized by the seller. (seller-readonly)
	HasBuyerSignedOff bool `json:"hasBuyerSignedOff,omitempty"`

	// HasSellerSignedOff: When an order is in an accepted state, indicates
	// whether the buyer has signed off Once both sides have signed off on a
	// deal, the order can be finalized by the seller. (buyer-readonly)
	HasSellerSignedOff bool `json:"hasSellerSignedOff,omitempty"`

	// IsRenegotiating: True if the order is being renegotiated (readonly).
	IsRenegotiating bool `json:"isRenegotiating,omitempty"`

	// IsSetupComplete: True, if the buyside inventory setup is complete for
	// this order. (readonly)
	IsSetupComplete bool `json:"isSetupComplete,omitempty"`

	// Kind: Identifies what kind of resource this is. Value: the fixed
	// string "adexchangebuyer#marketplaceOrder".
	Kind string `json:"kind,omitempty"`

	// Labels: List of labels associated with the order. (readonly)
	Labels []*MarketplaceLabel `json:"labels,omitempty"`

	// LastUpdaterOrCommentorRole: The role of the last user that either
	// updated the order or left a comment. (readonly)
	LastUpdaterOrCommentorRole string `json:"lastUpdaterOrCommentorRole,omitempty"`

	LastUpdaterRole string `json:"lastUpdaterRole,omitempty"`

	// Name: The name for the order (updatable)
	Name string `json:"name,omitempty"`

	// OrderId: The unique id of the order. (readonly).
	OrderId string `json:"orderId,omitempty"`

	// OrderState: The current state of the order. (readonly)
	OrderState string `json:"orderState,omitempty"`

	// OriginatorRole: Indicates whether the buyer/seller created the
	// offer.(readonly)
	OriginatorRole string `json:"originatorRole,omitempty"`

	// RevisionNumber: The revision number for the order (readonly).
	RevisionNumber int64 `json:"revisionNumber,omitempty,string"`

	// RevisionTimeMs: The time (ms since epoch) when the order was last
	// revised (readonly).
	RevisionTimeMs int64 `json:"revisionTimeMs,omitempty,string"`

	// Seller: Reference to the seller on the order. (readonly, except on
	// create)
	Seller *Seller `json:"seller,omitempty"`

	// SellerContacts: Optional contact information for the seller
	// (buyer-readonly).
	SellerContacts []*ContactInformation `json:"sellerContacts,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "BilledBuyer") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MarketplaceOrder) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceOrder
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type MoneyDto struct {
	CurrencyCode string `json:"currencyCode,omitempty"`

	Micros int64 `json:"micros,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "CurrencyCode") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MoneyDto) MarshalJSON() ([]byte, error) {
	type noMethod MoneyDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type NegotiationDto struct {
	// BilledBuyer: The billed buyer; Specified by a buyer buying through an
	// intermediary.
	BilledBuyer *DealPartyDto `json:"billedBuyer,omitempty"`

	// Buyer: Details of the buyer party in this negotiation.
	Buyer *DealPartyDto `json:"buyer,omitempty"`

	// BuyerEmailContacts: The buyer party's contact email.
	BuyerEmailContacts []string `json:"buyerEmailContacts,omitempty"`

	// DealType: The type of this deal.
	DealType string `json:"dealType,omitempty"`

	// ExternalDealId: For finalized negotiations, the ID of the finalized
	// deal.
	ExternalDealId int64 `json:"externalDealId,omitempty,string"`

	Kind string `json:"kind,omitempty"`

	// LabelNames: A list of label names applicable to this negotiation.
	LabelNames []string `json:"labelNames,omitempty"`

	// NegotiationId: The unique ID of this negotiation.
	NegotiationId int64 `json:"negotiationId,omitempty,string"`

	// NegotiationRounds: The series of negotiation rounds for this
	// negotiation.
	NegotiationRounds []*NegotiationRoundDto `json:"negotiationRounds,omitempty"`

	// NegotiationState: The state of this negotiation.
	NegotiationState string `json:"negotiationState,omitempty"`

	// OfferId: The ID of this negotiation's original offer.
	OfferId int64 `json:"offerId,omitempty,string"`

	// Seller: Details of the seller party in this negotiation.
	Seller *DealPartyDto `json:"seller,omitempty"`

	// SellerEmailContacts: The seller party's contact email.
	SellerEmailContacts []string `json:"sellerEmailContacts,omitempty"`

	// Stats: The stats for this negotiation.
	Stats *StatsDto `json:"stats,omitempty"`

	// Status: The status of this negotiation.
	Status string `json:"status,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "BilledBuyer") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *NegotiationDto) MarshalJSON() ([]byte, error) {
	type noMethod NegotiationDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type NegotiationRoundDto struct {
	// Action: The action performed by this negotiation round.
	Action string `json:"action,omitempty"`

	// DbmPartnerId: Stores DBM partner ID for use by DBM
	DbmPartnerId int64 `json:"dbmPartnerId,omitempty,string"`

	// EditHistory: The edit history of this negotiation round.
	EditHistory *EditHistoryDto `json:"editHistory,omitempty"`

	Kind string `json:"kind,omitempty"`

	// NegotiationId: The ID of the negotiation to which this negotiation
	// round applies.
	NegotiationId int64 `json:"negotiationId,omitempty,string"`

	// Notes: Notes regarding this negotiation round.
	Notes string `json:"notes,omitempty"`

	// OriginatorRole: The role, either buyer or seller, initiating this
	// negotiation round.
	OriginatorRole string `json:"originatorRole,omitempty"`

	// RoundNumber: The number of this negotiation round, in sequence.
	RoundNumber int64 `json:"roundNumber,omitempty,string"`

	// Terms: The detailed terms proposed in this negotiation round.
	Terms *TermsDto `json:"terms,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Action") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *NegotiationRoundDto) MarshalJSON() ([]byte, error) {
	type noMethod NegotiationRoundDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type OfferDto struct {
	// Anonymous: Whether this offer is anonymous.
	Anonymous bool `json:"anonymous,omitempty"`

	// BilledBuyer: The billed buyer; For buyer initiated offers, buying
	// through an intermediary.
	BilledBuyer *DealPartyDto `json:"billedBuyer,omitempty"`

	// ClosedToDealParties: The list of buyer or seller parties this offer
	// is closed to.
	ClosedToDealParties []*DealPartyDto `json:"closedToDealParties,omitempty"`

	// Creator: The creator of this offer.
	Creator *DealPartyDto `json:"creator,omitempty"`

	// EmailContacts: The list of email contacts for this offer.
	EmailContacts []string `json:"emailContacts,omitempty"`

	// IsOpen: Whether this offer is open.
	IsOpen bool `json:"isOpen,omitempty"`

	Kind string `json:"kind,omitempty"`

	// LabelNames: The list of label names applicable to this offer.
	LabelNames []string `json:"labelNames,omitempty"`

	// OfferId: The unique ID of this offer.
	OfferId int64 `json:"offerId,omitempty,string"`

	// OfferState: The state of this offer.
	OfferState string `json:"offerState,omitempty"`

	// OpenToDealParties: The list of buyer or seller parties this offer is
	// open to.
	OpenToDealParties []*DealPartyDto `json:"openToDealParties,omitempty"`

	// PointOfContact: The point of contact for this offer.
	PointOfContact string `json:"pointOfContact,omitempty"`

	// Status: The status of this offer.
	Status string `json:"status,omitempty"`

	// Terms: The terms of this offer.
	Terms *TermsDto `json:"terms,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Anonymous") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *OfferDto) MarshalJSON() ([]byte, error) {
	type noMethod OfferDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PerformanceReport: The configuration data for an Ad Exchange
// performance report list.
type PerformanceReport struct {
	// BidRate: The number of bid responses with an ad.
	BidRate float64 `json:"bidRate,omitempty"`

	// BidRequestRate: The number of bid requests sent to your bidder.
	BidRequestRate float64 `json:"bidRequestRate,omitempty"`

	// CalloutStatusRate: Rate of various prefiltering statuses per match.
	// Please refer to the callout-status-codes.txt file for different
	// statuses.
	CalloutStatusRate []interface{} `json:"calloutStatusRate,omitempty"`

	// CookieMatcherStatusRate: Average QPS for cookie matcher operations.
	CookieMatcherStatusRate []interface{} `json:"cookieMatcherStatusRate,omitempty"`

	// CreativeStatusRate: Rate of ads with a given status. Please refer to
	// the creative-status-codes.txt file for different statuses.
	CreativeStatusRate []interface{} `json:"creativeStatusRate,omitempty"`

	// FilteredBidRate: The number of bid responses that were filtered due
	// to a policy violation or other errors.
	FilteredBidRate float64 `json:"filteredBidRate,omitempty"`

	// HostedMatchStatusRate: Average QPS for hosted match operations.
	HostedMatchStatusRate []interface{} `json:"hostedMatchStatusRate,omitempty"`

	// InventoryMatchRate: The number of potential queries based on your
	// pretargeting settings.
	InventoryMatchRate float64 `json:"inventoryMatchRate,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// Latency50thPercentile: The 50th percentile round trip latency(ms) as
	// perceived from Google servers for the duration period covered by the
	// report.
	Latency50thPercentile float64 `json:"latency50thPercentile,omitempty"`

	// Latency85thPercentile: The 85th percentile round trip latency(ms) as
	// perceived from Google servers for the duration period covered by the
	// report.
	Latency85thPercentile float64 `json:"latency85thPercentile,omitempty"`

	// Latency95thPercentile: The 95th percentile round trip latency(ms) as
	// perceived from Google servers for the duration period covered by the
	// report.
	Latency95thPercentile float64 `json:"latency95thPercentile,omitempty"`

	// NoQuotaInRegion: Rate of various quota account statuses per quota
	// check.
	NoQuotaInRegion float64 `json:"noQuotaInRegion,omitempty"`

	// OutOfQuota: Rate of various quota account statuses per quota check.
	OutOfQuota float64 `json:"outOfQuota,omitempty"`

	// PixelMatchRequests: Average QPS for pixel match requests from
	// clients.
	PixelMatchRequests float64 `json:"pixelMatchRequests,omitempty"`

	// PixelMatchResponses: Average QPS for pixel match responses from
	// clients.
	PixelMatchResponses float64 `json:"pixelMatchResponses,omitempty"`

	// QuotaConfiguredLimit: The configured quota limits for this account.
	QuotaConfiguredLimit float64 `json:"quotaConfiguredLimit,omitempty"`

	// QuotaThrottledLimit: The throttled quota limits for this account.
	QuotaThrottledLimit float64 `json:"quotaThrottledLimit,omitempty"`

	// Region: The trading location of this data.
	Region string `json:"region,omitempty"`

	// SuccessfulRequestRate: The number of properly formed bid responses
	// received by our servers within the deadline.
	SuccessfulRequestRate float64 `json:"successfulRequestRate,omitempty"`

	// Timestamp: The unix timestamp of the starting time of this
	// performance data.
	Timestamp int64 `json:"timestamp,omitempty,string"`

	// UnsuccessfulRequestRate: The number of bid responses that were
	// unsuccessful due to timeouts, incorrect formatting, etc.
	UnsuccessfulRequestRate float64 `json:"unsuccessfulRequestRate,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BidRate") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PerformanceReport) MarshalJSON() ([]byte, error) {
	type noMethod PerformanceReport
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PerformanceReportList: The configuration data for an Ad Exchange
// performance report list.
// https://sites.google.com/a/google.com/adx-integration/Home/engineering/binary-releases/rtb-api-release
// https://cs.corp.google.com/#piper///depot/google3/contentads/adx/tools/rtb_api/adxrtb.py
type PerformanceReportList struct {
	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// PerformanceReport: A list of performance reports relevant for the
	// account.
	PerformanceReport []*PerformanceReport `json:"performanceReport,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Kind") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PerformanceReportList) MarshalJSON() ([]byte, error) {
	type noMethod PerformanceReportList
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PretargetingConfig struct {
	// BillingId: The id for billing purposes, provided for reference. Leave
	// this field blank for insert requests; the id will be generated
	// automatically.
	BillingId int64 `json:"billingId,omitempty,string"`

	// ConfigId: The config id; generated automatically. Leave this field
	// blank for insert requests.
	ConfigId int64 `json:"configId,omitempty,string"`

	// ConfigName: The name of the config. Must be unique. Required for all
	// requests.
	ConfigName string `json:"configName,omitempty"`

	// CreativeType: List must contain exactly one of
	// PRETARGETING_CREATIVE_TYPE_HTML or PRETARGETING_CREATIVE_TYPE_VIDEO.
	CreativeType []string `json:"creativeType,omitempty"`

	// Dimensions: Requests which allow one of these (width, height) pairs
	// will match. All pairs must be supported ad dimensions.
	Dimensions []*PretargetingConfigDimensions `json:"dimensions,omitempty"`

	// ExcludedContentLabels: Requests with any of these content labels will
	// not match. Values are from content-labels.txt in the downloadable
	// files section.
	ExcludedContentLabels googleapi.Int64s `json:"excludedContentLabels,omitempty"`

	// ExcludedGeoCriteriaIds: Requests containing any of these geo criteria
	// ids will not match.
	ExcludedGeoCriteriaIds googleapi.Int64s `json:"excludedGeoCriteriaIds,omitempty"`

	// ExcludedPlacements: Requests containing any of these placements will
	// not match.
	ExcludedPlacements []*PretargetingConfigExcludedPlacements `json:"excludedPlacements,omitempty"`

	// ExcludedUserLists: Requests containing any of these users list ids
	// will not match.
	ExcludedUserLists googleapi.Int64s `json:"excludedUserLists,omitempty"`

	// ExcludedVerticals: Requests containing any of these vertical ids will
	// not match. Values are from the publisher-verticals.txt file in the
	// downloadable files section.
	ExcludedVerticals googleapi.Int64s `json:"excludedVerticals,omitempty"`

	// GeoCriteriaIds: Requests containing any of these geo criteria ids
	// will match.
	GeoCriteriaIds googleapi.Int64s `json:"geoCriteriaIds,omitempty"`

	// IsActive: Whether this config is active. Required for all requests.
	IsActive bool `json:"isActive,omitempty"`

	// Kind: The kind of the resource, i.e.
	// "adexchangebuyer#pretargetingConfig".
	Kind string `json:"kind,omitempty"`

	// Languages: Request containing any of these language codes will match.
	Languages []string `json:"languages,omitempty"`

	// MobileCarriers: Requests containing any of these mobile carrier ids
	// will match. Values are from mobile-carriers.csv in the downloadable
	// files section.
	MobileCarriers googleapi.Int64s `json:"mobileCarriers,omitempty"`

	// MobileDevices: Requests containing any of these mobile device ids
	// will match. Values are from mobile-devices.csv in the downloadable
	// files section.
	MobileDevices googleapi.Int64s `json:"mobileDevices,omitempty"`

	// MobileOperatingSystemVersions: Requests containing any of these
	// mobile operating system version ids will match. Values are from
	// mobile-os.csv in the downloadable files section.
	MobileOperatingSystemVersions googleapi.Int64s `json:"mobileOperatingSystemVersions,omitempty"`

	// Placements: Requests containing any of these placements will match.
	Placements []*PretargetingConfigPlacements `json:"placements,omitempty"`

	// Platforms: Requests matching any of these platforms will match.
	// Possible values are PRETARGETING_PLATFORM_MOBILE,
	// PRETARGETING_PLATFORM_DESKTOP, and PRETARGETING_PLATFORM_TABLET.
	Platforms []string `json:"platforms,omitempty"`

	// SupportedCreativeAttributes: Creative attributes should be declared
	// here if all creatives corresponding to this pretargeting
	// configuration have that creative attribute. Values are from
	// pretargetable-creative-attributes.txt in the downloadable files
	// section.
	SupportedCreativeAttributes googleapi.Int64s `json:"supportedCreativeAttributes,omitempty"`

	// UserLists: Requests containing any of these user list ids will match.
	UserLists googleapi.Int64s `json:"userLists,omitempty"`

	// VendorTypes: Requests that allow any of these vendor ids will match.
	// Values are from vendors.txt in the downloadable files section.
	VendorTypes googleapi.Int64s `json:"vendorTypes,omitempty"`

	// Verticals: Requests containing any of these vertical ids will match.
	Verticals googleapi.Int64s `json:"verticals,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "BillingId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PretargetingConfig) MarshalJSON() ([]byte, error) {
	type noMethod PretargetingConfig
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PretargetingConfigDimensions struct {
	// Height: Height in pixels.
	Height int64 `json:"height,omitempty,string"`

	// Width: Width in pixels.
	Width int64 `json:"width,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PretargetingConfigDimensions) MarshalJSON() ([]byte, error) {
	type noMethod PretargetingConfigDimensions
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PretargetingConfigExcludedPlacements struct {
	// Token: The value of the placement. Interpretation depends on the
	// placement type, e.g. URL for a site placement, channel name for a
	// channel placement, app id for a mobile app placement.
	Token string `json:"token,omitempty"`

	// Type: The type of the placement.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Token") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PretargetingConfigExcludedPlacements) MarshalJSON() ([]byte, error) {
	type noMethod PretargetingConfigExcludedPlacements
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PretargetingConfigPlacements struct {
	// Token: The value of the placement. Interpretation depends on the
	// placement type, e.g. URL for a site placement, channel name for a
	// channel placement, app id for a mobile app placement.
	Token string `json:"token,omitempty"`

	// Type: The type of the placement.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Token") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PretargetingConfigPlacements) MarshalJSON() ([]byte, error) {
	type noMethod PretargetingConfigPlacements
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PretargetingConfigList struct {
	// Items: A list of pretargeting configs
	Items []*PretargetingConfig `json:"items,omitempty"`

	// Kind: Resource type.
	Kind string `json:"kind,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Items") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PretargetingConfigList) MarshalJSON() ([]byte, error) {
	type noMethod PretargetingConfigList
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Price struct {
	// AmountMicros: The CPM value in micros.
	AmountMicros float64 `json:"amountMicros,omitempty"`

	// CurrencyCode: The currency code for the price.
	CurrencyCode string `json:"currencyCode,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AmountMicros") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Price) MarshalJSON() ([]byte, error) {
	type noMethod Price
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PricePerBuyer: Used to specify pricing rules for buyers/advertisers.
// Each PricePerBuyer in an offer can become [0,1] deals. To check if
// there is a PricePerBuyer for a particular buyer or buyer/advertiser
// pair, we look for the most specific matching rule - we first look for
// a rule matching the buyer and advertiser, next a rule with the buyer
// but an empty advertiser list, and otherwise look for a matching rule
// where no buyer is set.
type PricePerBuyer struct {
	// Buyer: The buyer who will pay this price. If unset, all buyers can
	// pay this price (if the advertisers match, and there's no more
	// specific rule matching the buyer).
	Buyer *Buyer `json:"buyer,omitempty"`

	// Price: The specified price
	Price *Price `json:"price,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Buyer") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PricePerBuyer) MarshalJSON() ([]byte, error) {
	type noMethod PricePerBuyer
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PrivateData struct {
	ReferenceId string `json:"referenceId,omitempty"`

	ReferencePayload string `json:"referencePayload,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ReferenceId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PrivateData) MarshalJSON() ([]byte, error) {
	type noMethod PrivateData
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type RuleKeyValuePair struct {
	KeyName string `json:"keyName,omitempty"`

	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "KeyName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RuleKeyValuePair) MarshalJSON() ([]byte, error) {
	type noMethod RuleKeyValuePair
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Seller struct {
	// AccountId: The unique id for the seller. The seller fills in this
	// field. The seller account id is then available to buyer in the offer.
	AccountId string `json:"accountId,omitempty"`

	// SubAccountId: Optional sub-account id for the seller.
	SubAccountId string `json:"subAccountId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Seller) MarshalJSON() ([]byte, error) {
	type noMethod Seller
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type SharedTargeting struct {
	// Exclusions: The list of values to exclude from targeting.
	Exclusions []*TargetingValue `json:"exclusions,omitempty"`

	// Inclusions: The list of value to include as part of the targeting.
	Inclusions []*TargetingValue `json:"inclusions,omitempty"`

	// Key: The key representing the shared targeting criterion.
	Key string `json:"key,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Exclusions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SharedTargeting) MarshalJSON() ([]byte, error) {
	type noMethod SharedTargeting
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type StatsDto struct {
	Bids int64 `json:"bids,omitempty,string"`

	GoodBids int64 `json:"goodBids,omitempty,string"`

	Impressions int64 `json:"impressions,omitempty,string"`

	Requests int64 `json:"requests,omitempty,string"`

	Revenue *MoneyDto `json:"revenue,omitempty"`

	Spend *MoneyDto `json:"spend,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Bids") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *StatsDto) MarshalJSON() ([]byte, error) {
	type noMethod StatsDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TargetingValue struct {
	// CreativeSizeValue: The creative size value to exclude/include.
	CreativeSizeValue *TargetingValueCreativeSize `json:"creativeSizeValue,omitempty"`

	// DayPartTargetingValue: The daypart targeting to include / exclude.
	// Filled in when the key is GOOG_DAYPART_TARGETING.
	DayPartTargetingValue *TargetingValueDayPartTargeting `json:"dayPartTargetingValue,omitempty"`

	// LongValue: The long value to exclude/include.
	LongValue int64 `json:"longValue,omitempty,string"`

	// StringValue: The string value to exclude/include.
	StringValue string `json:"stringValue,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CreativeSizeValue")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TargetingValue) MarshalJSON() ([]byte, error) {
	type noMethod TargetingValue
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TargetingValueCreativeSize struct {
	// CompanionSizes: For video size type, the list of companion sizes.
	CompanionSizes []*TargetingValueSize `json:"companionSizes,omitempty"`

	// CreativeSizeType: The Creative size type.
	CreativeSizeType string `json:"creativeSizeType,omitempty"`

	// Size: For regular creative size type, specifies the size of the
	// creative.
	Size *TargetingValueSize `json:"size,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CompanionSizes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TargetingValueCreativeSize) MarshalJSON() ([]byte, error) {
	type noMethod TargetingValueCreativeSize
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TargetingValueDayPartTargeting struct {
	DayParts []*TargetingValueDayPartTargetingDayPart `json:"dayParts,omitempty"`

	TimeZoneType string `json:"timeZoneType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DayParts") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TargetingValueDayPartTargeting) MarshalJSON() ([]byte, error) {
	type noMethod TargetingValueDayPartTargeting
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TargetingValueDayPartTargetingDayPart struct {
	DayOfWeek string `json:"dayOfWeek,omitempty"`

	EndHour int64 `json:"endHour,omitempty"`

	EndMinute int64 `json:"endMinute,omitempty"`

	StartHour int64 `json:"startHour,omitempty"`

	StartMinute int64 `json:"startMinute,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DayOfWeek") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TargetingValueDayPartTargetingDayPart) MarshalJSON() ([]byte, error) {
	type noMethod TargetingValueDayPartTargetingDayPart
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TargetingValueSize struct {
	// Height: The height of the creative.
	Height int64 `json:"height,omitempty"`

	// Width: The width of the creative.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TargetingValueSize) MarshalJSON() ([]byte, error) {
	type noMethod TargetingValueSize
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type TermsDto struct {
	// AdSlots: The particular ad slots targeted by the offer.
	AdSlots []*AdSlotDto `json:"adSlots,omitempty"`

	// Advertisers: A list of advertisers for this offer.
	Advertisers []*AdvertiserDto `json:"advertisers,omitempty"`

	// AudienceSegment: The audience segment for the offer.
	AudienceSegment *AudienceSegment `json:"audienceSegment,omitempty"`

	// AudienceSegmentDescription: A description of the audience segment for
	// the offer.
	AudienceSegmentDescription string `json:"audienceSegmentDescription,omitempty"`

	// BillingTerms: The billing terms.
	BillingTerms string `json:"billingTerms,omitempty"`

	// BuyerBillingType: The buyer billing type.
	BuyerBillingType string `json:"buyerBillingType,omitempty"`

	// Cpm: The cpm terms.
	Cpm *MoneyDto `json:"cpm,omitempty"`

	// CreativeBlockingLevel: Whether to use or ignore publisher blocking
	// rules.
	CreativeBlockingLevel string `json:"creativeBlockingLevel,omitempty"`

	// CreativeReviewPolicy: Whether to use publisher review policy or AdX
	// review policy.
	CreativeReviewPolicy string `json:"creativeReviewPolicy,omitempty"`

	// DealPremium: The premium terms.
	DealPremium *MoneyDto `json:"dealPremium,omitempty"`

	// Description: A description for these terms.
	Description string `json:"description,omitempty"`

	// DescriptiveName: A descriptive name for these terms.
	DescriptiveName string `json:"descriptiveName,omitempty"`

	// EndDate: The end date for the offer.
	EndDate *DateTime `json:"endDate,omitempty"`

	// EstimatedImpressionsPerDay: The estimated daily impressions for the
	// offer.
	EstimatedImpressionsPerDay int64 `json:"estimatedImpressionsPerDay,omitempty,string"`

	// EstimatedSpend: The estimated spend for the offer.
	EstimatedSpend *MoneyDto `json:"estimatedSpend,omitempty"`

	// FinalizeAutomatically: If true, the offer will finalize automatically
	// when accepted.
	FinalizeAutomatically bool `json:"finalizeAutomatically,omitempty"`

	// InventorySegmentTargeting: The inventory segment targeting for the
	// offer.
	InventorySegmentTargeting *InventorySegmentTargeting `json:"inventorySegmentTargeting,omitempty"`

	// IsReservation: Whether the offer is a reservation.
	IsReservation bool `json:"isReservation,omitempty"`

	// MinimumSpendMicros: The minimum spend for the offer.
	MinimumSpendMicros int64 `json:"minimumSpendMicros,omitempty,string"`

	// MinimumTrueLooks: The minimum true looks for the offer.
	MinimumTrueLooks int64 `json:"minimumTrueLooks,omitempty,string"`

	// MonetizerType: The monetizer type.
	MonetizerType string `json:"monetizerType,omitempty"`

	// SemiTransparent: Whether this offer is semi-transparent.
	SemiTransparent bool `json:"semiTransparent,omitempty"`

	// StartDate: The start date for the offer.
	StartDate *DateTime `json:"startDate,omitempty"`

	// TargetByDealId: Whether to target by deal id.
	TargetByDealId bool `json:"targetByDealId,omitempty"`

	// TargetingAllAdSlots: If true, the offer targets all ad slots.
	TargetingAllAdSlots bool `json:"targetingAllAdSlots,omitempty"`

	// TermsAttributes: A list of terms attributes.
	TermsAttributes []string `json:"termsAttributes,omitempty"`

	// Urls: The urls applicable to the offer.
	Urls []string `json:"urls,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AdSlots") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TermsDto) MarshalJSON() ([]byte, error) {
	type noMethod TermsDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type WebPropertyDto struct {
	AllowInterestTargetedAds bool `json:"allowInterestTargetedAds,omitempty"`

	EnabledForPreferredDeals bool `json:"enabledForPreferredDeals,omitempty"`

	Id int64 `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	PropertyCode string `json:"propertyCode,omitempty"`

	SiteUrls []string `json:"siteUrls,omitempty"`

	SyndicationProduct string `json:"syndicationProduct,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "AllowInterestTargetedAds") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *WebPropertyDto) MarshalJSON() ([]byte, error) {
	type noMethod WebPropertyDto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// method id "adexchangebuyer.accounts.get":

type AccountsGetCall struct {
	s    *Service
	id   int64
	opt_ map[string]interface{}
	ctx_ context.Context
}

// Get: Gets one account by ID.
func (r *AccountsService) Get(id int64) *AccountsGetCall {
	c := &AccountsGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.id = id
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsGetCall) Fields(s ...googleapi.Field) *AccountsGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsGetCall) IfNoneMatch(entityTag string) *AccountsGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *AccountsGetCall) Context(ctx context.Context) *AccountsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "accounts/{id}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"id": strconv.FormatInt(c.id, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.accounts.get" call.
// Exactly one of *Account or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Account.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsGetCall) Do() (*Account, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Account{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets one account by ID.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.accounts.get",
	//   "parameterOrder": [
	//     "id"
	//   ],
	//   "parameters": {
	//     "id": {
	//       "description": "The account id",
	//       "format": "int32",
	//       "location": "path",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "accounts/{id}",
	//   "response": {
	//     "$ref": "Account"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.accounts.list":

type AccountsListCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// List: Retrieves the authenticated user's list of accounts.
func (r *AccountsService) List() *AccountsListCall {
	c := &AccountsListCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsListCall) Fields(s ...googleapi.Field) *AccountsListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsListCall) IfNoneMatch(entityTag string) *AccountsListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *AccountsListCall) Context(ctx context.Context) *AccountsListCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "accounts")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.accounts.list" call.
// Exactly one of *AccountsList or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *AccountsList.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *AccountsListCall) Do() (*AccountsList, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &AccountsList{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves the authenticated user's list of accounts.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.accounts.list",
	//   "path": "accounts",
	//   "response": {
	//     "$ref": "AccountsList"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.accounts.patch":

type AccountsPatchCall struct {
	s       *Service
	id      int64
	account *Account
	opt_    map[string]interface{}
	ctx_    context.Context
}

// Patch: Updates an existing account. This method supports patch
// semantics.
func (r *AccountsService) Patch(id int64, account *Account) *AccountsPatchCall {
	c := &AccountsPatchCall{s: r.s, opt_: make(map[string]interface{})}
	c.id = id
	c.account = account
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsPatchCall) Fields(s ...googleapi.Field) *AccountsPatchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *AccountsPatchCall) Context(ctx context.Context) *AccountsPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.account)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "accounts/{id}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"id": strconv.FormatInt(c.id, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.accounts.patch" call.
// Exactly one of *Account or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Account.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsPatchCall) Do() (*Account, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Account{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing account. This method supports patch semantics.",
	//   "httpMethod": "PATCH",
	//   "id": "adexchangebuyer.accounts.patch",
	//   "parameterOrder": [
	//     "id"
	//   ],
	//   "parameters": {
	//     "id": {
	//       "description": "The account id",
	//       "format": "int32",
	//       "location": "path",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "accounts/{id}",
	//   "request": {
	//     "$ref": "Account"
	//   },
	//   "response": {
	//     "$ref": "Account"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.accounts.update":

type AccountsUpdateCall struct {
	s       *Service
	id      int64
	account *Account
	opt_    map[string]interface{}
	ctx_    context.Context
}

// Update: Updates an existing account.
func (r *AccountsService) Update(id int64, account *Account) *AccountsUpdateCall {
	c := &AccountsUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.id = id
	c.account = account
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsUpdateCall) Fields(s ...googleapi.Field) *AccountsUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *AccountsUpdateCall) Context(ctx context.Context) *AccountsUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.account)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "accounts/{id}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"id": strconv.FormatInt(c.id, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.accounts.update" call.
// Exactly one of *Account or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Account.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsUpdateCall) Do() (*Account, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Account{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing account.",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer.accounts.update",
	//   "parameterOrder": [
	//     "id"
	//   ],
	//   "parameters": {
	//     "id": {
	//       "description": "The account id",
	//       "format": "int32",
	//       "location": "path",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "accounts/{id}",
	//   "request": {
	//     "$ref": "Account"
	//   },
	//   "response": {
	//     "$ref": "Account"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.billingInfo.get":

type BillingInfoGetCall struct {
	s         *Service
	accountId int64
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Get: Returns the billing information for one account specified by
// account ID.
func (r *BillingInfoService) Get(accountId int64) *BillingInfoGetCall {
	c := &BillingInfoGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *BillingInfoGetCall) Fields(s ...googleapi.Field) *BillingInfoGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *BillingInfoGetCall) IfNoneMatch(entityTag string) *BillingInfoGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *BillingInfoGetCall) Context(ctx context.Context) *BillingInfoGetCall {
	c.ctx_ = ctx
	return c
}

func (c *BillingInfoGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "billinginfo/{accountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.billingInfo.get" call.
// Exactly one of *BillingInfo or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *BillingInfo.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *BillingInfoGetCall) Do() (*BillingInfo, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &BillingInfo{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns the billing information for one account specified by account ID.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.billingInfo.get",
	//   "parameterOrder": [
	//     "accountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id.",
	//       "format": "int32",
	//       "location": "path",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "billinginfo/{accountId}",
	//   "response": {
	//     "$ref": "BillingInfo"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.billingInfo.list":

type BillingInfoListCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// List: Retrieves a list of billing information for all accounts of the
// authenticated user.
func (r *BillingInfoService) List() *BillingInfoListCall {
	c := &BillingInfoListCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *BillingInfoListCall) Fields(s ...googleapi.Field) *BillingInfoListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *BillingInfoListCall) IfNoneMatch(entityTag string) *BillingInfoListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *BillingInfoListCall) Context(ctx context.Context) *BillingInfoListCall {
	c.ctx_ = ctx
	return c
}

func (c *BillingInfoListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "billinginfo")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.billingInfo.list" call.
// Exactly one of *BillingInfoList or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *BillingInfoList.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *BillingInfoListCall) Do() (*BillingInfoList, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &BillingInfoList{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves a list of billing information for all accounts of the authenticated user.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.billingInfo.list",
	//   "path": "billinginfo",
	//   "response": {
	//     "$ref": "BillingInfoList"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.budget.get":

type BudgetGetCall struct {
	s         *Service
	accountId int64
	billingId int64
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Get: Returns the budget information for the adgroup specified by the
// accountId and billingId.
func (r *BudgetService) Get(accountId int64, billingId int64) *BudgetGetCall {
	c := &BudgetGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.billingId = billingId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *BudgetGetCall) Fields(s ...googleapi.Field) *BudgetGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *BudgetGetCall) IfNoneMatch(entityTag string) *BudgetGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *BudgetGetCall) Context(ctx context.Context) *BudgetGetCall {
	c.ctx_ = ctx
	return c
}

func (c *BudgetGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "billinginfo/{accountId}/{billingId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"billingId": strconv.FormatInt(c.billingId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.budget.get" call.
// Exactly one of *Budget or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Budget.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *BudgetGetCall) Do() (*Budget, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Budget{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns the budget information for the adgroup specified by the accountId and billingId.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.budget.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "billingId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to get the budget information for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "billingId": {
	//       "description": "The billing id to get the budget information for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "billinginfo/{accountId}/{billingId}",
	//   "response": {
	//     "$ref": "Budget"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.budget.patch":

type BudgetPatchCall struct {
	s         *Service
	accountId int64
	billingId int64
	budget    *Budget
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Patch: Updates the budget amount for the budget of the adgroup
// specified by the accountId and billingId, with the budget amount in
// the request. This method supports patch semantics.
func (r *BudgetService) Patch(accountId int64, billingId int64, budget *Budget) *BudgetPatchCall {
	c := &BudgetPatchCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.billingId = billingId
	c.budget = budget
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *BudgetPatchCall) Fields(s ...googleapi.Field) *BudgetPatchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *BudgetPatchCall) Context(ctx context.Context) *BudgetPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *BudgetPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.budget)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "billinginfo/{accountId}/{billingId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"billingId": strconv.FormatInt(c.billingId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.budget.patch" call.
// Exactly one of *Budget or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Budget.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *BudgetPatchCall) Do() (*Budget, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Budget{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates the budget amount for the budget of the adgroup specified by the accountId and billingId, with the budget amount in the request. This method supports patch semantics.",
	//   "httpMethod": "PATCH",
	//   "id": "adexchangebuyer.budget.patch",
	//   "parameterOrder": [
	//     "accountId",
	//     "billingId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id associated with the budget being updated.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "billingId": {
	//       "description": "The billing id associated with the budget being updated.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "billinginfo/{accountId}/{billingId}",
	//   "request": {
	//     "$ref": "Budget"
	//   },
	//   "response": {
	//     "$ref": "Budget"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.budget.update":

type BudgetUpdateCall struct {
	s         *Service
	accountId int64
	billingId int64
	budget    *Budget
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Update: Updates the budget amount for the budget of the adgroup
// specified by the accountId and billingId, with the budget amount in
// the request.
func (r *BudgetService) Update(accountId int64, billingId int64, budget *Budget) *BudgetUpdateCall {
	c := &BudgetUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.billingId = billingId
	c.budget = budget
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *BudgetUpdateCall) Fields(s ...googleapi.Field) *BudgetUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *BudgetUpdateCall) Context(ctx context.Context) *BudgetUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *BudgetUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.budget)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "billinginfo/{accountId}/{billingId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"billingId": strconv.FormatInt(c.billingId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.budget.update" call.
// Exactly one of *Budget or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Budget.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *BudgetUpdateCall) Do() (*Budget, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Budget{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates the budget amount for the budget of the adgroup specified by the accountId and billingId, with the budget amount in the request.",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer.budget.update",
	//   "parameterOrder": [
	//     "accountId",
	//     "billingId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id associated with the budget being updated.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "billingId": {
	//       "description": "The billing id associated with the budget being updated.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "billinginfo/{accountId}/{billingId}",
	//   "request": {
	//     "$ref": "Budget"
	//   },
	//   "response": {
	//     "$ref": "Budget"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.delete":

type ClientaccessDeleteCall struct {
	s                *Service
	clientAccountId  int64
	sponsorAccountId int64
	opt_             map[string]interface{}
	ctx_             context.Context
}

// Delete:
func (r *ClientaccessService) Delete(clientAccountId int64, sponsorAccountId int64) *ClientaccessDeleteCall {
	c := &ClientaccessDeleteCall{s: r.s, opt_: make(map[string]interface{})}
	c.clientAccountId = clientAccountId
	c.sponsorAccountId = sponsorAccountId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessDeleteCall) Fields(s ...googleapi.Field) *ClientaccessDeleteCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessDeleteCall) Context(ctx context.Context) *ClientaccessDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("sponsorAccountId", fmt.Sprintf("%v", c.sponsorAccountId))
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess/{clientAccountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.delete" call.
func (c *ClientaccessDeleteCall) Do() error {
	res, err := c.doRequest("json")
	if err != nil {
		return err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return err
	}
	return nil
	// {
	//   "httpMethod": "DELETE",
	//   "id": "adexchangebuyer.clientaccess.delete",
	//   "parameterOrder": [
	//     "clientAccountId",
	//     "sponsorAccountId"
	//   ],
	//   "parameters": {
	//     "clientAccountId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sponsorAccountId": {
	//       "format": "int32",
	//       "location": "query",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "clientAccess/{clientAccountId}",
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.get":

type ClientaccessGetCall struct {
	s                *Service
	clientAccountId  int64
	sponsorAccountId int64
	opt_             map[string]interface{}
	ctx_             context.Context
}

// Get:
func (r *ClientaccessService) Get(clientAccountId int64, sponsorAccountId int64) *ClientaccessGetCall {
	c := &ClientaccessGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.clientAccountId = clientAccountId
	c.sponsorAccountId = sponsorAccountId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessGetCall) Fields(s ...googleapi.Field) *ClientaccessGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ClientaccessGetCall) IfNoneMatch(entityTag string) *ClientaccessGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessGetCall) Context(ctx context.Context) *ClientaccessGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("sponsorAccountId", fmt.Sprintf("%v", c.sponsorAccountId))
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess/{clientAccountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.get" call.
// Exactly one of *ClientAccessCapabilities or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ClientAccessCapabilities.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ClientaccessGetCall) Do() (*ClientAccessCapabilities, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientAccessCapabilities{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.clientaccess.get",
	//   "parameterOrder": [
	//     "clientAccountId",
	//     "sponsorAccountId"
	//   ],
	//   "parameters": {
	//     "clientAccountId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sponsorAccountId": {
	//       "format": "int32",
	//       "location": "query",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "clientAccess/{clientAccountId}",
	//   "response": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.insert":

type ClientaccessInsertCall struct {
	s                        *Service
	clientaccesscapabilities *ClientAccessCapabilities
	opt_                     map[string]interface{}
	ctx_                     context.Context
}

// Insert:
func (r *ClientaccessService) Insert(clientaccesscapabilities *ClientAccessCapabilities) *ClientaccessInsertCall {
	c := &ClientaccessInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.clientaccesscapabilities = clientaccesscapabilities
	return c
}

// ClientAccountId sets the optional parameter "clientAccountId":
func (c *ClientaccessInsertCall) ClientAccountId(clientAccountId int64) *ClientaccessInsertCall {
	c.opt_["clientAccountId"] = clientAccountId
	return c
}

// SponsorAccountId sets the optional parameter "sponsorAccountId":
func (c *ClientaccessInsertCall) SponsorAccountId(sponsorAccountId int64) *ClientaccessInsertCall {
	c.opt_["sponsorAccountId"] = sponsorAccountId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessInsertCall) Fields(s ...googleapi.Field) *ClientaccessInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessInsertCall) Context(ctx context.Context) *ClientaccessInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.clientaccesscapabilities)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["clientAccountId"]; ok {
		params.Set("clientAccountId", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["sponsorAccountId"]; ok {
		params.Set("sponsorAccountId", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.insert" call.
// Exactly one of *ClientAccessCapabilities or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ClientAccessCapabilities.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ClientaccessInsertCall) Do() (*ClientAccessCapabilities, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientAccessCapabilities{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.clientaccess.insert",
	//   "parameters": {
	//     "clientAccountId": {
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sponsorAccountId": {
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     }
	//   },
	//   "path": "clientAccess",
	//   "request": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "response": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.list":

type ClientaccessListCall struct {
	s                                   *Service
	listclientaccesscapabilitiesrequest *ListClientAccessCapabilitiesRequest
	opt_                                map[string]interface{}
	ctx_                                context.Context
}

// List:
func (r *ClientaccessService) List(listclientaccesscapabilitiesrequest *ListClientAccessCapabilitiesRequest) *ClientaccessListCall {
	c := &ClientaccessListCall{s: r.s, opt_: make(map[string]interface{})}
	c.listclientaccesscapabilitiesrequest = listclientaccesscapabilitiesrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessListCall) Fields(s ...googleapi.Field) *ClientaccessListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ClientaccessListCall) IfNoneMatch(entityTag string) *ClientaccessListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessListCall) Context(ctx context.Context) *ClientaccessListCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.list" call.
// Exactly one of *ListClientAccessCapabilitiesResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *ListClientAccessCapabilitiesResponse.ServerResponse.Header or
// (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ClientaccessListCall) Do() (*ListClientAccessCapabilitiesResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListClientAccessCapabilitiesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.clientaccess.list",
	//   "path": "clientAccess",
	//   "request": {
	//     "$ref": "ListClientAccessCapabilitiesRequest"
	//   },
	//   "response": {
	//     "$ref": "ListClientAccessCapabilitiesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.patch":

type ClientaccessPatchCall struct {
	s                        *Service
	clientAccountId          int64
	sponsorAccountId         int64
	clientaccesscapabilities *ClientAccessCapabilities
	opt_                     map[string]interface{}
	ctx_                     context.Context
}

// Patch:
func (r *ClientaccessService) Patch(clientAccountId int64, sponsorAccountId int64, clientaccesscapabilities *ClientAccessCapabilities) *ClientaccessPatchCall {
	c := &ClientaccessPatchCall{s: r.s, opt_: make(map[string]interface{})}
	c.clientAccountId = clientAccountId
	c.sponsorAccountId = sponsorAccountId
	c.clientaccesscapabilities = clientaccesscapabilities
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessPatchCall) Fields(s ...googleapi.Field) *ClientaccessPatchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessPatchCall) Context(ctx context.Context) *ClientaccessPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.clientaccesscapabilities)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("sponsorAccountId", fmt.Sprintf("%v", c.sponsorAccountId))
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess/{clientAccountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.patch" call.
// Exactly one of *ClientAccessCapabilities or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ClientAccessCapabilities.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ClientaccessPatchCall) Do() (*ClientAccessCapabilities, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientAccessCapabilities{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "PATCH",
	//   "id": "adexchangebuyer.clientaccess.patch",
	//   "parameterOrder": [
	//     "clientAccountId",
	//     "sponsorAccountId"
	//   ],
	//   "parameters": {
	//     "clientAccountId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sponsorAccountId": {
	//       "format": "int32",
	//       "location": "query",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "clientAccess/{clientAccountId}",
	//   "request": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "response": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.clientaccess.update":

type ClientaccessUpdateCall struct {
	s                        *Service
	clientAccountId          int64
	sponsorAccountId         int64
	clientaccesscapabilities *ClientAccessCapabilities
	opt_                     map[string]interface{}
	ctx_                     context.Context
}

// Update:
func (r *ClientaccessService) Update(clientAccountId int64, sponsorAccountId int64, clientaccesscapabilities *ClientAccessCapabilities) *ClientaccessUpdateCall {
	c := &ClientaccessUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.clientAccountId = clientAccountId
	c.sponsorAccountId = sponsorAccountId
	c.clientaccesscapabilities = clientaccesscapabilities
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ClientaccessUpdateCall) Fields(s ...googleapi.Field) *ClientaccessUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ClientaccessUpdateCall) Context(ctx context.Context) *ClientaccessUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *ClientaccessUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.clientaccesscapabilities)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("sponsorAccountId", fmt.Sprintf("%v", c.sponsorAccountId))
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "clientAccess/{clientAccountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.clientaccess.update" call.
// Exactly one of *ClientAccessCapabilities or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ClientAccessCapabilities.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ClientaccessUpdateCall) Do() (*ClientAccessCapabilities, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientAccessCapabilities{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer.clientaccess.update",
	//   "parameterOrder": [
	//     "clientAccountId",
	//     "sponsorAccountId"
	//   ],
	//   "parameters": {
	//     "clientAccountId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sponsorAccountId": {
	//       "format": "int32",
	//       "location": "query",
	//       "required": true,
	//       "type": "integer"
	//     }
	//   },
	//   "path": "clientAccess/{clientAccountId}",
	//   "request": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "response": {
	//     "$ref": "ClientAccessCapabilities"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.creatives.get":

type CreativesGetCall struct {
	s               *Service
	accountId       int64
	buyerCreativeId string
	opt_            map[string]interface{}
	ctx_            context.Context
}

// Get: Gets the status for a single creative. A creative will be
// available 30-40 minutes after submission.
func (r *CreativesService) Get(accountId int64, buyerCreativeId string) *CreativesGetCall {
	c := &CreativesGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.buyerCreativeId = buyerCreativeId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CreativesGetCall) Fields(s ...googleapi.Field) *CreativesGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *CreativesGetCall) IfNoneMatch(entityTag string) *CreativesGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *CreativesGetCall) Context(ctx context.Context) *CreativesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *CreativesGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "creatives/{accountId}/{buyerCreativeId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"buyerCreativeId": c.buyerCreativeId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.creatives.get" call.
// Exactly one of *Creative or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Creative.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *CreativesGetCall) Do() (*Creative, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Creative{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the status for a single creative. A creative will be available 30-40 minutes after submission.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.creatives.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "buyerCreativeId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The id for the account that will serve this creative.",
	//       "format": "int32",
	//       "location": "path",
	//       "required": true,
	//       "type": "integer"
	//     },
	//     "buyerCreativeId": {
	//       "description": "The buyer-specific id for this creative.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "creatives/{accountId}/{buyerCreativeId}",
	//   "response": {
	//     "$ref": "Creative"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.creatives.insert":

type CreativesInsertCall struct {
	s        *Service
	creative *Creative
	opt_     map[string]interface{}
	ctx_     context.Context
}

// Insert: Submit a new creative.
func (r *CreativesService) Insert(creative *Creative) *CreativesInsertCall {
	c := &CreativesInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.creative = creative
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CreativesInsertCall) Fields(s ...googleapi.Field) *CreativesInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *CreativesInsertCall) Context(ctx context.Context) *CreativesInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *CreativesInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.creative)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "creatives")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.creatives.insert" call.
// Exactly one of *Creative or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Creative.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *CreativesInsertCall) Do() (*Creative, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Creative{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Submit a new creative.",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.creatives.insert",
	//   "path": "creatives",
	//   "request": {
	//     "$ref": "Creative"
	//   },
	//   "response": {
	//     "$ref": "Creative"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.creatives.list":

type CreativesListCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// List: Retrieves a list of the authenticated user's active creatives.
// A creative will be available 30-40 minutes after submission.
func (r *CreativesService) List() *CreativesListCall {
	c := &CreativesListCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// AccountId sets the optional parameter "accountId": When specified,
// only creatives for the given account ids are returned.
func (c *CreativesListCall) AccountId(accountId int64) *CreativesListCall {
	c.opt_["accountId"] = accountId
	return c
}

// BuyerCreativeId sets the optional parameter "buyerCreativeId": When
// specified, only creatives for the given buyer creative ids are
// returned.
func (c *CreativesListCall) BuyerCreativeId(buyerCreativeId string) *CreativesListCall {
	c.opt_["buyerCreativeId"] = buyerCreativeId
	return c
}

// DealsStatusFilter sets the optional parameter "dealsStatusFilter":
// When specified, only creatives having the given direct deals status
// are returned.
//
// Possible values:
//   "approved" - Creatives which have been approved for serving on
// direct deals.
//   "conditionally_approved" - Creatives which have been conditionally
// approved for serving on direct deals.
//   "disapproved" - Creatives which have been disapproved for serving
// on direct deals.
//   "not_checked" - Creatives whose direct deals status is not yet
// checked.
func (c *CreativesListCall) DealsStatusFilter(dealsStatusFilter string) *CreativesListCall {
	c.opt_["dealsStatusFilter"] = dealsStatusFilter
	return c
}

// MaxResults sets the optional parameter "maxResults": Maximum number
// of entries returned on one result page. If not set, the default is
// 100.
func (c *CreativesListCall) MaxResults(maxResults int64) *CreativesListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// OpenAuctionStatusFilter sets the optional parameter
// "openAuctionStatusFilter": When specified, only creatives having the
// given open auction status are returned.
//
// Possible values:
//   "approved" - Creatives which have been approved for serving on the
// open auction.
//   "conditionally_approved" - Creatives which have been conditionally
// approved for serving on the open auction.
//   "disapproved" - Creatives which have been disapproved for serving
// on the open auction.
//   "not_checked" - Creatives whose open auction status is not yet
// checked.
func (c *CreativesListCall) OpenAuctionStatusFilter(openAuctionStatusFilter string) *CreativesListCall {
	c.opt_["openAuctionStatusFilter"] = openAuctionStatusFilter
	return c
}

// PageToken sets the optional parameter "pageToken": A continuation
// token, used to page through ad clients. To retrieve the next page,
// set this parameter to the value of "nextPageToken" from the previous
// response.
func (c *CreativesListCall) PageToken(pageToken string) *CreativesListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CreativesListCall) Fields(s ...googleapi.Field) *CreativesListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *CreativesListCall) IfNoneMatch(entityTag string) *CreativesListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *CreativesListCall) Context(ctx context.Context) *CreativesListCall {
	c.ctx_ = ctx
	return c
}

func (c *CreativesListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["accountId"]; ok {
		params.Set("accountId", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["buyerCreativeId"]; ok {
		params.Set("buyerCreativeId", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["dealsStatusFilter"]; ok {
		params.Set("dealsStatusFilter", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["openAuctionStatusFilter"]; ok {
		params.Set("openAuctionStatusFilter", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "creatives")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.creatives.list" call.
// Exactly one of *CreativesList or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *CreativesList.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *CreativesListCall) Do() (*CreativesList, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &CreativesList{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves a list of the authenticated user's active creatives. A creative will be available 30-40 minutes after submission.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.creatives.list",
	//   "parameters": {
	//     "accountId": {
	//       "description": "When specified, only creatives for the given account ids are returned.",
	//       "format": "int32",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "integer"
	//     },
	//     "buyerCreativeId": {
	//       "description": "When specified, only creatives for the given buyer creative ids are returned.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "dealsStatusFilter": {
	//       "description": "When specified, only creatives having the given direct deals status are returned.",
	//       "enum": [
	//         "approved",
	//         "conditionally_approved",
	//         "disapproved",
	//         "not_checked"
	//       ],
	//       "enumDescriptions": [
	//         "Creatives which have been approved for serving on direct deals.",
	//         "Creatives which have been conditionally approved for serving on direct deals.",
	//         "Creatives which have been disapproved for serving on direct deals.",
	//         "Creatives whose direct deals status is not yet checked."
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "description": "Maximum number of entries returned on one result page. If not set, the default is 100. Optional.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "1000",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "openAuctionStatusFilter": {
	//       "description": "When specified, only creatives having the given open auction status are returned.",
	//       "enum": [
	//         "approved",
	//         "conditionally_approved",
	//         "disapproved",
	//         "not_checked"
	//       ],
	//       "enumDescriptions": [
	//         "Creatives which have been approved for serving on the open auction.",
	//         "Creatives which have been conditionally approved for serving on the open auction.",
	//         "Creatives which have been disapproved for serving on the open auction.",
	//         "Creatives whose open auction status is not yet checked."
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "A continuation token, used to page through ad clients. To retrieve the next page, set this parameter to the value of \"nextPageToken\" from the previous response. Optional.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "creatives",
	//   "response": {
	//     "$ref": "CreativesList"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.deals.get":

type DealsGetCall struct {
	s                                              *Service
	dealId                                         int64
	getfinalizednegotiationbyexternaldealidrequest *GetFinalizedNegotiationByExternalDealIdRequest
	opt_                                           map[string]interface{}
	ctx_                                           context.Context
}

// Get: Gets the requested deal.
func (r *DealsService) Get(dealId int64, getfinalizednegotiationbyexternaldealidrequest *GetFinalizedNegotiationByExternalDealIdRequest) *DealsGetCall {
	c := &DealsGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.dealId = dealId
	c.getfinalizednegotiationbyexternaldealidrequest = getfinalizednegotiationbyexternaldealidrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *DealsGetCall) Fields(s ...googleapi.Field) *DealsGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *DealsGetCall) IfNoneMatch(entityTag string) *DealsGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *DealsGetCall) Context(ctx context.Context) *DealsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *DealsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "deals/{dealId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"dealId": strconv.FormatInt(c.dealId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.deals.get" call.
// Exactly one of *NegotiationDto or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *NegotiationDto.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *DealsGetCall) Do() (*NegotiationDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &NegotiationDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the requested deal.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.deals.get",
	//   "parameterOrder": [
	//     "dealId"
	//   ],
	//   "parameters": {
	//     "dealId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "deals/{dealId}",
	//   "request": {
	//     "$ref": "GetFinalizedNegotiationByExternalDealIdRequest"
	//   },
	//   "response": {
	//     "$ref": "NegotiationDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacedeals.delete":

type MarketplacedealsDeleteCall struct {
	s                       *Service
	orderId                 string
	deleteorderdealsrequest *DeleteOrderDealsRequest
	opt_                    map[string]interface{}
	ctx_                    context.Context
}

// Delete: Delete the specified deals from the order
func (r *MarketplacedealsService) Delete(orderId string, deleteorderdealsrequest *DeleteOrderDealsRequest) *MarketplacedealsDeleteCall {
	c := &MarketplacedealsDeleteCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.deleteorderdealsrequest = deleteorderdealsrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacedealsDeleteCall) Fields(s ...googleapi.Field) *MarketplacedealsDeleteCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacedealsDeleteCall) Context(ctx context.Context) *MarketplacedealsDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacedealsDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.deleteorderdealsrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/deals/delete")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacedeals.delete" call.
// Exactly one of *DeleteOrderDealsResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *DeleteOrderDealsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacedealsDeleteCall) Do() (*DeleteOrderDealsResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &DeleteOrderDealsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Delete the specified deals from the order",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.marketplacedeals.delete",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The orderId to delete deals from.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/deals/delete",
	//   "request": {
	//     "$ref": "DeleteOrderDealsRequest"
	//   },
	//   "response": {
	//     "$ref": "DeleteOrderDealsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacedeals.insert":

type MarketplacedealsInsertCall struct {
	s                    *Service
	orderId              string
	addorderdealsrequest *AddOrderDealsRequest
	opt_                 map[string]interface{}
	ctx_                 context.Context
}

// Insert: Add new deals for the specified order
func (r *MarketplacedealsService) Insert(orderId string, addorderdealsrequest *AddOrderDealsRequest) *MarketplacedealsInsertCall {
	c := &MarketplacedealsInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.addorderdealsrequest = addorderdealsrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacedealsInsertCall) Fields(s ...googleapi.Field) *MarketplacedealsInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacedealsInsertCall) Context(ctx context.Context) *MarketplacedealsInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacedealsInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.addorderdealsrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/deals/insert")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacedeals.insert" call.
// Exactly one of *AddOrderDealsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *AddOrderDealsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacedealsInsertCall) Do() (*AddOrderDealsResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &AddOrderDealsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Add new deals for the specified order",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.marketplacedeals.insert",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "OrderId for which deals need to be added.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/deals/insert",
	//   "request": {
	//     "$ref": "AddOrderDealsRequest"
	//   },
	//   "response": {
	//     "$ref": "AddOrderDealsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacedeals.list":

type MarketplacedealsListCall struct {
	s       *Service
	orderId string
	opt_    map[string]interface{}
	ctx_    context.Context
}

// List: List all the deals for a given order
func (r *MarketplacedealsService) List(orderId string) *MarketplacedealsListCall {
	c := &MarketplacedealsListCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacedealsListCall) Fields(s ...googleapi.Field) *MarketplacedealsListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplacedealsListCall) IfNoneMatch(entityTag string) *MarketplacedealsListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacedealsListCall) Context(ctx context.Context) *MarketplacedealsListCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacedealsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/deals")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacedeals.list" call.
// Exactly one of *GetOrderDealsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetOrderDealsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacedealsListCall) Do() (*GetOrderDealsResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetOrderDealsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "List all the deals for a given order",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplacedeals.list",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The orderId to get deals for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/deals",
	//   "response": {
	//     "$ref": "GetOrderDealsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacedeals.update":

type MarketplacedealsUpdateCall struct {
	s                        *Service
	orderId                  string
	editallorderdealsrequest *EditAllOrderDealsRequest
	opt_                     map[string]interface{}
	ctx_                     context.Context
}

// Update: Replaces all the deals in the order with the passed in deals
func (r *MarketplacedealsService) Update(orderId string, editallorderdealsrequest *EditAllOrderDealsRequest) *MarketplacedealsUpdateCall {
	c := &MarketplacedealsUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.editallorderdealsrequest = editallorderdealsrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacedealsUpdateCall) Fields(s ...googleapi.Field) *MarketplacedealsUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacedealsUpdateCall) Context(ctx context.Context) *MarketplacedealsUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacedealsUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.editallorderdealsrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/deals/update")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacedeals.update" call.
// Exactly one of *EditAllOrderDealsResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *EditAllOrderDealsResponse.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacedealsUpdateCall) Do() (*EditAllOrderDealsResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &EditAllOrderDealsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Replaces all the deals in the order with the passed in deals",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.marketplacedeals.update",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The orderId to edit deals on.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/deals/update",
	//   "request": {
	//     "$ref": "EditAllOrderDealsRequest"
	//   },
	//   "response": {
	//     "$ref": "EditAllOrderDealsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacenotes.insert":

type MarketplacenotesInsertCall struct {
	s                    *Service
	orderId              string
	addordernotesrequest *AddOrderNotesRequest
	opt_                 map[string]interface{}
	ctx_                 context.Context
}

// Insert: Add notes to the order
func (r *MarketplacenotesService) Insert(orderId string, addordernotesrequest *AddOrderNotesRequest) *MarketplacenotesInsertCall {
	c := &MarketplacenotesInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.addordernotesrequest = addordernotesrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacenotesInsertCall) Fields(s ...googleapi.Field) *MarketplacenotesInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacenotesInsertCall) Context(ctx context.Context) *MarketplacenotesInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacenotesInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.addordernotesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/notes/insert")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacenotes.insert" call.
// Exactly one of *AddOrderNotesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *AddOrderNotesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacenotesInsertCall) Do() (*AddOrderNotesResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &AddOrderNotesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Add notes to the order",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.marketplacenotes.insert",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The orderId to add notes for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/notes/insert",
	//   "request": {
	//     "$ref": "AddOrderNotesRequest"
	//   },
	//   "response": {
	//     "$ref": "AddOrderNotesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplacenotes.list":

type MarketplacenotesListCall struct {
	s       *Service
	orderId string
	opt_    map[string]interface{}
	ctx_    context.Context
}

// List: Get all the notes associated with an order
func (r *MarketplacenotesService) List(orderId string) *MarketplacenotesListCall {
	c := &MarketplacenotesListCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplacenotesListCall) Fields(s ...googleapi.Field) *MarketplacenotesListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplacenotesListCall) IfNoneMatch(entityTag string) *MarketplacenotesListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplacenotesListCall) Context(ctx context.Context) *MarketplacenotesListCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplacenotesListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/notes")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplacenotes.list" call.
// Exactly one of *GetOrderNotesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetOrderNotesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplacenotesListCall) Do() (*GetOrderNotesResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetOrderNotesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Get all the notes associated with an order",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplacenotes.list",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The orderId to get notes for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/notes",
	//   "response": {
	//     "$ref": "GetOrderNotesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceoffers.get":

type MarketplaceoffersGetCall struct {
	s       *Service
	offerId string
	opt_    map[string]interface{}
	ctx_    context.Context
}

// Get: Gets the requested negotiation.
func (r *MarketplaceoffersService) Get(offerId string) *MarketplaceoffersGetCall {
	c := &MarketplaceoffersGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.offerId = offerId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceoffersGetCall) Fields(s ...googleapi.Field) *MarketplaceoffersGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplaceoffersGetCall) IfNoneMatch(entityTag string) *MarketplaceoffersGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceoffersGetCall) Context(ctx context.Context) *MarketplaceoffersGetCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceoffersGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOffers/{offerId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"offerId": c.offerId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceoffers.get" call.
// Exactly one of *MarketplaceOffer or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MarketplaceOffer.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceoffersGetCall) Do() (*MarketplaceOffer, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MarketplaceOffer{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the requested negotiation.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplaceoffers.get",
	//   "parameterOrder": [
	//     "offerId"
	//   ],
	//   "parameters": {
	//     "offerId": {
	//       "description": "The offerId for the offer to get the head revision for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOffers/{offerId}",
	//   "response": {
	//     "$ref": "MarketplaceOffer"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceoffers.search":

type MarketplaceoffersSearchCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// Search: Gets the requested negotiation.
func (r *MarketplaceoffersService) Search() *MarketplaceoffersSearchCall {
	c := &MarketplaceoffersSearchCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// PqlQuery sets the optional parameter "pqlQuery": The pql query used
// to query for offers.
func (c *MarketplaceoffersSearchCall) PqlQuery(pqlQuery string) *MarketplaceoffersSearchCall {
	c.opt_["pqlQuery"] = pqlQuery
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceoffersSearchCall) Fields(s ...googleapi.Field) *MarketplaceoffersSearchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplaceoffersSearchCall) IfNoneMatch(entityTag string) *MarketplaceoffersSearchCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceoffersSearchCall) Context(ctx context.Context) *MarketplaceoffersSearchCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceoffersSearchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["pqlQuery"]; ok {
		params.Set("pqlQuery", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOffers/search")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceoffers.search" call.
// Exactly one of *GetOffersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetOffersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceoffersSearchCall) Do() (*GetOffersResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetOffersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the requested negotiation.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplaceoffers.search",
	//   "parameters": {
	//     "pqlQuery": {
	//       "description": "The pql query used to query for offers.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOffers/search",
	//   "response": {
	//     "$ref": "GetOffersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceorders.get":

type MarketplaceordersGetCall struct {
	s       *Service
	orderId string
	opt_    map[string]interface{}
	ctx_    context.Context
}

// Get: Get an order given its id
func (r *MarketplaceordersService) Get(orderId string) *MarketplaceordersGetCall {
	c := &MarketplaceordersGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceordersGetCall) Fields(s ...googleapi.Field) *MarketplaceordersGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplaceordersGetCall) IfNoneMatch(entityTag string) *MarketplaceordersGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceordersGetCall) Context(ctx context.Context) *MarketplaceordersGetCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceordersGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId": c.orderId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceorders.get" call.
// Exactly one of *MarketplaceOrder or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MarketplaceOrder.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceordersGetCall) Do() (*MarketplaceOrder, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MarketplaceOrder{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Get an order given its id",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplaceorders.get",
	//   "parameterOrder": [
	//     "orderId"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "Id of the order to retrieve.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}",
	//   "response": {
	//     "$ref": "MarketplaceOrder"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceorders.insert":

type MarketplaceordersInsertCall struct {
	s                   *Service
	createordersrequest *CreateOrdersRequest
	opt_                map[string]interface{}
	ctx_                context.Context
}

// Insert: Create the given list of orders
func (r *MarketplaceordersService) Insert(createordersrequest *CreateOrdersRequest) *MarketplaceordersInsertCall {
	c := &MarketplaceordersInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.createordersrequest = createordersrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceordersInsertCall) Fields(s ...googleapi.Field) *MarketplaceordersInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceordersInsertCall) Context(ctx context.Context) *MarketplaceordersInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceordersInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createordersrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/insert")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceorders.insert" call.
// Exactly one of *CreateOrdersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *CreateOrdersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceordersInsertCall) Do() (*CreateOrdersResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &CreateOrdersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Create the given list of orders",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.marketplaceorders.insert",
	//   "path": "marketplaceOrders/insert",
	//   "request": {
	//     "$ref": "CreateOrdersRequest"
	//   },
	//   "response": {
	//     "$ref": "CreateOrdersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceorders.patch":

type MarketplaceordersPatchCall struct {
	s                *Service
	orderId          string
	revisionNumber   int64
	updateAction     string
	marketplaceorder *MarketplaceOrder
	opt_             map[string]interface{}
	ctx_             context.Context
}

// Patch: Update the given order. This method supports patch semantics.
func (r *MarketplaceordersService) Patch(orderId string, revisionNumber int64, updateAction string, marketplaceorder *MarketplaceOrder) *MarketplaceordersPatchCall {
	c := &MarketplaceordersPatchCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.revisionNumber = revisionNumber
	c.updateAction = updateAction
	c.marketplaceorder = marketplaceorder
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceordersPatchCall) Fields(s ...googleapi.Field) *MarketplaceordersPatchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceordersPatchCall) Context(ctx context.Context) *MarketplaceordersPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceordersPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.marketplaceorder)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/{revisionNumber}/{updateAction}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId":        c.orderId,
		"revisionNumber": strconv.FormatInt(c.revisionNumber, 10),
		"updateAction":   c.updateAction,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceorders.patch" call.
// Exactly one of *MarketplaceOrder or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MarketplaceOrder.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceordersPatchCall) Do() (*MarketplaceOrder, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MarketplaceOrder{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Update the given order. This method supports patch semantics.",
	//   "httpMethod": "PATCH",
	//   "id": "adexchangebuyer.marketplaceorders.patch",
	//   "parameterOrder": [
	//     "orderId",
	//     "revisionNumber",
	//     "updateAction"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The order id to update.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionNumber": {
	//       "description": "The last known revision number to update. If the head revision in the marketplace database has since changed, an error will be thrown. The caller should then fetch the lastest order at head revision and retry the update at that revision.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "updateAction": {
	//       "description": "The proposed action to take on the order.",
	//       "enum": [
	//         "accept",
	//         "cancel",
	//         "propose",
	//         "unknownAction",
	//         "updateFinalized"
	//       ],
	//       "enumDescriptions": [
	//         "",
	//         "",
	//         "",
	//         "",
	//         ""
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/{revisionNumber}/{updateAction}",
	//   "request": {
	//     "$ref": "MarketplaceOrder"
	//   },
	//   "response": {
	//     "$ref": "MarketplaceOrder"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceorders.search":

type MarketplaceordersSearchCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// Search: Search for orders using pql query
func (r *MarketplaceordersService) Search() *MarketplaceordersSearchCall {
	c := &MarketplaceordersSearchCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// PqlQuery sets the optional parameter "pqlQuery": Query string to
// retrieve specific orders.
func (c *MarketplaceordersSearchCall) PqlQuery(pqlQuery string) *MarketplaceordersSearchCall {
	c.opt_["pqlQuery"] = pqlQuery
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceordersSearchCall) Fields(s ...googleapi.Field) *MarketplaceordersSearchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MarketplaceordersSearchCall) IfNoneMatch(entityTag string) *MarketplaceordersSearchCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceordersSearchCall) Context(ctx context.Context) *MarketplaceordersSearchCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceordersSearchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["pqlQuery"]; ok {
		params.Set("pqlQuery", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/search")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceorders.search" call.
// Exactly one of *GetOrdersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetOrdersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceordersSearchCall) Do() (*GetOrdersResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetOrdersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Search for orders using pql query",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.marketplaceorders.search",
	//   "parameters": {
	//     "pqlQuery": {
	//       "description": "Query string to retrieve specific orders.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/search",
	//   "response": {
	//     "$ref": "GetOrdersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.marketplaceorders.update":

type MarketplaceordersUpdateCall struct {
	s                *Service
	orderId          string
	revisionNumber   int64
	updateAction     string
	marketplaceorder *MarketplaceOrder
	opt_             map[string]interface{}
	ctx_             context.Context
}

// Update: Update the given order
func (r *MarketplaceordersService) Update(orderId string, revisionNumber int64, updateAction string, marketplaceorder *MarketplaceOrder) *MarketplaceordersUpdateCall {
	c := &MarketplaceordersUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.orderId = orderId
	c.revisionNumber = revisionNumber
	c.updateAction = updateAction
	c.marketplaceorder = marketplaceorder
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MarketplaceordersUpdateCall) Fields(s ...googleapi.Field) *MarketplaceordersUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MarketplaceordersUpdateCall) Context(ctx context.Context) *MarketplaceordersUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *MarketplaceordersUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.marketplaceorder)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "marketplaceOrders/{orderId}/{revisionNumber}/{updateAction}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"orderId":        c.orderId,
		"revisionNumber": strconv.FormatInt(c.revisionNumber, 10),
		"updateAction":   c.updateAction,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.marketplaceorders.update" call.
// Exactly one of *MarketplaceOrder or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MarketplaceOrder.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *MarketplaceordersUpdateCall) Do() (*MarketplaceOrder, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MarketplaceOrder{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Update the given order",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer.marketplaceorders.update",
	//   "parameterOrder": [
	//     "orderId",
	//     "revisionNumber",
	//     "updateAction"
	//   ],
	//   "parameters": {
	//     "orderId": {
	//       "description": "The order id to update.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionNumber": {
	//       "description": "The last known revision number to update. If the head revision in the marketplace database has since changed, an error will be thrown. The caller should then fetch the lastest order at head revision and retry the update at that revision.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "updateAction": {
	//       "description": "The proposed action to take on the order.",
	//       "enum": [
	//         "accept",
	//         "cancel",
	//         "propose",
	//         "unknownAction",
	//         "updateFinalized"
	//       ],
	//       "enumDescriptions": [
	//         "",
	//         "",
	//         "",
	//         "",
	//         ""
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "marketplaceOrders/{orderId}/{revisionNumber}/{updateAction}",
	//   "request": {
	//     "$ref": "MarketplaceOrder"
	//   },
	//   "response": {
	//     "$ref": "MarketplaceOrder"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.negotiationrounds.insert":

type NegotiationroundsInsertCall struct {
	s                   *Service
	negotiationId       int64
	negotiationrounddto *NegotiationRoundDto
	opt_                map[string]interface{}
	ctx_                context.Context
}

// Insert: Adds the requested negotiationRound to the requested
// negotiation.
func (r *NegotiationroundsService) Insert(negotiationId int64, negotiationrounddto *NegotiationRoundDto) *NegotiationroundsInsertCall {
	c := &NegotiationroundsInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.negotiationId = negotiationId
	c.negotiationrounddto = negotiationrounddto
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *NegotiationroundsInsertCall) Fields(s ...googleapi.Field) *NegotiationroundsInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *NegotiationroundsInsertCall) Context(ctx context.Context) *NegotiationroundsInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *NegotiationroundsInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.negotiationrounddto)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "negotiations/{negotiationId}/negotiationrounds")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"negotiationId": strconv.FormatInt(c.negotiationId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.negotiationrounds.insert" call.
// Exactly one of *NegotiationRoundDto or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *NegotiationRoundDto.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *NegotiationroundsInsertCall) Do() (*NegotiationRoundDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &NegotiationRoundDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Adds the requested negotiationRound to the requested negotiation.",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.negotiationrounds.insert",
	//   "parameterOrder": [
	//     "negotiationId"
	//   ],
	//   "parameters": {
	//     "negotiationId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "negotiations/{negotiationId}/negotiationrounds",
	//   "request": {
	//     "$ref": "NegotiationRoundDto"
	//   },
	//   "response": {
	//     "$ref": "NegotiationRoundDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.negotiations.get":

type NegotiationsGetCall struct {
	s                         *Service
	negotiationId             int64
	getnegotiationbyidrequest *GetNegotiationByIdRequest
	opt_                      map[string]interface{}
	ctx_                      context.Context
}

// Get: Gets the requested negotiation.
func (r *NegotiationsService) Get(negotiationId int64, getnegotiationbyidrequest *GetNegotiationByIdRequest) *NegotiationsGetCall {
	c := &NegotiationsGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.negotiationId = negotiationId
	c.getnegotiationbyidrequest = getnegotiationbyidrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *NegotiationsGetCall) Fields(s ...googleapi.Field) *NegotiationsGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *NegotiationsGetCall) IfNoneMatch(entityTag string) *NegotiationsGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *NegotiationsGetCall) Context(ctx context.Context) *NegotiationsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *NegotiationsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "negotiations/{negotiationId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"negotiationId": strconv.FormatInt(c.negotiationId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.negotiations.get" call.
// Exactly one of *NegotiationDto or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *NegotiationDto.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *NegotiationsGetCall) Do() (*NegotiationDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &NegotiationDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the requested negotiation.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.negotiations.get",
	//   "parameterOrder": [
	//     "negotiationId"
	//   ],
	//   "parameters": {
	//     "negotiationId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "negotiations/{negotiationId}",
	//   "request": {
	//     "$ref": "GetNegotiationByIdRequest"
	//   },
	//   "response": {
	//     "$ref": "NegotiationDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.negotiations.insert":

type NegotiationsInsertCall struct {
	s              *Service
	negotiationdto *NegotiationDto
	opt_           map[string]interface{}
	ctx_           context.Context
}

// Insert: Creates or updates the requested negotiation.
func (r *NegotiationsService) Insert(negotiationdto *NegotiationDto) *NegotiationsInsertCall {
	c := &NegotiationsInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.negotiationdto = negotiationdto
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *NegotiationsInsertCall) Fields(s ...googleapi.Field) *NegotiationsInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *NegotiationsInsertCall) Context(ctx context.Context) *NegotiationsInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *NegotiationsInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.negotiationdto)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "negotiations")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.negotiations.insert" call.
// Exactly one of *NegotiationDto or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *NegotiationDto.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *NegotiationsInsertCall) Do() (*NegotiationDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &NegotiationDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates or updates the requested negotiation.",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.negotiations.insert",
	//   "path": "negotiations",
	//   "request": {
	//     "$ref": "NegotiationDto"
	//   },
	//   "response": {
	//     "$ref": "NegotiationDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.negotiations.list":

type NegotiationsListCall struct {
	s                      *Service
	getnegotiationsrequest *GetNegotiationsRequest
	opt_                   map[string]interface{}
	ctx_                   context.Context
}

// List: Lists all negotiations the authenticated user has access to.
func (r *NegotiationsService) List(getnegotiationsrequest *GetNegotiationsRequest) *NegotiationsListCall {
	c := &NegotiationsListCall{s: r.s, opt_: make(map[string]interface{})}
	c.getnegotiationsrequest = getnegotiationsrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *NegotiationsListCall) Fields(s ...googleapi.Field) *NegotiationsListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *NegotiationsListCall) IfNoneMatch(entityTag string) *NegotiationsListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *NegotiationsListCall) Context(ctx context.Context) *NegotiationsListCall {
	c.ctx_ = ctx
	return c
}

func (c *NegotiationsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "negotiations")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.negotiations.list" call.
// Exactly one of *GetNegotiationsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetNegotiationsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *NegotiationsListCall) Do() (*GetNegotiationsResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetNegotiationsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all negotiations the authenticated user has access to.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.negotiations.list",
	//   "path": "negotiations",
	//   "request": {
	//     "$ref": "GetNegotiationsRequest"
	//   },
	//   "response": {
	//     "$ref": "GetNegotiationsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.offers.get":

type OffersGetCall struct {
	s       *Service
	offerId int64
	opt_    map[string]interface{}
	ctx_    context.Context
}

// Get: Gets the requested offer.
func (r *OffersService) Get(offerId int64) *OffersGetCall {
	c := &OffersGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.offerId = offerId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OffersGetCall) Fields(s ...googleapi.Field) *OffersGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *OffersGetCall) IfNoneMatch(entityTag string) *OffersGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *OffersGetCall) Context(ctx context.Context) *OffersGetCall {
	c.ctx_ = ctx
	return c
}

func (c *OffersGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "offers/{offerId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"offerId": strconv.FormatInt(c.offerId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.offers.get" call.
// Exactly one of *OfferDto or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *OfferDto.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *OffersGetCall) Do() (*OfferDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &OfferDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets the requested offer.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.offers.get",
	//   "parameterOrder": [
	//     "offerId"
	//   ],
	//   "parameters": {
	//     "offerId": {
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "offers/{offerId}",
	//   "response": {
	//     "$ref": "OfferDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.offers.insert":

type OffersInsertCall struct {
	s        *Service
	offerdto *OfferDto
	opt_     map[string]interface{}
	ctx_     context.Context
}

// Insert: Creates or updates the requested offer.
func (r *OffersService) Insert(offerdto *OfferDto) *OffersInsertCall {
	c := &OffersInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.offerdto = offerdto
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OffersInsertCall) Fields(s ...googleapi.Field) *OffersInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *OffersInsertCall) Context(ctx context.Context) *OffersInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *OffersInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.offerdto)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "offers")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.offers.insert" call.
// Exactly one of *OfferDto or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *OfferDto.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *OffersInsertCall) Do() (*OfferDto, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &OfferDto{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates or updates the requested offer.",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.offers.insert",
	//   "path": "offers",
	//   "request": {
	//     "$ref": "OfferDto"
	//   },
	//   "response": {
	//     "$ref": "OfferDto"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.offers.list":

type OffersListCall struct {
	s                 *Service
	listoffersrequest *ListOffersRequest
	opt_              map[string]interface{}
	ctx_              context.Context
}

// List: Lists all offers the authenticated user has access to.
func (r *OffersService) List(listoffersrequest *ListOffersRequest) *OffersListCall {
	c := &OffersListCall{s: r.s, opt_: make(map[string]interface{})}
	c.listoffersrequest = listoffersrequest
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OffersListCall) Fields(s ...googleapi.Field) *OffersListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *OffersListCall) IfNoneMatch(entityTag string) *OffersListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *OffersListCall) Context(ctx context.Context) *OffersListCall {
	c.ctx_ = ctx
	return c
}

func (c *OffersListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "offers")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.offers.list" call.
// Exactly one of *ListOffersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListOffersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *OffersListCall) Do() (*ListOffersResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListOffersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all offers the authenticated user has access to.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.offers.list",
	//   "path": "offers",
	//   "request": {
	//     "$ref": "ListOffersRequest"
	//   },
	//   "response": {
	//     "$ref": "ListOffersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.performanceReport.list":

type PerformanceReportListCall struct {
	s             *Service
	accountId     int64
	endDateTime   string
	startDateTime string
	opt_          map[string]interface{}
	ctx_          context.Context
}

// List: Retrieves the authenticated user's list of performance metrics.
func (r *PerformanceReportService) List(accountId int64, endDateTime string, startDateTime string) *PerformanceReportListCall {
	c := &PerformanceReportListCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.endDateTime = endDateTime
	c.startDateTime = startDateTime
	return c
}

// MaxResults sets the optional parameter "maxResults": Maximum number
// of entries returned on one result page. If not set, the default is
// 100.
func (c *PerformanceReportListCall) MaxResults(maxResults int64) *PerformanceReportListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": A continuation
// token, used to page through performance reports. To retrieve the next
// page, set this parameter to the value of "nextPageToken" from the
// previous response.
func (c *PerformanceReportListCall) PageToken(pageToken string) *PerformanceReportListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PerformanceReportListCall) Fields(s ...googleapi.Field) *PerformanceReportListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PerformanceReportListCall) IfNoneMatch(entityTag string) *PerformanceReportListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PerformanceReportListCall) Context(ctx context.Context) *PerformanceReportListCall {
	c.ctx_ = ctx
	return c
}

func (c *PerformanceReportListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("accountId", fmt.Sprintf("%v", c.accountId))
	params.Set("endDateTime", fmt.Sprintf("%v", c.endDateTime))
	params.Set("startDateTime", fmt.Sprintf("%v", c.startDateTime))
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "performancereport")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.performanceReport.list" call.
// Exactly one of *PerformanceReportList or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PerformanceReportList.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PerformanceReportListCall) Do() (*PerformanceReportList, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PerformanceReportList{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves the authenticated user's list of performance metrics.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.performanceReport.list",
	//   "parameterOrder": [
	//     "accountId",
	//     "endDateTime",
	//     "startDateTime"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to get the reports.",
	//       "format": "int64",
	//       "location": "query",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "endDateTime": {
	//       "description": "The end time of the report in ISO 8601 timestamp format using UTC.",
	//       "location": "query",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "description": "Maximum number of entries returned on one result page. If not set, the default is 100. Optional.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "1000",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "A continuation token, used to page through performance reports. To retrieve the next page, set this parameter to the value of \"nextPageToken\" from the previous response. Optional.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "startDateTime": {
	//       "description": "The start time of the report in ISO 8601 timestamp format using UTC.",
	//       "location": "query",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "performancereport",
	//   "response": {
	//     "$ref": "PerformanceReportList"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.delete":

type PretargetingConfigDeleteCall struct {
	s         *Service
	accountId int64
	configId  int64
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Delete: Deletes an existing pretargeting config.
func (r *PretargetingConfigService) Delete(accountId int64, configId int64) *PretargetingConfigDeleteCall {
	c := &PretargetingConfigDeleteCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.configId = configId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigDeleteCall) Fields(s ...googleapi.Field) *PretargetingConfigDeleteCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigDeleteCall) Context(ctx context.Context) *PretargetingConfigDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}/{configId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"configId":  strconv.FormatInt(c.configId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.delete" call.
func (c *PretargetingConfigDeleteCall) Do() error {
	res, err := c.doRequest("json")
	if err != nil {
		return err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return err
	}
	return nil
	// {
	//   "description": "Deletes an existing pretargeting config.",
	//   "httpMethod": "DELETE",
	//   "id": "adexchangebuyer.pretargetingConfig.delete",
	//   "parameterOrder": [
	//     "accountId",
	//     "configId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to delete the pretargeting config for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "configId": {
	//       "description": "The specific id of the configuration to delete.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}/{configId}",
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.get":

type PretargetingConfigGetCall struct {
	s         *Service
	accountId int64
	configId  int64
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Get: Gets a specific pretargeting configuration
func (r *PretargetingConfigService) Get(accountId int64, configId int64) *PretargetingConfigGetCall {
	c := &PretargetingConfigGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.configId = configId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigGetCall) Fields(s ...googleapi.Field) *PretargetingConfigGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PretargetingConfigGetCall) IfNoneMatch(entityTag string) *PretargetingConfigGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigGetCall) Context(ctx context.Context) *PretargetingConfigGetCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}/{configId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"configId":  strconv.FormatInt(c.configId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.get" call.
// Exactly one of *PretargetingConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PretargetingConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PretargetingConfigGetCall) Do() (*PretargetingConfig, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PretargetingConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a specific pretargeting configuration",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.pretargetingConfig.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "configId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to get the pretargeting config for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "configId": {
	//       "description": "The specific id of the configuration to retrieve.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}/{configId}",
	//   "response": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.insert":

type PretargetingConfigInsertCall struct {
	s                  *Service
	accountId          int64
	pretargetingconfig *PretargetingConfig
	opt_               map[string]interface{}
	ctx_               context.Context
}

// Insert: Inserts a new pretargeting configuration.
func (r *PretargetingConfigService) Insert(accountId int64, pretargetingconfig *PretargetingConfig) *PretargetingConfigInsertCall {
	c := &PretargetingConfigInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.pretargetingconfig = pretargetingconfig
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigInsertCall) Fields(s ...googleapi.Field) *PretargetingConfigInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigInsertCall) Context(ctx context.Context) *PretargetingConfigInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.pretargetingconfig)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.insert" call.
// Exactly one of *PretargetingConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PretargetingConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PretargetingConfigInsertCall) Do() (*PretargetingConfig, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PretargetingConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Inserts a new pretargeting configuration.",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer.pretargetingConfig.insert",
	//   "parameterOrder": [
	//     "accountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to insert the pretargeting config for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}",
	//   "request": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "response": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.list":

type PretargetingConfigListCall struct {
	s         *Service
	accountId int64
	opt_      map[string]interface{}
	ctx_      context.Context
}

// List: Retrieves a list of the authenticated user's pretargeting
// configurations.
func (r *PretargetingConfigService) List(accountId int64) *PretargetingConfigListCall {
	c := &PretargetingConfigListCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigListCall) Fields(s ...googleapi.Field) *PretargetingConfigListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PretargetingConfigListCall) IfNoneMatch(entityTag string) *PretargetingConfigListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigListCall) Context(ctx context.Context) *PretargetingConfigListCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.list" call.
// Exactly one of *PretargetingConfigList or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PretargetingConfigList.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PretargetingConfigListCall) Do() (*PretargetingConfigList, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PretargetingConfigList{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves a list of the authenticated user's pretargeting configurations.",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer.pretargetingConfig.list",
	//   "parameterOrder": [
	//     "accountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to get the pretargeting configs for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}",
	//   "response": {
	//     "$ref": "PretargetingConfigList"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.patch":

type PretargetingConfigPatchCall struct {
	s                  *Service
	accountId          int64
	configId           int64
	pretargetingconfig *PretargetingConfig
	opt_               map[string]interface{}
	ctx_               context.Context
}

// Patch: Updates an existing pretargeting config. This method supports
// patch semantics.
func (r *PretargetingConfigService) Patch(accountId int64, configId int64, pretargetingconfig *PretargetingConfig) *PretargetingConfigPatchCall {
	c := &PretargetingConfigPatchCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.configId = configId
	c.pretargetingconfig = pretargetingconfig
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigPatchCall) Fields(s ...googleapi.Field) *PretargetingConfigPatchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigPatchCall) Context(ctx context.Context) *PretargetingConfigPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.pretargetingconfig)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}/{configId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"configId":  strconv.FormatInt(c.configId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.patch" call.
// Exactly one of *PretargetingConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PretargetingConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PretargetingConfigPatchCall) Do() (*PretargetingConfig, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PretargetingConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing pretargeting config. This method supports patch semantics.",
	//   "httpMethod": "PATCH",
	//   "id": "adexchangebuyer.pretargetingConfig.patch",
	//   "parameterOrder": [
	//     "accountId",
	//     "configId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to update the pretargeting config for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "configId": {
	//       "description": "The specific id of the configuration to update.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}/{configId}",
	//   "request": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "response": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer.pretargetingConfig.update":

type PretargetingConfigUpdateCall struct {
	s                  *Service
	accountId          int64
	configId           int64
	pretargetingconfig *PretargetingConfig
	opt_               map[string]interface{}
	ctx_               context.Context
}

// Update: Updates an existing pretargeting config.
func (r *PretargetingConfigService) Update(accountId int64, configId int64, pretargetingconfig *PretargetingConfig) *PretargetingConfigUpdateCall {
	c := &PretargetingConfigUpdateCall{s: r.s, opt_: make(map[string]interface{})}
	c.accountId = accountId
	c.configId = configId
	c.pretargetingconfig = pretargetingconfig
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PretargetingConfigUpdateCall) Fields(s ...googleapi.Field) *PretargetingConfigUpdateCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PretargetingConfigUpdateCall) Context(ctx context.Context) *PretargetingConfigUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *PretargetingConfigUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.pretargetingconfig)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "pretargetingconfigs/{accountId}/{configId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
		"configId":  strconv.FormatInt(c.configId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer.pretargetingConfig.update" call.
// Exactly one of *PretargetingConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *PretargetingConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PretargetingConfigUpdateCall) Do() (*PretargetingConfig, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &PretargetingConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing pretargeting config.",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer.pretargetingConfig.update",
	//   "parameterOrder": [
	//     "accountId",
	//     "configId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "The account id to update the pretargeting config for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "configId": {
	//       "description": "The specific id of the configuration to update.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "pretargetingconfigs/{accountId}/{configId}",
	//   "request": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "response": {
	//     "$ref": "PretargetingConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}
