package observability

import "testing"

func TestDatabaseOperationUsesBoundedLabels(t *testing.T) {
	tests := map[string]string{
		" SELECT * FROM users":   "SELECT",
		"insert into users":      "INSERT",
		"WITH recent AS (SELECT": "WITH",
		"VACUUM users":           "OTHER",
		"":                       "OTHER",
	}
	for sql, expected := range tests {
		if actual := databaseOperation(sql); actual != expected {
			t.Fatalf("databaseOperation(%q) = %q, want %q", sql, actual, expected)
		}
	}
}
