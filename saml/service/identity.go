package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/model"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
)

// ErrAccessDenied is returned by CheckAccessGate when an entity is not in any
// of an application's required-flagged groups.
var (
	ErrAccessDenied = errors.New("access denied: user does not meet the required group membership for this application")
	ErrNameIDEmpty  = errors.New("SAML NameID resolved to an empty value")
	ErrNameIDEmail  = errors.New("SAML NameID is not a valid email address")
)

type entity struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	EmailAuth struct {
		Email string `json:"email"`
	} `json:"email_auth"`
	User *struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

type groupResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupRef is a group's stable ID and human-readable name.
type GroupRef struct {
	ID   string
	Name string
}

type applicationGroupLink struct {
	GroupID  string `json:"id"`
	Required bool   `json:"required"`
}

type AssertionPreview struct {
	NameID       string              `json:"name_id"`
	NameIDFormat string              `json:"name_id_format"`
	Attributes   []ResolvedAttribute `json:"attributes"`
}

type ResolvedAttribute struct {
	Name         string   `json:"name"`
	FriendlyName string   `json:"friendly_name"`
	NameFormat   string   `json:"name_format"`
	Values       []string `json:"values"`
}

type assertionIdentity struct {
	EntityID    string
	Email       string
	Username    string
	FirstName   string
	LastName    string
	DisplayName string
	GroupNames  []string
	GroupIDs    []string
}

func DefaultAttributeMappings() model.AttributeMappings {
	return model.AttributeMappings{
		{Name: "urn:oid:0.9.2342.19200300.100.1.1", FriendlyName: "uid", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceUsername, OmitIfEmpty: true},
		{Name: "urn:oid:0.9.2342.19200300.100.1.3", FriendlyName: "mail", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceEmail, OmitIfEmpty: true},
		{Name: "urn:oid:1.3.6.1.4.1.5923.1.1.1.6", FriendlyName: "eduPersonPrincipalName", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceEmail, OmitIfEmpty: true},
		{Name: "urn:oid:2.5.4.4", FriendlyName: "sn", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceLastName, OmitIfEmpty: true},
		{Name: "urn:oid:2.5.4.42", FriendlyName: "givenName", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceFirstName, OmitIfEmpty: true},
		{Name: "urn:oid:2.5.4.3", FriendlyName: "cn", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceDisplayName, OmitIfEmpty: true},
		{Name: "urn:oid:1.3.6.1.4.1.5923.1.1.1.1", FriendlyName: "eduPersonAffiliation", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceGroupNames},
		{Name: "groups", FriendlyName: "groups", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceGroupNames},
		{Name: "group_ids", FriendlyName: "group_ids", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceGroupIDs},
		{Name: "entity_id", FriendlyName: "entity_id", NameFormat: model.AttributeNameFormatBasic, Source: model.AttributeSourceEntityID},
		{Name: "urn:oasis:names:tc:SAML:attribute:subject-id", NameFormat: model.AttributeNameFormatURI, Source: model.AttributeSourceEntityID},
	}
}

func PreviewAssertion(applicationID, entityID string) (AssertionPreview, error) {
	sp, err := GetServiceProvider(applicationID)
	if err != nil {
		return AssertionPreview{}, err
	}
	resolved, err := resolveSP(sp)
	if err != nil {
		return AssertionPreview{}, err
	}
	return resolveAssertion(entityID, resolved)
}

func PreviewAssertionConfiguration(sp model.ServiceProvider, entityID string) (AssertionPreview, error) {
	sp, err := NormalizeServiceProvider(sp)
	if err != nil {
		return AssertionPreview{}, err
	}
	resolved, err := resolveSP(sp)
	if err != nil {
		return AssertionPreview{}, err
	}
	return resolveAssertion(entityID, resolved)
}

func BuildSession(entityID string, sp ResolvedSP) (*saml.Session, error) {
	preview, err := resolveAssertion(entityID, sp)
	if err != nil {
		return nil, err
	}
	attributes := make([]saml.Attribute, 0, len(preview.Attributes))
	for _, attribute := range preview.Attributes {
		attributes = append(attributes, stringAttribute(attribute))
	}
	now := time.Now().UTC()
	return &saml.Session{
		ID:               entityID,
		CreateTime:       now,
		ExpireTime:       now.Add(time.Hour),
		Index:            entityID,
		NameID:           preview.NameID,
		NameIDFormat:     preview.NameIDFormat,
		CustomAttributes: attributes,
	}, nil
}

func resolveAssertion(entityID string, sp ResolvedSP) (AssertionPreview, error) {
	e, err := fetchEntity(entityID)
	if err != nil {
		return AssertionPreview{}, err
	}
	groups, err := FilteredGroups(entityID, sp.ClientID)
	if err != nil {
		return AssertionPreview{}, err
	}
	identity := assertionIdentity{EntityID: e.ID, Email: e.EmailAuth.Email}
	if e.User != nil {
		identity.Username = e.User.Username
		identity.FirstName = e.User.FirstName
		identity.LastName = e.User.LastName
		identity.DisplayName = strings.TrimSpace(e.User.FirstName + " " + e.User.LastName)
		if identity.Email == "" {
			identity.Email = e.User.Email
		}
	}
	for _, group := range groups {
		identity.GroupNames = append(identity.GroupNames, group.Name)
		identity.GroupIDs = append(identity.GroupIDs, group.ID)
	}
	return resolveAssertionIdentity(identity, sp)
}

func resolveAssertionIdentity(identity assertionIdentity, sp ResolvedSP) (AssertionPreview, error) {
	nameID := nameIDValue(sp.NameIDSource, identity)
	if nameID == "" {
		return AssertionPreview{}, ErrNameIDEmpty
	}
	if sp.Profile == model.ProfileAWSIdentityCenter {
		address, err := mail.ParseAddress(nameID)
		if err != nil || address.Address != nameID {
			return AssertionPreview{}, ErrNameIDEmail
		}
	}
	preview := AssertionPreview{
		NameID:       nameID,
		NameIDFormat: sp.NameIDFormat,
		Attributes:   make([]ResolvedAttribute, 0, len(sp.AttributeMappings)),
	}
	for _, mapping := range sp.AttributeMappings {
		values := attributeValues(mapping, identity)
		if mapping.OmitIfEmpty && allValuesEmpty(values) {
			continue
		}
		preview.Attributes = append(preview.Attributes, ResolvedAttribute{
			Name:         mapping.Name,
			FriendlyName: mapping.FriendlyName,
			NameFormat:   mapping.NameFormat,
			Values:       values,
		})
	}
	return preview, nil
}

func stringAttribute(attribute ResolvedAttribute) saml.Attribute {
	vals := make([]saml.AttributeValue, 0, len(attribute.Values))
	for _, v := range attribute.Values {
		vals = append(vals, saml.AttributeValue{Type: "xs:string", Value: v})
	}
	return saml.Attribute{
		FriendlyName: attribute.FriendlyName,
		Name:         attribute.Name,
		NameFormat:   attribute.NameFormat,
		Values:       vals,
	}
}

func nameIDValue(source model.NameIDSource, identity assertionIdentity) string {
	switch source {
	case model.NameIDSourceEmail:
		return identity.Email
	case model.NameIDSourceUsername:
		return identity.Username
	case model.NameIDSourceEntityID:
		return identity.EntityID
	default:
		return ""
	}
}

func attributeValues(mapping model.AttributeMapping, identity assertionIdentity) []string {
	switch mapping.Source {
	case model.AttributeSourceEmail:
		return []string{identity.Email}
	case model.AttributeSourceUsername:
		return []string{identity.Username}
	case model.AttributeSourceFirstName:
		return []string{identity.FirstName}
	case model.AttributeSourceLastName:
		return []string{identity.LastName}
	case model.AttributeSourceDisplayName:
		return []string{identity.DisplayName}
	case model.AttributeSourceEntityID:
		return []string{identity.EntityID}
	case model.AttributeSourceGroupNames:
		return append([]string(nil), identity.GroupNames...)
	case model.AttributeSourceGroupIDs:
		return append([]string(nil), identity.GroupIDs...)
	case model.AttributeSourceConstant:
		return []string{mapping.Constant}
	default:
		return nil
	}
}

func allValuesEmpty(values []string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if value != "" {
			return false
		}
	}
	return true
}

// CheckAccessGate returns ErrAccessDenied when the entity is in none of the
// application's required-flagged groups. Apps with no required links are open.
// Fails closed: a non-ErrAccessDenied error means the gate couldn't be
// evaluated and callers must treat it as a denial.
func CheckAccessGate(entityID, clientID string) error {
	links, err := getAppGroupLinks(clientID)
	if err != nil {
		return err
	}
	required := make([]string, 0, len(links))
	for _, link := range links {
		if link.Required {
			required = append(required, link.GroupID)
		}
	}
	if len(required) == 0 {
		return nil
	}
	userGroups, err := getEntityGroups(entityID)
	if err != nil {
		return err
	}
	user := make(map[string]struct{}, len(userGroups))
	for _, g := range userGroups {
		user[g.ID] = struct{}{}
	}
	for _, g := range required {
		if _, ok := user[g]; ok {
			return nil
		}
	}
	return ErrAccessDenied
}

// FilteredGroups resolves the groups an entity should expose to a client: the
// Sentinel client sees all of the user's groups; any other client sees the
// user's groups intersected with the union of the client's linked groups and
// Sentinel's linked groups (the global default).
func FilteredGroups(entityID string, clientID string) ([]GroupRef, error) {
	userGroups, err := getEntityGroups(entityID)
	if err != nil {
		return nil, err
	}
	if clientID == config.SentinelClientID {
		return userGroups, nil
	}
	allowed := map[string]struct{}{}
	appLinks, err := getAppGroupLinks(clientID)
	if err != nil {
		return nil, err
	}
	sentinelLinks, err := getAppGroupLinks(config.SentinelClientID)
	if err != nil {
		return nil, err
	}
	for _, link := range append(appLinks, sentinelLinks...) {
		allowed[link.GroupID] = struct{}{}
	}
	filtered := make([]GroupRef, 0, len(userGroups))
	for _, g := range userGroups {
		if _, ok := allowed[g.ID]; ok {
			filtered = append(filtered, g)
		}
	}
	return filtered, nil
}

func fetchEntity(entityID string) (entity, error) {
	var e entity
	if err := sentinel.Get("/api/core/entity/"+entityID, &e); err != nil {
		return entity{}, fmt.Errorf("load entity %s: %w", entityID, err)
	}
	return e, nil
}

func getEntityGroups(entityID string) ([]GroupRef, error) {
	var groups []groupResponse
	if err := sentinel.Get("/api/core/entity/"+entityID+"/groups", &groups); err != nil {
		return nil, fmt.Errorf("load groups for entity %s: %w", entityID, err)
	}
	refs := make([]GroupRef, 0, len(groups))
	for _, g := range groups {
		refs = append(refs, GroupRef{ID: g.ID, Name: g.Name})
	}
	return refs, nil
}

func getAppGroupLinks(clientID string) ([]applicationGroupLink, error) {
	if clientID == "" {
		return nil, nil
	}
	var links []applicationGroupLink
	if err := sentinel.Get("/api/core/applications/client/"+clientID+"/groups", &links); err != nil {
		return nil, fmt.Errorf("load group links for client %s: %w", clientID, err)
	}
	return links, nil
}
