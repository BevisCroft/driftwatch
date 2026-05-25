// Package ownership maintains a registry that maps service names to their
// responsible teams and contact information.
//
// During a drift cycle the registry can be queried to annotate drift results
// with the owning team, making it straightforward to route alerts to the
// correct on-call channel.
//
// Usage:
//
//	reg := ownership.New()
//	_ = reg.Set(ownership.Entry{
//		Service:  "api-gateway",
//		Team:     "platform",
//		Contacts: []string{"platform@example.com"},
//	})
//
//	entry, ok := reg.Get("api-gateway")
package ownership
