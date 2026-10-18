package service

import "strings"

var usStateNames = map[string]string{
	"AL": "Alabama", "AK": "Alaska", "AZ": "Arizona", "AR": "Arkansas", "CA": "California",
	"CO": "Colorado", "CT": "Connecticut", "DE": "Delaware", "DC": "District of Columbia", "FL": "Florida",
	"GA": "Georgia", "HI": "Hawaii", "ID": "Idaho", "IL": "Illinois", "IN": "Indiana",
	"IA": "Iowa", "KS": "Kansas", "KY": "Kentucky", "LA": "Louisiana", "ME": "Maine",
	"MD": "Maryland", "MA": "Massachusetts", "MI": "Michigan", "MN": "Minnesota", "MS": "Mississippi",
	"MO": "Missouri", "MT": "Montana", "NE": "Nebraska", "NV": "Nevada", "NH": "New Hampshire",
	"NJ": "New Jersey", "NM": "New Mexico", "NY": "New York", "NC": "North Carolina", "ND": "North Dakota",
	"OH": "Ohio", "OK": "Oklahoma", "OR": "Oregon", "PA": "Pennsylvania", "RI": "Rhode Island",
	"SC": "South Carolina", "SD": "South Dakota", "TN": "Tennessee", "TX": "Texas", "UT": "Utah",
	"VT": "Vermont", "VA": "Virginia", "WA": "Washington", "WV": "West Virginia", "WI": "Wisconsin",
	"WY": "Wyoming",
}

var usStateCodes = func() map[string]string {
	codes := make(map[string]string, len(usStateNames)+3)
	for code, name := range usStateNames {
		codes[strings.ToLower(name)] = code
	}
	codes["washington state"] = "WA"
	codes["washington d.c."] = "DC"
	codes["washington, d.c."] = "DC"
	return codes
}()

// usState maps libphonenumber's US geocoding ("California", "Washington
// State", or "San Francisco, CA") to a state code and name.
func usState(geo string) (string, string, bool) {
	geo = strings.TrimSpace(geo)
	if code, ok := usStateCodes[strings.ToLower(geo)]; ok {
		return code, usStateNames[code], true
	}
	if i := strings.LastIndex(geo, ", "); i >= 0 {
		code := strings.ToUpper(geo[i+2:])
		if name, ok := usStateNames[code]; ok {
			return code, name, true
		}
	}
	return "", "", false
}
