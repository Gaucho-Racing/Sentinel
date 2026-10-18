package service

import (
	"sort"
	"strings"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/nyaruka/phonenumbers"
)

const minDomainMembers = 3

type RegionCount struct {
	Code     string `json:"code"`
	Name     string `json:"name,omitempty"`
	Count    int    `json:"count"`
	Verified int    `json:"verified"`
}

type AreaCodeCount struct {
	AreaCode string `json:"area_code"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Count    int    `json:"count"`
}

type PhoneAnalytics struct {
	Total       int             `json:"total"`
	Verified    int             `json:"verified"`
	Unparseable int             `json:"unparseable"`
	Countries   []RegionCount   `json:"countries"`
	USStates    []RegionCount   `json:"us_states"`
	AreaCodes   []AreaCodeCount `json:"area_codes"`
}

type DomainCount struct {
	Domain   string `json:"domain"`
	Kind     string `json:"kind"`
	Count    int    `json:"count"`
	Verified int    `json:"verified"`
}

type EmailAnalytics struct {
	Total    int             `json:"total"`
	Verified int             `json:"verified"`
	Domains  []DomainCount   `json:"domains"`
	Kinds    []CategoryCount `json:"kinds"`
}

type ContactAnalytics struct {
	Phones PhoneAnalytics `json:"phones"`
	Emails EmailAnalytics `json:"emails"`
}

type contactRow struct {
	Value    string
	Verified bool
}

func GetContactAnalytics() (ContactAnalytics, error) {
	var phones []contactRow
	if err := database.DB.Raw(`
		SELECT p.phone_number AS value, p.verified_at IS NOT NULL AS verified
		FROM auth_entity_phone p
		JOIN auth_entity e ON e.id = p.entity_id AND e.type = 'USER'
		WHERE p.phone_number <> ''
	`).Scan(&phones).Error; err != nil {
		return ContactAnalytics{}, err
	}
	var emails []contactRow
	if err := database.DB.Raw(`
		SELECT LOWER(m.email) AS value, m.verified_at IS NOT NULL AS verified
		FROM auth_entity_email m
		JOIN auth_entity e ON e.id = m.entity_id AND e.type = 'USER'
		WHERE m.email <> ''
	`).Scan(&emails).Error; err != nil {
		return ContactAnalytics{}, err
	}
	return ContactAnalytics{Phones: phoneAnalytics(phones), Emails: emailAnalytics(emails)}, nil
}

func phoneAnalytics(rows []contactRow) PhoneAnalytics {
	result := PhoneAnalytics{Countries: []RegionCount{}, USStates: []RegionCount{}, AreaCodes: []AreaCodeCount{}}
	countries := map[string]*RegionCount{}
	states := map[string]*RegionCount{}
	areaCodes := map[string]*AreaCodeCount{}

	for _, row := range rows {
		result.Total++
		if row.Verified {
			result.Verified++
		}
		number, err := phonenumbers.Parse(row.Value, "US")
		if err != nil || !phonenumbers.IsValidNumber(number) {
			result.Unparseable++
			continue
		}
		country := phonenumbers.GetRegionCodeForNumber(number)
		bump(countries, country, "", row.Verified)

		geo, _ := phonenumbers.GetGeocodingForNumber(number, "en")
		if country == "US" {
			if code, name, ok := usState(geo); ok {
				bump(states, code, name, row.Verified)
				if geo == "" {
					geo = name
				}
			}
		}
		if number.GetCountryCode() == 1 {
			national := phonenumbers.GetNationalSignificantNumber(number)
			if len(national) == 10 {
				areaCode := national[:3]
				entry, ok := areaCodes[areaCode]
				if !ok {
					entry = &AreaCodeCount{AreaCode: areaCode, Country: country, Region: areaCodeRegion(country, geo)}
					areaCodes[areaCode] = entry
				}
				entry.Count++
			}
		}
	}

	result.Countries = sortedRegions(countries)
	result.USStates = sortedRegions(states)
	for _, entry := range areaCodes {
		result.AreaCodes = append(result.AreaCodes, *entry)
	}
	sort.Slice(result.AreaCodes, func(i, j int) bool {
		if result.AreaCodes[i].Count != result.AreaCodes[j].Count {
			return result.AreaCodes[i].Count > result.AreaCodes[j].Count
		}
		return result.AreaCodes[i].AreaCode < result.AreaCodes[j].AreaCode
	})
	return result
}

func bump(regions map[string]*RegionCount, code string, name string, verified bool) {
	entry, ok := regions[code]
	if !ok {
		entry = &RegionCount{Code: code, Name: name}
		regions[code] = entry
	}
	entry.Count++
	if verified {
		entry.Verified++
	}
}

func sortedRegions(regions map[string]*RegionCount) []RegionCount {
	sorted := make([]RegionCount, 0, len(regions))
	for _, entry := range regions {
		sorted = append(sorted, *entry)
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}
		return sorted[i].Code < sorted[j].Code
	})
	return sorted
}

// areaCodeRegion labels an area code by its state or province rather than a
// city, since one area code spans many cities.
func areaCodeRegion(country string, geo string) string {
	if country == "US" {
		if _, name, ok := usState(geo); ok {
			return name
		}
	}
	if i := strings.LastIndex(geo, ", "); i >= 0 {
		return geo[i+2:]
	}
	return geo
}

func emailAnalytics(rows []contactRow) EmailAnalytics {
	result := EmailAnalytics{Domains: []DomainCount{}, Kinds: []CategoryCount{}}
	domains := map[string]*DomainCount{}
	kinds := map[string]int{}
	for _, row := range rows {
		result.Total++
		if row.Verified {
			result.Verified++
		}
		at := strings.LastIndex(row.Value, "@")
		if at < 0 {
			continue
		}
		domain := row.Value[at+1:]
		kind := emailKind(domain)
		kinds[kind]++
		entry, ok := domains[domain]
		if !ok {
			entry = &DomainCount{Domain: domain, Kind: kind}
			domains[domain] = entry
		}
		entry.Count++
		if row.Verified {
			entry.Verified++
		}
	}

	// Rare domains are usually someone's own, so they're folded into Other
	// instead of pointing at individual members.
	other := DomainCount{Domain: "Other", Kind: "other"}
	for _, entry := range domains {
		if entry.Count < minDomainMembers {
			other.Count += entry.Count
			other.Verified += entry.Verified
			continue
		}
		result.Domains = append(result.Domains, *entry)
	}
	sort.Slice(result.Domains, func(i, j int) bool {
		if result.Domains[i].Count != result.Domains[j].Count {
			return result.Domains[i].Count > result.Domains[j].Count
		}
		return result.Domains[i].Domain < result.Domains[j].Domain
	})
	if other.Count > 0 {
		result.Domains = append(result.Domains, other)
	}

	for _, kind := range []string{"school", "team", "personal", "other"} {
		if kinds[kind] > 0 {
			result.Kinds = append(result.Kinds, CategoryCount{Label: kind, Count: int64(kinds[kind])})
		}
	}
	return result
}

var schoolDomains = map[string]bool{"ucsb.edu": true, "pipeline.sbcc.edu": true}

var personalDomains = map[string]bool{
	"gmail.com": true, "googlemail.com": true, "outlook.com": true, "hotmail.com": true, "live.com": true,
	"icloud.com": true, "me.com": true, "mac.com": true, "yahoo.com": true, "aol.com": true,
	"protonmail.com": true, "proton.me": true, "qq.com": true, "163.com": true,
}

func emailKind(domain string) string {
	switch {
	case schoolDomains[domain]:
		return "school"
	case domain == "gauchoracing.com":
		return "team"
	case personalDomains[domain]:
		return "personal"
	default:
		return "other"
	}
}
