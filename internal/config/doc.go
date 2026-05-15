// Package config provides loading and validation of the driftwatch daemon
// configuration file.
//
// Configuration is expressed as YAML and supports the following top-level
// fields:
//
//	manifest_dir   – directory that contains source manifest files (default: ./manifests)
//	poll_interval  – how often the daemon checks for drift (default: 30s)
//	log_level      – logging verbosity: debug, info, warn, error (default: info)
//	reporter:
//	  format       – output format: text or json (default: text)
//	  out_file     – path to write results; empty means stdout
package config
