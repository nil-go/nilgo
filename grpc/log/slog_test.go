// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/grpclog"

	"github.com/nil-go/nilgo/grpc/log"
)

func TestSlogger(t *testing.T) {
	t.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
	t.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "1")

	buf := new(bytes.Buffer)
	logger := log.NewSlogger(slog.NewTextHandler(buf, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return attr
		},
	}))
	grpclog.SetLoggerV2(logger)
	grpclog.Info("info")
	grpclog.Infoln("info", " ", "ln")
	grpclog.Infof("info %s", "f")
	grpclog.Warning("warning")
	grpclog.Warningln("warning", " ", "ln")
	grpclog.Warningf("warning %s", "f")
	grpclog.Error("error")
	grpclog.Errorln("error", " ", "ln")
	grpclog.Errorf("error %s", "f")

	expected := `level=INFO source=/slog_test.go:35 msg=info
level=INFO source=/slog_test.go:36 msg="info ln"
level=INFO source=/slog_test.go:37 msg="info f"
level=WARN source=/slog_test.go:38 msg=warning
level=WARN source=/slog_test.go:39 msg="warning ln"
level=WARN source=/slog_test.go:40 msg="warning f"
level=ERROR source=/slog_test.go:41 msg=error
level=ERROR source=/slog_test.go:42 msg="error ln"
level=ERROR source=/slog_test.go:43 msg="error f"
`
	pwd, _ := os.Getwd()
	assert.Equal(t, expected, strings.ReplaceAll(buf.String(), pwd, ""))
}
