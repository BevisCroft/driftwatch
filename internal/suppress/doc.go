// Package suppress implements a suppression list that allows operators
// to silence known or accepted configuration drift for specific services
// and manifest fields.
//
// Suppression entries can be time-bounded via an optional ExpiresAt field.
// Expired entries are ignored during lookup and can be removed with Purge.
package suppress
