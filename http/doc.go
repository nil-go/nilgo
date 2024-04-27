// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package http provides opinionated production-ready HTTP server and client.
//
// The server is configured with the following features:
// - Health check service
// - Graceful shutdown
// - Panic recovery
// - Unix domain socket support
// - Open telemetry instrumentation
package http
