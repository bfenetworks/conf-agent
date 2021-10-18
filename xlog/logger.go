// Copyright (c) 2021 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xlog

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/baidu/conf-agent/config"
	"github.com/baidu/go-lib/log"
	"github.com/baidu/go-lib/log/log4go"
)

type Logger interface {
	Debug(arg0 interface{}, args ...interface{})
	Info(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{}) error
}

func Init(c *config.LoggerConfig) error {
	log4go.SetLogBufferLength(10000)
	log4go.SetLogWithBlocking(false)
	log4go.SetLogFormat(c.Format)

	logWriter, err := log.Create(c.LogName, c.LogLevel, c.LogDir, c.StdOut, c.RotateWhen, c.BackupCount)
	if err != nil {
		return err
	}

	Default = logWriter
	return nil
}

var Default Logger = &fakeLogger{}

type fakeLogger struct{}

func (fl *fakeLogger) Debug(arg0 interface{}, args ...interface{}) {
	if s, ok := arg0.(string); ok {
		fmt.Printf(s, args)
	} else {
		fmt.Print(arg0, args)
	}
	fmt.Println()
}

func (fl *fakeLogger) Info(arg0 interface{}, args ...interface{}) {
	if s, ok := arg0.(string); ok {
		fmt.Printf(s, args)
	} else {
		fmt.Print(arg0, args)
	}
	fmt.Println()
}

func (fl *fakeLogger) Error(arg0 interface{}, args ...interface{}) error {
	if s, ok := arg0.(string); ok {
		fmt.Printf(s, args)
	} else {
		fmt.Print(arg0, args)
	}
	fmt.Println()
	return nil
}

var ran = rand.NewSource(time.Now().Unix())

var RandomLogID = func() string {
	return fmt.Sprintf("%d_%03d", time.Now().UnixNano(), ran.Int63()%1000)
}

type logCtx string

var logCtxKey logCtx = "log_ctx"

type LogContext struct {
	LogID        string
	ReloaderName string
}

func NewContext(ctx context.Context, moduleName string) context.Context {
	if ctx.Value(logCtxKey) != nil {
		return ctx
	}

	return context.WithValue(ctx, logCtxKey, &LogContext{
		LogID:        RandomLogID(),
		ReloaderName: moduleName,
	})
}

func getLogContext(ctx context.Context) *LogContext {
	id := ctx.Value(logCtxKey)
	if id == nil {
		return &LogContext{}
	}

	return id.(*LogContext)
}

func ErrLogFormat(ctx context.Context, topic string, err error) string {
	logCtx := getLogContext(ctx)
	return fmt.Sprintf("[%s] module[%16s] topic[%s] err[%v]", logCtx.LogID, logCtx.ReloaderName, topic, err)
}

func InfoLogFormat(ctx context.Context, topic string, ss ...interface{}) string {
	logCtx := getLogContext(ctx)
	if len(ss) > 0 {
		return fmt.Sprintf("[%s] module[%16s] topic[%s] info[%s]", logCtx.LogID, logCtx.ReloaderName, topic, fmt.Sprint(ss...))
	}

	return fmt.Sprintf("[%s] module[%16s] topic[%s]", logCtx.LogID, logCtx.ReloaderName, topic)
}
