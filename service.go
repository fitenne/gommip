package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	*http.Server
}

func zapAccessLogger(output string) gin.HandlerFunc {
	c := zap.NewProductionConfig()
	c.DisableCaller = true
	c.DisableStacktrace = true
	c.Encoding = "console"
	c.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       zapcore.OmitKey,
		NameKey:        zapcore.OmitKey,
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	c.OutputPaths = []string{output}
	logger, err := c.Build()
	if err != nil {
		zap.S().Fatalln("failed to construct a logger:", err)
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}

			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}
			return fmt.Sprintf("%s %3d %s %13v  %15s %s %-7s %s  %s",
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.ErrorMessage,
			)
		},
		Output:    &writerWarpper{logger: logger},
		SkipPaths: []string{"/health"},
	})
}

func handleIp(ctx *gin.Context, rawIp string) {
	if rawIp == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "no address found",
		})
		return
	}

	ip := net.ParseIP(rawIp)
	if !isIpv4(ip) {
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "only ipv4 address is supported",
		})
		return
	}

	if ip.IsPrivate() {
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "you asked a private address",
		})
		return
	}

	if !ip.IsGlobalUnicast() {
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "non-public address is not allowed",
		})
		return
	}

	result, err := lookup(ip)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "something wrong with the server",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":    "OK",
		"result": result,
	})
}

func NewServer(config *Config) *Server {
	server := gin.New()

	server.Use(gin.RecoveryWithWriter(&writerWarpper{logger: zap.L()}))
	if config.Log.Access != "" {
		server.Use(zapAccessLogger(config.Log.Access))
	}

	server.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	server.GET("/:ip", func(ctx *gin.Context) {
		handleIp(ctx, ctx.Param("ip"))
	})

	server.GET("/", func(ctx *gin.Context) {
		ip := ""
		for _, h := range config.RealIpHeader {
			check := ctx.GetHeader(h)
			if check != "" {
				ip = check
				break
			}
		}
		if ip == "" {
			ip = ctx.RemoteIP()
		}

		handleIp(ctx, ip)
	})

	return &Server{
		Server: &http.Server{
			Addr:    config.Listen,
			Handler: server,
		},
	}
}

func isIpv4(ip net.IP) bool {
	return ip.To4() != nil
}
