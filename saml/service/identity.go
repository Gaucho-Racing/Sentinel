package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/gaucho-racing/sentinel/saml/config"
	"github.com/gaucho-racing/sentinel/saml/pkg/sentinel"
)

// ErrAccessDenied is returned by CheckAccessGate when an entity is not in any
// of an application's required-flagged groups.
var ErrAccessDenied = errors.New("access denied: user does not meet the required group membership for this application")

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

// BuildSession assembles the SAML session for an entity, scoped to the SP's
// owning application. The NameID is the entity's email (the stable identifier
// SPs key on); identity attributes and the per-client filtered group set are
// attached for the assertion. Groups are exposed both as session.Groups (which
// the default assertion maker emits as eduPersonAffiliation) and as a plain
// `groups` attribute, since most relying parties key on the latter.
func BuildSession(entityID string, clientID string) (*saml.Session, error) {
	e, err := fetchEntity(entityID)
	if err != nil {
		return nil, err
	}
	groups, err := FilteredGroups(entityID, clientID)
	if err != nil {
		return nil, err
	}

	email := e.EmailAuth.Email
	if email == "" && e.User != nil {
		email = e.User.Email
	}

	session := &saml.Session{
		ID:           entityID,
		CreateTime:   time.Now(),
		ExpireTime:   time.Now().Add(time.Hour),
		Index:        entityID,
		NameID:       email,
		NameIDFormat: string(saml.EmailAddressNameIDFormat),
		SubjectID:    entityID,
		UserEmail:    email,
	}
	if e.User != nil {
		session.UserName = e.User.Username
		session.UserGivenName = e.User.FirstName
		session.UserSurname = e.User.LastName
		session.UserCommonName = strings.TrimSpace(e.User.FirstName + " " + e.User.LastName)
	}
	if email == "" {
		// Fall back to the entity ID as NameID so the assertion always carries a
		// subject, even for service accounts without an email auth record.
		session.NameID = entityID
		session.NameIDFormat = string(saml.UnspecifiedNameIDFormat)
	}

	names := make([]string, 0, len(groups))
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		session.Groups = append(session.Groups, g.Name)
		names = append(names, g.Name)
		ids = append(ids, g.ID)
	}
	session.CustomAttributes = []saml.Attribute{
		stringAttribute("groups", names),
		stringAttribute("group_ids", ids),
		stringAttribute("entity_id", []string{entityID}),
	}
	return session, nil
}

func stringAttribute(name string, values []string) saml.Attribute {
	vals := make([]saml.AttributeValue, 0, len(values))
	for _, v := range values {
		vals = append(vals, saml.AttributeValue{Type: "xs:string", Value: v})
	}
	return saml.Attribute{
		FriendlyName: name,
		Name:         name,
		NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
		Values:       vals,
	}
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
