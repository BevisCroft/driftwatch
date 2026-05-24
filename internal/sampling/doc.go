// Package sampling provides probabilistic and deterministic sampling
// strategies for controlling the volume of drift events processed by
// driftwatch in high-frequency environments.
//
// Two strategies are available:
//
//   - random: each event is independently admitted with probability Rate.
//   - deterministic: every 1/Rate-th event per service is admitted,
//     ensuring even distribution across time.
//
// A Rate of 0 blocks all events; a Rate of 1 admits all events.
package sampling
