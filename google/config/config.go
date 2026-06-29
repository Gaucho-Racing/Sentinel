package config

import (
	"os"
	"time"
)

const Name = "sentinel-google"
const Version = "5.8.1"

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

// Kerbecs admin API — the gateway doubles as the service registry. The sentinel
// client resolves gateway-form paths (/api/core/...) to upstream URLs via its
// /admin-gw/resolve endpoint, which sits behind basic auth.
var KerbecsEndpoint = os.Getenv("KERBECS_ENDPOINT")
var KerbecsUser = os.Getenv("KERBECS_USER")
var KerbecsPassword = os.Getenv("KERBECS_PASSWORD")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

// InternalBootstrapSecret is the shared secret this service uses at
// startup to exchange for its pre-seeded bearer JWT from core. Must
// match core's INTERNAL_BOOTSTRAP_SECRET.
var InternalBootstrapSecret = os.Getenv("INTERNAL_BOOTSTRAP_SECRET")

// InternalServiceName is the SA name on core that this service exchanges
// the bootstrap secret for. Must match a value in
// core/jobs/init.go::InternalServiceAccountNames.
const InternalServiceName = "sentinel-google"

// GoogleServiceAccount is the JSON key for the service account used to call the
// Admin SDK Directory API. It must have domain-wide delegation granted for the
// admin.directory.group.member scope. When empty, Google sync is disabled and
// the service runs as a no-op (binding CRUD still works).
var GoogleServiceAccount = os.Getenv("GOOGLE_SERVICE_ACCOUNT")

// GoogleAdminSubject is the super-admin user the service account impersonates
// (domain-wide delegation requires a subject). Required when GoogleServiceAccount
// is set.
var GoogleAdminSubject = os.Getenv("GOOGLE_ADMIN_SUBJECT")

// GoogleSyncInterval is how often the reconcile cron fires. Sync is one-way and
// not latency-critical (mailing-list membership), so a periodic full sweep —
// rather than per-change event triggers from core — keeps the dependency graph
// clean while landing changes within the interval.
const GoogleSyncInterval = 5 * time.Minute

// GoogleSyncMaxRemovals caps how many members a single per-group reconcile may
// delete. If a run wants to remove more than this, it skips the removals for
// that group and logs loudly — a guard against draining a group when core
// returns an empty/partial member set (e.g. mid-outage).
const GoogleSyncMaxRemovals = 100

func IsProduction() bool {
	return Env == "PROD"
}

// GoogleSyncEnabled reports whether the service has the credentials needed to
// talk to Google. When false, the reconcile engine no-ops.
func GoogleSyncEnabled() bool {
	return GoogleServiceAccount != "" && GoogleAdminSubject != ""
}
