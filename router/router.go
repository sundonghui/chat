package router

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sundonghui/chat/api"
	"github.com/sundonghui/chat/api/stream"
	"github.com/sundonghui/chat/config"
	"github.com/sundonghui/chat/database"
	gerror "github.com/sundonghui/chat/error"
	"github.com/sundonghui/chat/location"
	"github.com/sundonghui/chat/model"
)

var tokenRegexp = regexp.MustCompile("token=[^&]+")

type onlyImageFS struct {
	inner http.FileSystem
}

func Create(db *database.GormDatabase, vInfo *model.VersionInfo, conf *config.Configuration) (*gin.Engine, func()) {
	g := gin.New()

	g.RemoteIPHeaders = []string{"X-Forwarded-For"}
	g.SetTrustedProxies(conf.Server.TrustedProxies)
	g.ForwardedByClientIP = true

	g.Use(func(ctx *gin.Context) {
		// Map sockets "@" to 127.0.0.1, because gin-gonic can only trust IPs.
		if ctx.Request.RemoteAddr == "@" {
			ctx.Request.RemoteAddr = "127.0.0.1:65535"
		}
	})

	g.Use(gin.LoggerWithFormatter(logFormatter), gin.Recovery(), gerror.Handler(), location.Default())
	g.NoRoute(gerror.NotFound())

	if conf.Server.SSL.Enabled != nil && conf.Server.SSL.RedirectToHTTPS != nil && *conf.Server.SSL.Enabled && *conf.Server.SSL.RedirectToHTTPS {
		g.Use(func(ctx *gin.Context) {
			if ctx.Request.TLS != nil {
				ctx.Next()
				return
			}
			if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
				ctx.Data(http.StatusBadRequest, "text/plain; charset=utf-8", []byte("Use HTTPS"))
				ctx.Abort()
				return
			}
			host := ctx.Request.Host
			if idx := strings.LastIndex(host, ":"); idx != -1 {
				host = host[:idx]
			}
			if conf.Server.SSL.Port != 443 {
				host = fmt.Sprintf("%s:%d", host, conf.Server.SSL.Port)
			}
			ctx.Redirect(http.StatusFound, fmt.Sprintf("https://%s%s", host, ctx.Request.RequestURI))
			ctx.Abort()
		})
	}

	streamHandler := stream.New(
		time.Duration(conf.Server.Stream.PingPeriodSeconds)*time.Second, 15*time.Second, conf.Server.Stream.AllowedOrigins)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			connectedTokens := streamHandler.CollectConnectedClientTokens()
			now := time.Now()
			db.UpdateClientTokensLastUsed(connectedTokens, &now)
		}
	}()
	return g, streamHandler.Close
}

func logFormatter(param gin.LogFormatterParams) string {
	if (param.ClientIP == "127.0.0.1" || param.ClientIP == "::1") && param.Path == "/health" {
		return ""
	}

	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency - param.Latency%time.Second
	}
	path := tokenRegexp.ReplaceAllString(param.Path, "token=[masked]")
	return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format(time.RFC3339),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		path,
		param.ErrorMessage,
	)
}

func (fs *onlyImageFS) Open(name string) (http.File, error) {
	ext := filepath.Ext(name)
	if !api.ValidApplicationImageExt(ext) {
		return nil, fmt.Errorf("invalid file")
	}
	return fs.inner.Open(name)
}
