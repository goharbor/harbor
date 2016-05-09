// Package plus provides access to the Google+ API.
//
// See https://developers.google.com/+/api/
//
// Usage example:
//
//   import "google.golang.org/api/plus/v1"
//   ...
//   plusService, err := plus.New(oauthHttpClient)
package plus // import "google.golang.org/api/plus/v1"

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

const apiId = "plus:v1"
const apiName = "plus"
const apiVersion = "v1"
const basePath = "https://www.googleapis.com/plus/v1/"

// OAuth2 scopes used by this API.
const (
	// Know your basic profile info and list of people in your circles.
	PlusLoginScope = "https://www.googleapis.com/auth/plus.login"

	// Know who you are on Google
	PlusMeScope = "https://www.googleapis.com/auth/plus.me"

	// View your email address
	UserinfoEmailScope = "https://www.googleapis.com/auth/userinfo.email"

	// View your basic profile info
	UserinfoProfileScope = "https://www.googleapis.com/auth/userinfo.profile"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Activities = NewActivitiesService(s)
	s.Comments = NewCommentsService(s)
	s.Moments = NewMomentsService(s)
	s.People = NewPeopleService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Activities *ActivitiesService

	Comments *CommentsService

	Moments *MomentsService

	People *PeopleService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewActivitiesService(s *Service) *ActivitiesService {
	rs := &ActivitiesService{s: s}
	return rs
}

type ActivitiesService struct {
	s *Service
}

func NewCommentsService(s *Service) *CommentsService {
	rs := &CommentsService{s: s}
	return rs
}

type CommentsService struct {
	s *Service
}

func NewMomentsService(s *Service) *MomentsService {
	rs := &MomentsService{s: s}
	return rs
}

type MomentsService struct {
	s *Service
}

func NewPeopleService(s *Service) *PeopleService {
	rs := &PeopleService{s: s}
	return rs
}

type PeopleService struct {
	s *Service
}

type Acl struct {
	// Description: Description of the access granted, suitable for display.
	Description string `json:"description,omitempty"`

	// Items: The list of access entries.
	Items []*PlusAclentryResource `json:"items,omitempty"`

	// Kind: Identifies this resource as a collection of access controls.
	// Value: "plus#acl".
	Kind string `json:"kind,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Acl) MarshalJSON() ([]byte, error) {
	type noMethod Acl
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Activity struct {
	// Access: Identifies who has access to see this activity.
	Access *Acl `json:"access,omitempty"`

	// Actor: The person who performed this activity.
	Actor *ActivityActor `json:"actor,omitempty"`

	// Address: Street address where this activity occurred.
	Address string `json:"address,omitempty"`

	// Annotation: Additional content added by the person who shared this
	// activity, applicable only when resharing an activity.
	Annotation string `json:"annotation,omitempty"`

	// CrosspostSource: If this activity is a crosspost from another system,
	// this property specifies the ID of the original activity.
	CrosspostSource string `json:"crosspostSource,omitempty"`

	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Geocode: Latitude and longitude where this activity occurred. Format
	// is latitude followed by longitude, space separated.
	Geocode string `json:"geocode,omitempty"`

	// Id: The ID of this activity.
	Id string `json:"id,omitempty"`

	// Kind: Identifies this resource as an activity. Value:
	// "plus#activity".
	Kind string `json:"kind,omitempty"`

	// Location: The location where this activity occurred.
	Location *Place `json:"location,omitempty"`

	// Object: The object of this activity.
	Object *ActivityObject `json:"object,omitempty"`

	// PlaceId: ID of the place where this activity occurred.
	PlaceId string `json:"placeId,omitempty"`

	// PlaceName: Name of the place where this activity occurred.
	PlaceName string `json:"placeName,omitempty"`

	// Provider: The service provider that initially published this
	// activity.
	Provider *ActivityProvider `json:"provider,omitempty"`

	// Published: The time at which this activity was initially published.
	// Formatted as an RFC 3339 timestamp.
	Published string `json:"published,omitempty"`

	// Radius: Radius, in meters, of the region where this activity
	// occurred, centered at the latitude and longitude identified in
	// geocode.
	Radius string `json:"radius,omitempty"`

	// Title: Title of this activity.
	Title string `json:"title,omitempty"`

	// Updated: The time at which this activity was last updated. Formatted
	// as an RFC 3339 timestamp.
	Updated string `json:"updated,omitempty"`

	// Url: The link to this activity.
	Url string `json:"url,omitempty"`

	// Verb: This activity's verb, which indicates the action that was
	// performed. Possible values include, but are not limited to, the
	// following values:
	// - "post" - Publish content to the stream.
	// - "share" - Reshare an activity.
	Verb string `json:"verb,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Access") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Activity) MarshalJSON() ([]byte, error) {
	type noMethod Activity
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActor: The person who performed this activity.
type ActivityActor struct {
	// ClientSpecificActorInfo: Actor info specific to particular clients.
	ClientSpecificActorInfo *ActivityActorClientSpecificActorInfo `json:"clientSpecificActorInfo,omitempty"`

	// DisplayName: The name of the actor, suitable for display.
	DisplayName string `json:"displayName,omitempty"`

	// Id: The ID of the actor's Person resource.
	Id string `json:"id,omitempty"`

	// Image: The image representation of the actor.
	Image *ActivityActorImage `json:"image,omitempty"`

	// Name: An object representation of the individual components of name.
	Name *ActivityActorName `json:"name,omitempty"`

	// Url: The link to the actor's Google profile.
	Url string `json:"url,omitempty"`

	// Verification: Verification status of actor.
	Verification *ActivityActorVerification `json:"verification,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "ClientSpecificActorInfo") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActor) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActor
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActorClientSpecificActorInfo: Actor info specific to
// particular clients.
type ActivityActorClientSpecificActorInfo struct {
	// YoutubeActorInfo: Actor info specific to YouTube clients.
	YoutubeActorInfo *ActivityActorClientSpecificActorInfoYoutubeActorInfo `json:"youtubeActorInfo,omitempty"`

	// ForceSendFields is a list of field names (e.g. "YoutubeActorInfo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActorClientSpecificActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActorClientSpecificActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActorClientSpecificActorInfoYoutubeActorInfo: Actor info
// specific to YouTube clients.
type ActivityActorClientSpecificActorInfoYoutubeActorInfo struct {
	// ChannelId: ID of the YouTube channel owned by the Actor.
	ChannelId string `json:"channelId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ChannelId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActorClientSpecificActorInfoYoutubeActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActorClientSpecificActorInfoYoutubeActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActorImage: The image representation of the actor.
type ActivityActorImage struct {
	// Url: The URL of the actor's profile photo. To resize the image and
	// crop it to a square, append the query string ?sz=x, where x is the
	// dimension in pixels of each side.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Url") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActorImage) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActorImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActorName: An object representation of the individual
// components of name.
type ActivityActorName struct {
	// FamilyName: The family name ("last name") of the actor.
	FamilyName string `json:"familyName,omitempty"`

	// GivenName: The given name ("first name") of the actor.
	GivenName string `json:"givenName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FamilyName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActorName) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActorName
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityActorVerification: Verification status of actor.
type ActivityActorVerification struct {
	// AdHocVerified: Verification for one-time or manual processes.
	AdHocVerified string `json:"adHocVerified,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AdHocVerified") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityActorVerification) MarshalJSON() ([]byte, error) {
	type noMethod ActivityActorVerification
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObject: The object of this activity.
type ActivityObject struct {
	// Actor: If this activity's object is itself another activity, such as
	// when a person reshares an activity, this property specifies the
	// original activity's actor.
	Actor *ActivityObjectActor `json:"actor,omitempty"`

	// Attachments: The media objects attached to this activity.
	Attachments []*ActivityObjectAttachments `json:"attachments,omitempty"`

	// Content: The HTML-formatted content, which is suitable for display.
	Content string `json:"content,omitempty"`

	// Id: The ID of the object. When resharing an activity, this is the ID
	// of the activity that is being reshared.
	Id string `json:"id,omitempty"`

	// ObjectType: The type of the object. Possible values include, but are
	// not limited to, the following values:
	// - "note" - Textual content.
	// - "activity" - A Google+ activity.
	ObjectType string `json:"objectType,omitempty"`

	// OriginalContent: The content (text) as provided by the author, which
	// is stored without any HTML formatting. When creating or updating an
	// activity, this value must be supplied as plain text in the request.
	OriginalContent string `json:"originalContent,omitempty"`

	// Plusoners: People who +1'd this activity.
	Plusoners *ActivityObjectPlusoners `json:"plusoners,omitempty"`

	// Replies: Comments in reply to this activity.
	Replies *ActivityObjectReplies `json:"replies,omitempty"`

	// Resharers: People who reshared this activity.
	Resharers *ActivityObjectResharers `json:"resharers,omitempty"`

	// Url: The URL that points to the linked resource.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Actor") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObject) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObject
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectActor: If this activity's object is itself another
// activity, such as when a person reshares an activity, this property
// specifies the original activity's actor.
type ActivityObjectActor struct {
	// ClientSpecificActorInfo: Actor info specific to particular clients.
	ClientSpecificActorInfo *ActivityObjectActorClientSpecificActorInfo `json:"clientSpecificActorInfo,omitempty"`

	// DisplayName: The original actor's name, which is suitable for
	// display.
	DisplayName string `json:"displayName,omitempty"`

	// Id: ID of the original actor.
	Id string `json:"id,omitempty"`

	// Image: The image representation of the original actor.
	Image *ActivityObjectActorImage `json:"image,omitempty"`

	// Url: A link to the original actor's Google profile.
	Url string `json:"url,omitempty"`

	// Verification: Verification status of actor.
	Verification *ActivityObjectActorVerification `json:"verification,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "ClientSpecificActorInfo") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectActor) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectActor
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectActorClientSpecificActorInfo: Actor info specific to
// particular clients.
type ActivityObjectActorClientSpecificActorInfo struct {
	// YoutubeActorInfo: Actor info specific to YouTube clients.
	YoutubeActorInfo *ActivityObjectActorClientSpecificActorInfoYoutubeActorInfo `json:"youtubeActorInfo,omitempty"`

	// ForceSendFields is a list of field names (e.g. "YoutubeActorInfo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectActorClientSpecificActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectActorClientSpecificActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectActorClientSpecificActorInfoYoutubeActorInfo: Actor
// info specific to YouTube clients.
type ActivityObjectActorClientSpecificActorInfoYoutubeActorInfo struct {
	// ChannelId: ID of the YouTube channel owned by the Actor.
	ChannelId string `json:"channelId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ChannelId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectActorClientSpecificActorInfoYoutubeActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectActorClientSpecificActorInfoYoutubeActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectActorImage: The image representation of the original
// actor.
type ActivityObjectActorImage struct {
	// Url: A URL that points to a thumbnail photo of the original actor.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Url") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectActorImage) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectActorImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectActorVerification: Verification status of actor.
type ActivityObjectActorVerification struct {
	// AdHocVerified: Verification for one-time or manual processes.
	AdHocVerified string `json:"adHocVerified,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AdHocVerified") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectActorVerification) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectActorVerification
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ActivityObjectAttachments struct {
	// Content: If the attachment is an article, this property contains a
	// snippet of text from the article. It can also include descriptions
	// for other types.
	Content string `json:"content,omitempty"`

	// DisplayName: The title of the attachment, such as a photo caption or
	// an article title.
	DisplayName string `json:"displayName,omitempty"`

	// Embed: If the attachment is a video, the embeddable link.
	Embed *ActivityObjectAttachmentsEmbed `json:"embed,omitempty"`

	// FullImage: The full image URL for photo attachments.
	FullImage *ActivityObjectAttachmentsFullImage `json:"fullImage,omitempty"`

	// Id: The ID of the attachment.
	Id string `json:"id,omitempty"`

	// Image: The preview image for photos or videos.
	Image *ActivityObjectAttachmentsImage `json:"image,omitempty"`

	// ObjectType: The type of media object. Possible values include, but
	// are not limited to, the following values:
	// - "photo" - A photo.
	// - "album" - A photo album.
	// - "video" - A video.
	// - "article" - An article, specified by a link.
	ObjectType string `json:"objectType,omitempty"`

	// Thumbnails: If the attachment is an album, this property is a list of
	// potential additional thumbnails from the album.
	Thumbnails []*ActivityObjectAttachmentsThumbnails `json:"thumbnails,omitempty"`

	// Url: The link to the attachment, which should be of type text/html.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Content") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachments) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachments
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectAttachmentsEmbed: If the attachment is a video, the
// embeddable link.
type ActivityObjectAttachmentsEmbed struct {
	// Type: Media type of the link.
	Type string `json:"type,omitempty"`

	// Url: URL of the link.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Type") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachmentsEmbed) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachmentsEmbed
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectAttachmentsFullImage: The full image URL for photo
// attachments.
type ActivityObjectAttachmentsFullImage struct {
	// Height: The height, in pixels, of the linked resource.
	Height int64 `json:"height,omitempty"`

	// Type: Media type of the link.
	Type string `json:"type,omitempty"`

	// Url: URL of the image.
	Url string `json:"url,omitempty"`

	// Width: The width, in pixels, of the linked resource.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachmentsFullImage) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachmentsFullImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectAttachmentsImage: The preview image for photos or
// videos.
type ActivityObjectAttachmentsImage struct {
	// Height: The height, in pixels, of the linked resource.
	Height int64 `json:"height,omitempty"`

	// Type: Media type of the link.
	Type string `json:"type,omitempty"`

	// Url: Image URL.
	Url string `json:"url,omitempty"`

	// Width: The width, in pixels, of the linked resource.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachmentsImage) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachmentsImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ActivityObjectAttachmentsThumbnails struct {
	// Description: Potential name of the thumbnail.
	Description string `json:"description,omitempty"`

	// Image: Image resource.
	Image *ActivityObjectAttachmentsThumbnailsImage `json:"image,omitempty"`

	// Url: URL of the webpage containing the image.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachmentsThumbnails) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachmentsThumbnails
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectAttachmentsThumbnailsImage: Image resource.
type ActivityObjectAttachmentsThumbnailsImage struct {
	// Height: The height, in pixels, of the linked resource.
	Height int64 `json:"height,omitempty"`

	// Type: Media type of the link.
	Type string `json:"type,omitempty"`

	// Url: Image url.
	Url string `json:"url,omitempty"`

	// Width: The width, in pixels, of the linked resource.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectAttachmentsThumbnailsImage) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectAttachmentsThumbnailsImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectPlusoners: People who +1'd this activity.
type ActivityObjectPlusoners struct {
	// SelfLink: The URL for the collection of people who +1'd this
	// activity.
	SelfLink string `json:"selfLink,omitempty"`

	// TotalItems: Total number of people who +1'd this activity.
	TotalItems int64 `json:"totalItems,omitempty"`

	// ForceSendFields is a list of field names (e.g. "SelfLink") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectPlusoners) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectPlusoners
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectReplies: Comments in reply to this activity.
type ActivityObjectReplies struct {
	// SelfLink: The URL for the collection of comments in reply to this
	// activity.
	SelfLink string `json:"selfLink,omitempty"`

	// TotalItems: Total number of comments on this activity.
	TotalItems int64 `json:"totalItems,omitempty"`

	// ForceSendFields is a list of field names (e.g. "SelfLink") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectReplies) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectReplies
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityObjectResharers: People who reshared this activity.
type ActivityObjectResharers struct {
	// SelfLink: The URL for the collection of resharers.
	SelfLink string `json:"selfLink,omitempty"`

	// TotalItems: Total number of people who reshared this activity.
	TotalItems int64 `json:"totalItems,omitempty"`

	// ForceSendFields is a list of field names (e.g. "SelfLink") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityObjectResharers) MarshalJSON() ([]byte, error) {
	type noMethod ActivityObjectResharers
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// ActivityProvider: The service provider that initially published this
// activity.
type ActivityProvider struct {
	// Title: Name of the service provider.
	Title string `json:"title,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Title") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityProvider) MarshalJSON() ([]byte, error) {
	type noMethod ActivityProvider
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ActivityFeed struct {
	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Id: The ID of this collection of activities. Deprecated.
	Id string `json:"id,omitempty"`

	// Items: The activities in this page of results.
	Items []*Activity `json:"items,omitempty"`

	// Kind: Identifies this resource as a collection of activities. Value:
	// "plus#activityFeed".
	Kind string `json:"kind,omitempty"`

	// NextLink: Link to the next page of activities.
	NextLink string `json:"nextLink,omitempty"`

	// NextPageToken: The continuation token, which is used to page through
	// large result sets. Provide this value in a subsequent request to
	// return the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// SelfLink: Link to this activity resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Title: The title of this collection of activities, which is a
	// truncated portion of the content.
	Title string `json:"title,omitempty"`

	// Updated: The time at which this collection of activities was last
	// updated. Formatted as an RFC 3339 timestamp.
	Updated string `json:"updated,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ActivityFeed) MarshalJSON() ([]byte, error) {
	type noMethod ActivityFeed
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Comment struct {
	// Actor: The person who posted this comment.
	Actor *CommentActor `json:"actor,omitempty"`

	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Id: The ID of this comment.
	Id string `json:"id,omitempty"`

	// InReplyTo: The activity this comment replied to.
	InReplyTo []*CommentInReplyTo `json:"inReplyTo,omitempty"`

	// Kind: Identifies this resource as a comment. Value: "plus#comment".
	Kind string `json:"kind,omitempty"`

	// Object: The object of this comment.
	Object *CommentObject `json:"object,omitempty"`

	// Plusoners: People who +1'd this comment.
	Plusoners *CommentPlusoners `json:"plusoners,omitempty"`

	// Published: The time at which this comment was initially published.
	// Formatted as an RFC 3339 timestamp.
	Published string `json:"published,omitempty"`

	// SelfLink: Link to this comment resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Updated: The time at which this comment was last updated. Formatted
	// as an RFC 3339 timestamp.
	Updated string `json:"updated,omitempty"`

	// Verb: This comment's verb, indicating what action was performed.
	// Possible values are:
	// - "post" - Publish content to the stream.
	Verb string `json:"verb,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Actor") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Comment) MarshalJSON() ([]byte, error) {
	type noMethod Comment
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentActor: The person who posted this comment.
type CommentActor struct {
	// ClientSpecificActorInfo: Actor info specific to particular clients.
	ClientSpecificActorInfo *CommentActorClientSpecificActorInfo `json:"clientSpecificActorInfo,omitempty"`

	// DisplayName: The name of this actor, suitable for display.
	DisplayName string `json:"displayName,omitempty"`

	// Id: The ID of the actor.
	Id string `json:"id,omitempty"`

	// Image: The image representation of this actor.
	Image *CommentActorImage `json:"image,omitempty"`

	// Url: A link to the Person resource for this actor.
	Url string `json:"url,omitempty"`

	// Verification: Verification status of actor.
	Verification *CommentActorVerification `json:"verification,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "ClientSpecificActorInfo") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentActor) MarshalJSON() ([]byte, error) {
	type noMethod CommentActor
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentActorClientSpecificActorInfo: Actor info specific to
// particular clients.
type CommentActorClientSpecificActorInfo struct {
	// YoutubeActorInfo: Actor info specific to YouTube clients.
	YoutubeActorInfo *CommentActorClientSpecificActorInfoYoutubeActorInfo `json:"youtubeActorInfo,omitempty"`

	// ForceSendFields is a list of field names (e.g. "YoutubeActorInfo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentActorClientSpecificActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod CommentActorClientSpecificActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentActorClientSpecificActorInfoYoutubeActorInfo: Actor info
// specific to YouTube clients.
type CommentActorClientSpecificActorInfoYoutubeActorInfo struct {
	// ChannelId: ID of the YouTube channel owned by the Actor.
	ChannelId string `json:"channelId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ChannelId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentActorClientSpecificActorInfoYoutubeActorInfo) MarshalJSON() ([]byte, error) {
	type noMethod CommentActorClientSpecificActorInfoYoutubeActorInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentActorImage: The image representation of this actor.
type CommentActorImage struct {
	// Url: The URL of the actor's profile photo. To resize the image and
	// crop it to a square, append the query string ?sz=x, where x is the
	// dimension in pixels of each side.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Url") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentActorImage) MarshalJSON() ([]byte, error) {
	type noMethod CommentActorImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentActorVerification: Verification status of actor.
type CommentActorVerification struct {
	// AdHocVerified: Verification for one-time or manual processes.
	AdHocVerified string `json:"adHocVerified,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AdHocVerified") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentActorVerification) MarshalJSON() ([]byte, error) {
	type noMethod CommentActorVerification
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CommentInReplyTo struct {
	// Id: The ID of the activity.
	Id string `json:"id,omitempty"`

	// Url: The URL of the activity.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Id") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentInReplyTo) MarshalJSON() ([]byte, error) {
	type noMethod CommentInReplyTo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentObject: The object of this comment.
type CommentObject struct {
	// Content: The HTML-formatted content, suitable for display.
	Content string `json:"content,omitempty"`

	// ObjectType: The object type of this comment. Possible values are:
	// - "comment" - A comment in reply to an activity.
	ObjectType string `json:"objectType,omitempty"`

	// OriginalContent: The content (text) as provided by the author, stored
	// without any HTML formatting. When creating or updating a comment,
	// this value must be supplied as plain text in the request.
	OriginalContent string `json:"originalContent,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Content") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentObject) MarshalJSON() ([]byte, error) {
	type noMethod CommentObject
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// CommentPlusoners: People who +1'd this comment.
type CommentPlusoners struct {
	// TotalItems: Total number of people who +1'd this comment.
	TotalItems int64 `json:"totalItems,omitempty"`

	// ForceSendFields is a list of field names (e.g. "TotalItems") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentPlusoners) MarshalJSON() ([]byte, error) {
	type noMethod CommentPlusoners
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type CommentFeed struct {
	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Id: The ID of this collection of comments.
	Id string `json:"id,omitempty"`

	// Items: The comments in this page of results.
	Items []*Comment `json:"items,omitempty"`

	// Kind: Identifies this resource as a collection of comments. Value:
	// "plus#commentFeed".
	Kind string `json:"kind,omitempty"`

	// NextLink: Link to the next page of activities.
	NextLink string `json:"nextLink,omitempty"`

	// NextPageToken: The continuation token, which is used to page through
	// large result sets. Provide this value in a subsequent request to
	// return the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Title: The title of this collection of comments.
	Title string `json:"title,omitempty"`

	// Updated: The time at which this collection of comments was last
	// updated. Formatted as an RFC 3339 timestamp.
	Updated string `json:"updated,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CommentFeed) MarshalJSON() ([]byte, error) {
	type noMethod CommentFeed
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type ItemScope struct {
	// About: The subject matter of the content.
	About *ItemScope `json:"about,omitempty"`

	// AdditionalName: An additional name for a Person, can be used for a
	// middle name.
	AdditionalName []string `json:"additionalName,omitempty"`

	// Address: Postal address.
	Address *ItemScope `json:"address,omitempty"`

	// AddressCountry: Address country.
	AddressCountry string `json:"addressCountry,omitempty"`

	// AddressLocality: Address locality.
	AddressLocality string `json:"addressLocality,omitempty"`

	// AddressRegion: Address region.
	AddressRegion string `json:"addressRegion,omitempty"`

	// AssociatedMedia: The encoding.
	AssociatedMedia []*ItemScope `json:"associated_media,omitempty"`

	// AttendeeCount: Number of attendees.
	AttendeeCount int64 `json:"attendeeCount,omitempty"`

	// Attendees: A person attending the event.
	Attendees []*ItemScope `json:"attendees,omitempty"`

	// Audio: From http://schema.org/MusicRecording, the audio file.
	Audio *ItemScope `json:"audio,omitempty"`

	// Author: The person or persons who created this result. In the example
	// of restaurant reviews, this might be the reviewer's name.
	Author []*ItemScope `json:"author,omitempty"`

	// BestRating: Best possible rating value that a result might obtain.
	// This property defines the upper bound for the ratingValue. For
	// example, you might have a 5 star rating scale, you would provide 5 as
	// the value for this property.
	BestRating string `json:"bestRating,omitempty"`

	// BirthDate: Date of birth.
	BirthDate string `json:"birthDate,omitempty"`

	// ByArtist: From http://schema.org/MusicRecording, the artist that
	// performed this recording.
	ByArtist *ItemScope `json:"byArtist,omitempty"`

	// Caption: The caption for this object.
	Caption string `json:"caption,omitempty"`

	// ContentSize: File size in (mega/kilo) bytes.
	ContentSize string `json:"contentSize,omitempty"`

	// ContentUrl: Actual bytes of the media object, for example the image
	// file or video file.
	ContentUrl string `json:"contentUrl,omitempty"`

	// Contributor: A list of contributors to this result.
	Contributor []*ItemScope `json:"contributor,omitempty"`

	// DateCreated: The date the result was created such as the date that a
	// review was first created.
	DateCreated string `json:"dateCreated,omitempty"`

	// DateModified: The date the result was last modified such as the date
	// that a review was last edited.
	DateModified string `json:"dateModified,omitempty"`

	// DatePublished: The initial date that the result was published. For
	// example, a user writes a comment on a blog, which has a
	// result.dateCreated of when they submit it. If the blog users comment
	// moderation, the result.datePublished value would match the date when
	// the owner approved the message.
	DatePublished string `json:"datePublished,omitempty"`

	// Description: The string that describes the content of the result.
	Description string `json:"description,omitempty"`

	// Duration: The duration of the item (movie, audio recording, event,
	// etc.) in ISO 8601 date format.
	Duration string `json:"duration,omitempty"`

	// EmbedUrl: A URL pointing to a player for a specific video. In
	// general, this is the information in the src element of an embed tag
	// and should not be the same as the content of the loc tag.
	EmbedUrl string `json:"embedUrl,omitempty"`

	// EndDate: The end date and time of the event (in ISO 8601 date
	// format).
	EndDate string `json:"endDate,omitempty"`

	// FamilyName: Family name. This property can be used with givenName
	// instead of the name property.
	FamilyName string `json:"familyName,omitempty"`

	// Gender: Gender of the person.
	Gender string `json:"gender,omitempty"`

	// Geo: Geo coordinates.
	Geo *ItemScope `json:"geo,omitempty"`

	// GivenName: Given name. This property can be used with familyName
	// instead of the name property.
	GivenName string `json:"givenName,omitempty"`

	// Height: The height of the media object.
	Height string `json:"height,omitempty"`

	// Id: An identifier for the object. Your app can choose how to identify
	// objects. The object.id is required if you are writing an action that
	// does not have a corresponding web page or object.url property.
	Id string `json:"id,omitempty"`

	// Image: A URL to the image that represents this result. For example,
	// if a user writes a review of a restaurant and attaches a photo of
	// their meal, you might use that photo as the result.image.
	Image string `json:"image,omitempty"`

	// InAlbum: From http://schema.org/MusicRecording, which album a song is
	// in.
	InAlbum *ItemScope `json:"inAlbum,omitempty"`

	// Kind: Identifies this resource as an itemScope.
	Kind string `json:"kind,omitempty"`

	// Latitude: Latitude.
	Latitude float64 `json:"latitude,omitempty"`

	// Location: The location of the event or organization.
	Location *ItemScope `json:"location,omitempty"`

	// Longitude: Longitude.
	Longitude float64 `json:"longitude,omitempty"`

	// Name: The name of the result. In the example of a restaurant review,
	// this might be the summary the user gave their review such as "Great
	// ambiance, but overpriced."
	Name string `json:"name,omitempty"`

	// PartOfTVSeries: Property of http://schema.org/TVEpisode indicating
	// which series the episode belongs to.
	PartOfTVSeries *ItemScope `json:"partOfTVSeries,omitempty"`

	// Performers: The main performer or performers of the event-for
	// example, a presenter, musician, or actor.
	Performers []*ItemScope `json:"performers,omitempty"`

	// PlayerType: Player type that is required. For example: Flash or
	// Silverlight.
	PlayerType string `json:"playerType,omitempty"`

	// PostOfficeBoxNumber: Post office box number.
	PostOfficeBoxNumber string `json:"postOfficeBoxNumber,omitempty"`

	// PostalCode: Postal code.
	PostalCode string `json:"postalCode,omitempty"`

	// RatingValue: Rating value.
	RatingValue string `json:"ratingValue,omitempty"`

	// ReviewRating: Review rating.
	ReviewRating *ItemScope `json:"reviewRating,omitempty"`

	// StartDate: The start date and time of the event (in ISO 8601 date
	// format).
	StartDate string `json:"startDate,omitempty"`

	// StreetAddress: Street address.
	StreetAddress string `json:"streetAddress,omitempty"`

	// Text: The text that is the result of the app activity. For example,
	// if a user leaves a review of a restaurant, this might be the text of
	// the review.
	Text string `json:"text,omitempty"`

	// Thumbnail: Thumbnail image for an image or video.
	Thumbnail *ItemScope `json:"thumbnail,omitempty"`

	// ThumbnailUrl: A URL to a thumbnail image that represents this result.
	ThumbnailUrl string `json:"thumbnailUrl,omitempty"`

	// TickerSymbol: The exchange traded instrument associated with a
	// Corporation object. The tickerSymbol is expressed as an exchange and
	// an instrument name separated by a space character. For the exchange
	// component of the tickerSymbol attribute, we recommend using the
	// controlled vocabulary of Market Identifier Codes (MIC) specified in
	// ISO15022.
	TickerSymbol string `json:"tickerSymbol,omitempty"`

	// Type: The schema.org URL that best describes the referenced object
	// and matches the type of moment.
	Type string `json:"type,omitempty"`

	// Url: The URL that points to the result object. For example, a
	// permalink directly to a restaurant reviewer's comment.
	Url string `json:"url,omitempty"`

	// Width: The width of the media object.
	Width string `json:"width,omitempty"`

	// WorstRating: Worst possible rating value that a result might obtain.
	// This property defines the lower bound for the ratingValue.
	WorstRating string `json:"worstRating,omitempty"`

	// ForceSendFields is a list of field names (e.g. "About") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ItemScope) MarshalJSON() ([]byte, error) {
	type noMethod ItemScope
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Moment struct {
	// Id: The moment ID.
	Id string `json:"id,omitempty"`

	// Kind: Identifies this resource as a moment.
	Kind string `json:"kind,omitempty"`

	// Object: The object on which the action was performed. Specifying this
	// is equivalent with specifying "target". Note that responses from the
	// server will use the "target" field instead for backward-compatibility
	// with older clients.
	Object *ItemScope `json:"object,omitempty"`

	// Result: The object generated by performing the action on the object.
	// For example, a user writes a review of a restaurant, the object is
	// the restaurant and the result is the review.
	Result *ItemScope `json:"result,omitempty"`

	// StartDate: Time stamp of when the action occurred in RFC3339 format.
	StartDate string `json:"startDate,omitempty"`

	// Target: The object on which the action was performed.
	Target *ItemScope `json:"target,omitempty"`

	// Type: The schema.org type for the type of moment to write. For
	// example, http://schema.org/AddAction. Note that responses from the
	// server will use the Google schema type instead for
	// backward-compatibility with older clients. For example,
	// http://schemas.google.com/AddActivity.
	Type string `json:"type,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Id") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Moment) MarshalJSON() ([]byte, error) {
	type noMethod Moment
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type MomentsFeed struct {
	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Items: The moments in this page of results.
	Items []*Moment `json:"items,omitempty"`

	// Kind: Identifies this resource as a collection of moments. Value:
	// "plus#momentsFeed".
	Kind string `json:"kind,omitempty"`

	// NextLink: Link to the next page of moments.
	NextLink string `json:"nextLink,omitempty"`

	// NextPageToken: The continuation token, which is used to page through
	// large result sets. Provide this value in a subsequent request to
	// return the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// SelfLink: Link to this page of moments.
	SelfLink string `json:"selfLink,omitempty"`

	// Title: The title of this collection of moments.
	Title string `json:"title,omitempty"`

	// Updated: The RFC 339 timestamp for when this collection of moments
	// was last updated.
	Updated string `json:"updated,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *MomentsFeed) MarshalJSON() ([]byte, error) {
	type noMethod MomentsFeed
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PeopleFeed struct {
	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Items: The people in this page of results. Each item includes the id,
	// displayName, image, and url for the person. To retrieve additional
	// profile data, see the people.get method.
	Items []*Person `json:"items,omitempty"`

	// Kind: Identifies this resource as a collection of people. Value:
	// "plus#peopleFeed".
	Kind string `json:"kind,omitempty"`

	// NextPageToken: The continuation token, which is used to page through
	// large result sets. Provide this value in a subsequent request to
	// return the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// SelfLink: Link to this resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Title: The title of this collection of people.
	Title string `json:"title,omitempty"`

	// TotalItems: The total number of people available in this list. The
	// number of people in a response might be smaller due to paging. This
	// might not be set for all collections.
	TotalItems int64 `json:"totalItems,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PeopleFeed) MarshalJSON() ([]byte, error) {
	type noMethod PeopleFeed
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Person struct {
	// AboutMe: A short biography for this person.
	AboutMe string `json:"aboutMe,omitempty"`

	// AgeRange: The age range of the person. Valid ranges are 17 or
	// younger, 18 to 20, and 21 or older. Age is determined from the user's
	// birthday using Western age reckoning.
	AgeRange *PersonAgeRange `json:"ageRange,omitempty"`

	// Birthday: The person's date of birth, represented as YYYY-MM-DD.
	Birthday string `json:"birthday,omitempty"`

	// BraggingRights: The "bragging rights" line of this person.
	BraggingRights string `json:"braggingRights,omitempty"`

	// CircledByCount: For followers who are visible, the number of people
	// who have added this person or page to a circle.
	CircledByCount int64 `json:"circledByCount,omitempty"`

	// Cover: The cover photo content.
	Cover *PersonCover `json:"cover,omitempty"`

	// CurrentLocation: (this field is not currently used)
	CurrentLocation string `json:"currentLocation,omitempty"`

	// DisplayName: The name of this person, which is suitable for display.
	DisplayName string `json:"displayName,omitempty"`

	// Domain: The hosted domain name for the user's Google Apps account.
	// For instance, example.com. The plus.profile.emails.read or email
	// scope is needed to get this domain name.
	Domain string `json:"domain,omitempty"`

	// Emails: A list of email addresses that this person has, including
	// their Google account email address, and the public verified email
	// addresses on their Google+ profile. The plus.profile.emails.read
	// scope is needed to retrieve these email addresses, or the email scope
	// can be used to retrieve just the Google account email address.
	Emails []*PersonEmails `json:"emails,omitempty"`

	// Etag: ETag of this response for caching purposes.
	Etag string `json:"etag,omitempty"`

	// Gender: The person's gender. Possible values include, but are not
	// limited to, the following values:
	// - "male" - Male gender.
	// - "female" - Female gender.
	// - "other" - Other.
	Gender string `json:"gender,omitempty"`

	// Id: The ID of this person.
	Id string `json:"id,omitempty"`

	// Image: The representation of the person's profile photo.
	Image *PersonImage `json:"image,omitempty"`

	// IsPlusUser: Whether this user has signed up for Google+.
	IsPlusUser bool `json:"isPlusUser,omitempty"`

	// Kind: Identifies this resource as a person. Value: "plus#person".
	Kind string `json:"kind,omitempty"`

	// Language: The user's preferred language for rendering.
	Language string `json:"language,omitempty"`

	// Name: An object representation of the individual components of a
	// person's name.
	Name *PersonName `json:"name,omitempty"`

	// Nickname: The nickname of this person.
	Nickname string `json:"nickname,omitempty"`

	// ObjectType: Type of person within Google+. Possible values include,
	// but are not limited to, the following values:
	// - "person" - represents an actual person.
	// - "page" - represents a page.
	ObjectType string `json:"objectType,omitempty"`

	// Occupation: The occupation of this person.
	Occupation string `json:"occupation,omitempty"`

	// Organizations: A list of current or past organizations with which
	// this person is associated.
	Organizations []*PersonOrganizations `json:"organizations,omitempty"`

	// PlacesLived: A list of places where this person has lived.
	PlacesLived []*PersonPlacesLived `json:"placesLived,omitempty"`

	// PlusOneCount: If a Google+ Page, the number of people who have +1'd
	// this page.
	PlusOneCount int64 `json:"plusOneCount,omitempty"`

	// RelationshipStatus: The person's relationship status. Possible values
	// include, but are not limited to, the following values:
	// - "single" - Person is single.
	// - "in_a_relationship" - Person is in a relationship.
	// - "engaged" - Person is engaged.
	// - "married" - Person is married.
	// - "its_complicated" - The relationship is complicated.
	// - "open_relationship" - Person is in an open relationship.
	// - "widowed" - Person is widowed.
	// - "in_domestic_partnership" - Person is in a domestic partnership.
	// - "in_civil_union" - Person is in a civil union.
	RelationshipStatus string `json:"relationshipStatus,omitempty"`

	// Skills: The person's skills.
	Skills string `json:"skills,omitempty"`

	// Tagline: The brief description (tagline) of this person.
	Tagline string `json:"tagline,omitempty"`

	// Url: The URL of this person's profile.
	Url string `json:"url,omitempty"`

	// Urls: A list of URLs for this person.
	Urls []*PersonUrls `json:"urls,omitempty"`

	// Verified: Whether the person or Google+ Page has been verified.
	Verified bool `json:"verified,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "AboutMe") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Person) MarshalJSON() ([]byte, error) {
	type noMethod Person
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonAgeRange: The age range of the person. Valid ranges are 17 or
// younger, 18 to 20, and 21 or older. Age is determined from the user's
// birthday using Western age reckoning.
type PersonAgeRange struct {
	// Max: The age range's upper bound, if any. Possible values include,
	// but are not limited to, the following:
	// - "17" - for age 17
	// - "20" - for age 20
	Max int64 `json:"max,omitempty"`

	// Min: The age range's lower bound, if any. Possible values include,
	// but are not limited to, the following:
	// - "21" - for age 21
	// - "18" - for age 18
	Min int64 `json:"min,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Max") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonAgeRange) MarshalJSON() ([]byte, error) {
	type noMethod PersonAgeRange
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonCover: The cover photo content.
type PersonCover struct {
	// CoverInfo: Extra information about the cover photo.
	CoverInfo *PersonCoverCoverInfo `json:"coverInfo,omitempty"`

	// CoverPhoto: The person's primary cover image.
	CoverPhoto *PersonCoverCoverPhoto `json:"coverPhoto,omitempty"`

	// Layout: The layout of the cover art. Possible values include, but are
	// not limited to, the following values:
	// - "banner" - One large image banner.
	Layout string `json:"layout,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CoverInfo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonCover) MarshalJSON() ([]byte, error) {
	type noMethod PersonCover
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonCoverCoverInfo: Extra information about the cover photo.
type PersonCoverCoverInfo struct {
	// LeftImageOffset: The difference between the left position of the
	// cover image and the actual displayed cover image. Only valid for
	// banner layout.
	LeftImageOffset int64 `json:"leftImageOffset,omitempty"`

	// TopImageOffset: The difference between the top position of the cover
	// image and the actual displayed cover image. Only valid for banner
	// layout.
	TopImageOffset int64 `json:"topImageOffset,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LeftImageOffset") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonCoverCoverInfo) MarshalJSON() ([]byte, error) {
	type noMethod PersonCoverCoverInfo
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonCoverCoverPhoto: The person's primary cover image.
type PersonCoverCoverPhoto struct {
	// Height: The height of the image.
	Height int64 `json:"height,omitempty"`

	// Url: The URL of the image.
	Url string `json:"url,omitempty"`

	// Width: The width of the image.
	Width int64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Height") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonCoverCoverPhoto) MarshalJSON() ([]byte, error) {
	type noMethod PersonCoverCoverPhoto
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PersonEmails struct {
	// Type: The type of address. Possible values include, but are not
	// limited to, the following values:
	// - "account" - Google account email address.
	// - "home" - Home email address.
	// - "work" - Work email address.
	// - "other" - Other.
	Type string `json:"type,omitempty"`

	// Value: The email address.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Type") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonEmails) MarshalJSON() ([]byte, error) {
	type noMethod PersonEmails
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonImage: The representation of the person's profile photo.
type PersonImage struct {
	// IsDefault: Whether the person's profile photo is the default one
	IsDefault bool `json:"isDefault,omitempty"`

	// Url: The URL of the person's profile photo. To resize the image and
	// crop it to a square, append the query string ?sz=x, where x is the
	// dimension in pixels of each side.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "IsDefault") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonImage) MarshalJSON() ([]byte, error) {
	type noMethod PersonImage
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PersonName: An object representation of the individual components of
// a person's name.
type PersonName struct {
	// FamilyName: The family name (last name) of this person.
	FamilyName string `json:"familyName,omitempty"`

	// Formatted: The full name of this person, including middle names,
	// suffixes, etc.
	Formatted string `json:"formatted,omitempty"`

	// GivenName: The given name (first name) of this person.
	GivenName string `json:"givenName,omitempty"`

	// HonorificPrefix: The honorific prefixes (such as "Dr." or "Mrs.") for
	// this person.
	HonorificPrefix string `json:"honorificPrefix,omitempty"`

	// HonorificSuffix: The honorific suffixes (such as "Jr.") for this
	// person.
	HonorificSuffix string `json:"honorificSuffix,omitempty"`

	// MiddleName: The middle name of this person.
	MiddleName string `json:"middleName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FamilyName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonName) MarshalJSON() ([]byte, error) {
	type noMethod PersonName
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PersonOrganizations struct {
	// Department: The department within the organization. Deprecated.
	Department string `json:"department,omitempty"`

	// Description: A short description of the person's role in this
	// organization. Deprecated.
	Description string `json:"description,omitempty"`

	// EndDate: The date that the person left this organization.
	EndDate string `json:"endDate,omitempty"`

	// Location: The location of this organization. Deprecated.
	Location string `json:"location,omitempty"`

	// Name: The name of the organization.
	Name string `json:"name,omitempty"`

	// Primary: If "true", indicates this organization is the person's
	// primary one, which is typically interpreted as the current one.
	Primary bool `json:"primary,omitempty"`

	// StartDate: The date that the person joined this organization.
	StartDate string `json:"startDate,omitempty"`

	// Title: The person's job title or role within the organization.
	Title string `json:"title,omitempty"`

	// Type: The type of organization. Possible values include, but are not
	// limited to, the following values:
	// - "work" - Work.
	// - "school" - School.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Department") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonOrganizations) MarshalJSON() ([]byte, error) {
	type noMethod PersonOrganizations
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PersonPlacesLived struct {
	// Primary: If "true", this place of residence is this person's primary
	// residence.
	Primary bool `json:"primary,omitempty"`

	// Value: A place where this person has lived. For example: "Seattle,
	// WA", "Near Toronto".
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Primary") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonPlacesLived) MarshalJSON() ([]byte, error) {
	type noMethod PersonPlacesLived
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PersonUrls struct {
	// Label: The label of the URL.
	Label string `json:"label,omitempty"`

	// Type: The type of URL. Possible values include, but are not limited
	// to, the following values:
	// - "otherProfile" - URL for another profile.
	// - "contributor" - URL to a site for which this person is a
	// contributor.
	// - "website" - URL for this Google+ Page's primary website.
	// - "other" - Other URL.
	Type string `json:"type,omitempty"`

	// Value: The URL value.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Label") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonUrls) MarshalJSON() ([]byte, error) {
	type noMethod PersonUrls
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type Place struct {
	// Address: The physical address of the place.
	Address *PlaceAddress `json:"address,omitempty"`

	// DisplayName: The display name of the place.
	DisplayName string `json:"displayName,omitempty"`

	// Id: The id of the place.
	Id string `json:"id,omitempty"`

	// Kind: Identifies this resource as a place. Value: "plus#place".
	Kind string `json:"kind,omitempty"`

	// Position: The position of the place.
	Position *PlacePosition `json:"position,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Address") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Place) MarshalJSON() ([]byte, error) {
	type noMethod Place
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PlaceAddress: The physical address of the place.
type PlaceAddress struct {
	// Formatted: The formatted address for display.
	Formatted string `json:"formatted,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Formatted") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PlaceAddress) MarshalJSON() ([]byte, error) {
	type noMethod PlaceAddress
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// PlacePosition: The position of the place.
type PlacePosition struct {
	// Latitude: The latitude of this position.
	Latitude float64 `json:"latitude,omitempty"`

	// Longitude: The longitude of this position.
	Longitude float64 `json:"longitude,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Latitude") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PlacePosition) MarshalJSON() ([]byte, error) {
	type noMethod PlacePosition
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

type PlusAclentryResource struct {
	// DisplayName: A descriptive name for this entry. Suitable for display.
	DisplayName string `json:"displayName,omitempty"`

	// Id: The ID of the entry. For entries of type "person" or "circle",
	// this is the ID of the resource. For other types, this property is not
	// set.
	Id string `json:"id,omitempty"`

	// Type: The type of entry describing to whom access is granted.
	// Possible values are:
	// - "person" - Access to an individual.
	// - "circle" - Access to members of a circle.
	// - "myCircles" - Access to members of all the person's circles.
	// - "extendedCircles" - Access to members of all the person's circles,
	// plus all of the people in their circles.
	// - "domain" - Access to members of the person's Google Apps domain.
	// - "public" - Access to anyone on the web.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PlusAclentryResource) MarshalJSON() ([]byte, error) {
	type noMethod PlusAclentryResource
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// method id "plus.activities.get":

type ActivitiesGetCall struct {
	s          *Service
	activityId string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// Get: Get an activity.
func (r *ActivitiesService) Get(activityId string) *ActivitiesGetCall {
	c := &ActivitiesGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.activityId = activityId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ActivitiesGetCall) Fields(s ...googleapi.Field) *ActivitiesGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ActivitiesGetCall) IfNoneMatch(entityTag string) *ActivitiesGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ActivitiesGetCall) Context(ctx context.Context) *ActivitiesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ActivitiesGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "activities/{activityId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"activityId": c.activityId,
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

// Do executes the "plus.activities.get" call.
// Exactly one of *Activity or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Activity.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ActivitiesGetCall) Do() (*Activity, error) {
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
	ret := &Activity{
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
	//   "description": "Get an activity.",
	//   "httpMethod": "GET",
	//   "id": "plus.activities.get",
	//   "parameterOrder": [
	//     "activityId"
	//   ],
	//   "parameters": {
	//     "activityId": {
	//       "description": "The ID of the activity to get.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "activities/{activityId}",
	//   "response": {
	//     "$ref": "Activity"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.activities.list":

type ActivitiesListCall struct {
	s          *Service
	userId     string
	collection string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// List: List all of the activities in the specified collection for a
// particular user.
func (r *ActivitiesService) List(userId string, collection string) *ActivitiesListCall {
	c := &ActivitiesListCall{s: r.s, opt_: make(map[string]interface{})}
	c.userId = userId
	c.collection = collection
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of activities to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *ActivitiesListCall) MaxResults(maxResults int64) *ActivitiesListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response.
func (c *ActivitiesListCall) PageToken(pageToken string) *ActivitiesListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ActivitiesListCall) Fields(s ...googleapi.Field) *ActivitiesListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ActivitiesListCall) IfNoneMatch(entityTag string) *ActivitiesListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ActivitiesListCall) Context(ctx context.Context) *ActivitiesListCall {
	c.ctx_ = ctx
	return c
}

func (c *ActivitiesListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people/{userId}/activities/{collection}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"userId":     c.userId,
		"collection": c.collection,
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

// Do executes the "plus.activities.list" call.
// Exactly one of *ActivityFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ActivityFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ActivitiesListCall) Do() (*ActivityFeed, error) {
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
	ret := &ActivityFeed{
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
	//   "description": "List all of the activities in the specified collection for a particular user.",
	//   "httpMethod": "GET",
	//   "id": "plus.activities.list",
	//   "parameterOrder": [
	//     "userId",
	//     "collection"
	//   ],
	//   "parameters": {
	//     "collection": {
	//       "description": "The collection of activities to list.",
	//       "enum": [
	//         "public"
	//       ],
	//       "enumDescriptions": [
	//         "All public activities created by the specified user."
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "20",
	//       "description": "The maximum number of activities to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "100",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "userId": {
	//       "description": "The ID of the user to get activities for. The special value \"me\" can be used to indicate the authenticated user.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people/{userId}/activities/{collection}",
	//   "response": {
	//     "$ref": "ActivityFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.activities.search":

type ActivitiesSearchCall struct {
	s     *Service
	query string
	opt_  map[string]interface{}
	ctx_  context.Context
}

// Search: Search public activities.
func (r *ActivitiesService) Search(query string) *ActivitiesSearchCall {
	c := &ActivitiesSearchCall{s: r.s, opt_: make(map[string]interface{})}
	c.query = query
	return c
}

// Language sets the optional parameter "language": Specify the
// preferred language to search with. See search language codes for
// available values.
func (c *ActivitiesSearchCall) Language(language string) *ActivitiesSearchCall {
	c.opt_["language"] = language
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of activities to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *ActivitiesSearchCall) MaxResults(maxResults int64) *ActivitiesSearchCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// OrderBy sets the optional parameter "orderBy": Specifies how to order
// search results.
//
// Possible values:
//   "best" - Sort activities by relevance to the user, most relevant
// first.
//   "recent" (default) - Sort activities by published date, most recent
// first.
func (c *ActivitiesSearchCall) OrderBy(orderBy string) *ActivitiesSearchCall {
	c.opt_["orderBy"] = orderBy
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response. This token can be of any
// length.
func (c *ActivitiesSearchCall) PageToken(pageToken string) *ActivitiesSearchCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ActivitiesSearchCall) Fields(s ...googleapi.Field) *ActivitiesSearchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ActivitiesSearchCall) IfNoneMatch(entityTag string) *ActivitiesSearchCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ActivitiesSearchCall) Context(ctx context.Context) *ActivitiesSearchCall {
	c.ctx_ = ctx
	return c
}

func (c *ActivitiesSearchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("query", fmt.Sprintf("%v", c.query))
	if v, ok := c.opt_["language"]; ok {
		params.Set("language", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["orderBy"]; ok {
		params.Set("orderBy", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "activities")
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

// Do executes the "plus.activities.search" call.
// Exactly one of *ActivityFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ActivityFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ActivitiesSearchCall) Do() (*ActivityFeed, error) {
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
	ret := &ActivityFeed{
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
	//   "description": "Search public activities.",
	//   "httpMethod": "GET",
	//   "id": "plus.activities.search",
	//   "parameterOrder": [
	//     "query"
	//   ],
	//   "parameters": {
	//     "language": {
	//       "default": "en-US",
	//       "description": "Specify the preferred language to search with. See search language codes for available values.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "10",
	//       "description": "The maximum number of activities to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "20",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "orderBy": {
	//       "default": "recent",
	//       "description": "Specifies how to order search results.",
	//       "enum": [
	//         "best",
	//         "recent"
	//       ],
	//       "enumDescriptions": [
	//         "Sort activities by relevance to the user, most relevant first.",
	//         "Sort activities by published date, most recent first."
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response. This token can be of any length.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "query": {
	//       "description": "Full-text search query string.",
	//       "location": "query",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "activities",
	//   "response": {
	//     "$ref": "ActivityFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.comments.get":

type CommentsGetCall struct {
	s         *Service
	commentId string
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Get: Get a comment.
func (r *CommentsService) Get(commentId string) *CommentsGetCall {
	c := &CommentsGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.commentId = commentId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CommentsGetCall) Fields(s ...googleapi.Field) *CommentsGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *CommentsGetCall) IfNoneMatch(entityTag string) *CommentsGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *CommentsGetCall) Context(ctx context.Context) *CommentsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *CommentsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "comments/{commentId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"commentId": c.commentId,
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

// Do executes the "plus.comments.get" call.
// Exactly one of *Comment or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Comment.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *CommentsGetCall) Do() (*Comment, error) {
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
	ret := &Comment{
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
	//   "description": "Get a comment.",
	//   "httpMethod": "GET",
	//   "id": "plus.comments.get",
	//   "parameterOrder": [
	//     "commentId"
	//   ],
	//   "parameters": {
	//     "commentId": {
	//       "description": "The ID of the comment to get.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "comments/{commentId}",
	//   "response": {
	//     "$ref": "Comment"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.comments.list":

type CommentsListCall struct {
	s          *Service
	activityId string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// List: List all of the comments for an activity.
func (r *CommentsService) List(activityId string) *CommentsListCall {
	c := &CommentsListCall{s: r.s, opt_: make(map[string]interface{})}
	c.activityId = activityId
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of comments to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *CommentsListCall) MaxResults(maxResults int64) *CommentsListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response.
func (c *CommentsListCall) PageToken(pageToken string) *CommentsListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// SortOrder sets the optional parameter "sortOrder": The order in which
// to sort the list of comments.
//
// Possible values:
//   "ascending" (default) - Sort oldest comments first.
//   "descending" - Sort newest comments first.
func (c *CommentsListCall) SortOrder(sortOrder string) *CommentsListCall {
	c.opt_["sortOrder"] = sortOrder
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CommentsListCall) Fields(s ...googleapi.Field) *CommentsListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *CommentsListCall) IfNoneMatch(entityTag string) *CommentsListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *CommentsListCall) Context(ctx context.Context) *CommentsListCall {
	c.ctx_ = ctx
	return c
}

func (c *CommentsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["sortOrder"]; ok {
		params.Set("sortOrder", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "activities/{activityId}/comments")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"activityId": c.activityId,
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

// Do executes the "plus.comments.list" call.
// Exactly one of *CommentFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *CommentFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *CommentsListCall) Do() (*CommentFeed, error) {
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
	ret := &CommentFeed{
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
	//   "description": "List all of the comments for an activity.",
	//   "httpMethod": "GET",
	//   "id": "plus.comments.list",
	//   "parameterOrder": [
	//     "activityId"
	//   ],
	//   "parameters": {
	//     "activityId": {
	//       "description": "The ID of the activity to get comments for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "20",
	//       "description": "The maximum number of comments to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "500",
	//       "minimum": "0",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sortOrder": {
	//       "default": "ascending",
	//       "description": "The order in which to sort the list of comments.",
	//       "enum": [
	//         "ascending",
	//         "descending"
	//       ],
	//       "enumDescriptions": [
	//         "Sort oldest comments first.",
	//         "Sort newest comments first."
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "activities/{activityId}/comments",
	//   "response": {
	//     "$ref": "CommentFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.moments.insert":

type MomentsInsertCall struct {
	s          *Service
	userId     string
	collection string
	moment     *Moment
	opt_       map[string]interface{}
	ctx_       context.Context
}

// Insert: Record a moment representing a user's action such as making a
// purchase or commenting on a blog.
func (r *MomentsService) Insert(userId string, collection string, moment *Moment) *MomentsInsertCall {
	c := &MomentsInsertCall{s: r.s, opt_: make(map[string]interface{})}
	c.userId = userId
	c.collection = collection
	c.moment = moment
	return c
}

// Debug sets the optional parameter "debug": Return the moment as
// written. Should be used only for debugging.
func (c *MomentsInsertCall) Debug(debug bool) *MomentsInsertCall {
	c.opt_["debug"] = debug
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MomentsInsertCall) Fields(s ...googleapi.Field) *MomentsInsertCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MomentsInsertCall) Context(ctx context.Context) *MomentsInsertCall {
	c.ctx_ = ctx
	return c
}

func (c *MomentsInsertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.moment)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["debug"]; ok {
		params.Set("debug", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people/{userId}/moments/{collection}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"userId":     c.userId,
		"collection": c.collection,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "plus.moments.insert" call.
// Exactly one of *Moment or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Moment.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *MomentsInsertCall) Do() (*Moment, error) {
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
	ret := &Moment{
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
	//   "description": "Record a moment representing a user's action such as making a purchase or commenting on a blog.",
	//   "httpMethod": "POST",
	//   "id": "plus.moments.insert",
	//   "parameterOrder": [
	//     "userId",
	//     "collection"
	//   ],
	//   "parameters": {
	//     "collection": {
	//       "description": "The collection to which to write moments.",
	//       "enum": [
	//         "vault"
	//       ],
	//       "enumDescriptions": [
	//         "The default collection for writing new moments."
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "debug": {
	//       "description": "Return the moment as written. Should be used only for debugging.",
	//       "location": "query",
	//       "type": "boolean"
	//     },
	//     "userId": {
	//       "description": "The ID of the user to record actions for. The only valid values are \"me\" and the ID of the authenticated user.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people/{userId}/moments/{collection}",
	//   "request": {
	//     "$ref": "Moment"
	//   },
	//   "response": {
	//     "$ref": "Moment"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.moments.list":

type MomentsListCall struct {
	s          *Service
	userId     string
	collection string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// List: List all of the moments for a particular user.
func (r *MomentsService) List(userId string, collection string) *MomentsListCall {
	c := &MomentsListCall{s: r.s, opt_: make(map[string]interface{})}
	c.userId = userId
	c.collection = collection
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of moments to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *MomentsListCall) MaxResults(maxResults int64) *MomentsListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response.
func (c *MomentsListCall) PageToken(pageToken string) *MomentsListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// TargetUrl sets the optional parameter "targetUrl": Only moments
// containing this targetUrl will be returned.
func (c *MomentsListCall) TargetUrl(targetUrl string) *MomentsListCall {
	c.opt_["targetUrl"] = targetUrl
	return c
}

// Type sets the optional parameter "type": Only moments of this type
// will be returned.
func (c *MomentsListCall) Type(type_ string) *MomentsListCall {
	c.opt_["type"] = type_
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MomentsListCall) Fields(s ...googleapi.Field) *MomentsListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MomentsListCall) IfNoneMatch(entityTag string) *MomentsListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *MomentsListCall) Context(ctx context.Context) *MomentsListCall {
	c.ctx_ = ctx
	return c
}

func (c *MomentsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["targetUrl"]; ok {
		params.Set("targetUrl", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["type"]; ok {
		params.Set("type", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people/{userId}/moments/{collection}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"userId":     c.userId,
		"collection": c.collection,
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

// Do executes the "plus.moments.list" call.
// Exactly one of *MomentsFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *MomentsFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *MomentsListCall) Do() (*MomentsFeed, error) {
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
	ret := &MomentsFeed{
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
	//   "description": "List all of the moments for a particular user.",
	//   "httpMethod": "GET",
	//   "id": "plus.moments.list",
	//   "parameterOrder": [
	//     "userId",
	//     "collection"
	//   ],
	//   "parameters": {
	//     "collection": {
	//       "description": "The collection of moments to list.",
	//       "enum": [
	//         "vault"
	//       ],
	//       "enumDescriptions": [
	//         "All moments created by the requesting application for the authenticated user."
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "20",
	//       "description": "The maximum number of moments to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "100",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "targetUrl": {
	//       "description": "Only moments containing this targetUrl will be returned.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "type": {
	//       "description": "Only moments of this type will be returned.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "userId": {
	//       "description": "The ID of the user to get moments for. The special value \"me\" can be used to indicate the authenticated user.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people/{userId}/moments/{collection}",
	//   "response": {
	//     "$ref": "MomentsFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.people.get":

type PeopleGetCall struct {
	s      *Service
	userId string
	opt_   map[string]interface{}
	ctx_   context.Context
}

// Get: Get a person's profile. If your app uses scope
// https://www.googleapis.com/auth/plus.login, this method is guaranteed
// to return ageRange and language.
func (r *PeopleService) Get(userId string) *PeopleGetCall {
	c := &PeopleGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.userId = userId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleGetCall) Fields(s ...googleapi.Field) *PeopleGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleGetCall) IfNoneMatch(entityTag string) *PeopleGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PeopleGetCall) Context(ctx context.Context) *PeopleGetCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people/{userId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"userId": c.userId,
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

// Do executes the "plus.people.get" call.
// Exactly one of *Person or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Person.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *PeopleGetCall) Do() (*Person, error) {
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
	ret := &Person{
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
	//   "description": "Get a person's profile. If your app uses scope https://www.googleapis.com/auth/plus.login, this method is guaranteed to return ageRange and language.",
	//   "httpMethod": "GET",
	//   "id": "plus.people.get",
	//   "parameterOrder": [
	//     "userId"
	//   ],
	//   "parameters": {
	//     "userId": {
	//       "description": "The ID of the person to get the profile for. The special value \"me\" can be used to indicate the authenticated user.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people/{userId}",
	//   "response": {
	//     "$ref": "Person"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me",
	//     "https://www.googleapis.com/auth/userinfo.email",
	//     "https://www.googleapis.com/auth/userinfo.profile"
	//   ]
	// }

}

// method id "plus.people.list":

type PeopleListCall struct {
	s          *Service
	userId     string
	collection string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// List: List all of the people in the specified collection.
func (r *PeopleService) List(userId string, collection string) *PeopleListCall {
	c := &PeopleListCall{s: r.s, opt_: make(map[string]interface{})}
	c.userId = userId
	c.collection = collection
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of people to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *PeopleListCall) MaxResults(maxResults int64) *PeopleListCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// OrderBy sets the optional parameter "orderBy": The order to return
// people in.
//
// Possible values:
//   "alphabetical" - Order the people by their display name.
//   "best" - Order people based on the relevence to the viewer.
func (c *PeopleListCall) OrderBy(orderBy string) *PeopleListCall {
	c.opt_["orderBy"] = orderBy
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response.
func (c *PeopleListCall) PageToken(pageToken string) *PeopleListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleListCall) Fields(s ...googleapi.Field) *PeopleListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleListCall) IfNoneMatch(entityTag string) *PeopleListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PeopleListCall) Context(ctx context.Context) *PeopleListCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["orderBy"]; ok {
		params.Set("orderBy", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people/{userId}/people/{collection}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"userId":     c.userId,
		"collection": c.collection,
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

// Do executes the "plus.people.list" call.
// Exactly one of *PeopleFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *PeopleFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PeopleListCall) Do() (*PeopleFeed, error) {
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
	ret := &PeopleFeed{
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
	//   "description": "List all of the people in the specified collection.",
	//   "httpMethod": "GET",
	//   "id": "plus.people.list",
	//   "parameterOrder": [
	//     "userId",
	//     "collection"
	//   ],
	//   "parameters": {
	//     "collection": {
	//       "description": "The collection of people to list.",
	//       "enum": [
	//         "connected",
	//         "visible"
	//       ],
	//       "enumDescriptions": [
	//         "The list of visible people in the authenticated user's circles who also use the requesting app. This list is limited to users who made their app activities visible to the authenticated user.",
	//         "The list of people who this user has added to one or more circles, limited to the circles visible to the requesting application."
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "100",
	//       "description": "The maximum number of people to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "100",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "orderBy": {
	//       "description": "The order to return people in.",
	//       "enum": [
	//         "alphabetical",
	//         "best"
	//       ],
	//       "enumDescriptions": [
	//         "Order the people by their display name.",
	//         "Order people based on the relevence to the viewer."
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "userId": {
	//       "description": "Get the collection of people for the person identified. Use \"me\" to indicate the authenticated user.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people/{userId}/people/{collection}",
	//   "response": {
	//     "$ref": "PeopleFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.people.listByActivity":

type PeopleListByActivityCall struct {
	s          *Service
	activityId string
	collection string
	opt_       map[string]interface{}
	ctx_       context.Context
}

// ListByActivity: List all of the people in the specified collection
// for a particular activity.
func (r *PeopleService) ListByActivity(activityId string, collection string) *PeopleListByActivityCall {
	c := &PeopleListByActivityCall{s: r.s, opt_: make(map[string]interface{})}
	c.activityId = activityId
	c.collection = collection
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of people to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *PeopleListByActivityCall) MaxResults(maxResults int64) *PeopleListByActivityCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response.
func (c *PeopleListByActivityCall) PageToken(pageToken string) *PeopleListByActivityCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleListByActivityCall) Fields(s ...googleapi.Field) *PeopleListByActivityCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleListByActivityCall) IfNoneMatch(entityTag string) *PeopleListByActivityCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PeopleListByActivityCall) Context(ctx context.Context) *PeopleListByActivityCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleListByActivityCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "activities/{activityId}/people/{collection}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"activityId": c.activityId,
		"collection": c.collection,
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

// Do executes the "plus.people.listByActivity" call.
// Exactly one of *PeopleFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *PeopleFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PeopleListByActivityCall) Do() (*PeopleFeed, error) {
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
	ret := &PeopleFeed{
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
	//   "description": "List all of the people in the specified collection for a particular activity.",
	//   "httpMethod": "GET",
	//   "id": "plus.people.listByActivity",
	//   "parameterOrder": [
	//     "activityId",
	//     "collection"
	//   ],
	//   "parameters": {
	//     "activityId": {
	//       "description": "The ID of the activity to get the list of people for.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "collection": {
	//       "description": "The collection of people to list.",
	//       "enum": [
	//         "plusoners",
	//         "resharers"
	//       ],
	//       "enumDescriptions": [
	//         "List all people who have +1'd this activity.",
	//         "List all people who have reshared this activity."
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "20",
	//       "description": "The maximum number of people to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "100",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "activities/{activityId}/people/{collection}",
	//   "response": {
	//     "$ref": "PeopleFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}

// method id "plus.people.search":

type PeopleSearchCall struct {
	s     *Service
	query string
	opt_  map[string]interface{}
	ctx_  context.Context
}

// Search: Search all public profiles.
func (r *PeopleService) Search(query string) *PeopleSearchCall {
	c := &PeopleSearchCall{s: r.s, opt_: make(map[string]interface{})}
	c.query = query
	return c
}

// Language sets the optional parameter "language": Specify the
// preferred language to search with. See search language codes for
// available values.
func (c *PeopleSearchCall) Language(language string) *PeopleSearchCall {
	c.opt_["language"] = language
	return c
}

// MaxResults sets the optional parameter "maxResults": The maximum
// number of people to include in the response, which is used for
// paging. For any response, the actual number returned might be less
// than the specified maxResults.
func (c *PeopleSearchCall) MaxResults(maxResults int64) *PeopleSearchCall {
	c.opt_["maxResults"] = maxResults
	return c
}

// PageToken sets the optional parameter "pageToken": The continuation
// token, which is used to page through large result sets. To get the
// next page of results, set this parameter to the value of
// "nextPageToken" from the previous response. This token can be of any
// length.
func (c *PeopleSearchCall) PageToken(pageToken string) *PeopleSearchCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleSearchCall) Fields(s ...googleapi.Field) *PeopleSearchCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleSearchCall) IfNoneMatch(entityTag string) *PeopleSearchCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *PeopleSearchCall) Context(ctx context.Context) *PeopleSearchCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleSearchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	params.Set("query", fmt.Sprintf("%v", c.query))
	if v, ok := c.opt_["language"]; ok {
		params.Set("language", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["maxResults"]; ok {
		params.Set("maxResults", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "people")
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

// Do executes the "plus.people.search" call.
// Exactly one of *PeopleFeed or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *PeopleFeed.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PeopleSearchCall) Do() (*PeopleFeed, error) {
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
	ret := &PeopleFeed{
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
	//   "description": "Search all public profiles.",
	//   "httpMethod": "GET",
	//   "id": "plus.people.search",
	//   "parameterOrder": [
	//     "query"
	//   ],
	//   "parameters": {
	//     "language": {
	//       "default": "en-US",
	//       "description": "Specify the preferred language to search with. See search language codes for available values.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "maxResults": {
	//       "default": "25",
	//       "description": "The maximum number of people to include in the response, which is used for paging. For any response, the actual number returned might be less than the specified maxResults.",
	//       "format": "uint32",
	//       "location": "query",
	//       "maximum": "50",
	//       "minimum": "1",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The continuation token, which is used to page through large result sets. To get the next page of results, set this parameter to the value of \"nextPageToken\" from the previous response. This token can be of any length.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "query": {
	//       "description": "Specify a query string for full text search of public text in all profiles.",
	//       "location": "query",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "people",
	//   "response": {
	//     "$ref": "PeopleFeed"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/plus.me"
	//   ]
	// }

}
