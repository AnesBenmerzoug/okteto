package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okteto/okteto/cmd"
	"github.com/okteto/okteto/cmd/namespace"
	"github.com/okteto/okteto/pkg/config"
	"github.com/okteto/okteto/pkg/errors"
	"github.com/okteto/okteto/pkg/log"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/instrumentation/process"
	"k8s.io/apimachinery/pkg/util/runtime"

	// Load the different library for authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func init() {
	// override client-go error handlers to downgrade the "logging before flag.Parse" error
	errorHandlers := []func(error){
		func(e error) {
			log.Debugf("unhandled error: %s", e)
		},
	}

	runtime.ErrorHandlers = errorHandlers
}

func main() {
	log.Init(logrus.WarnLevel)
	log.Info("start")
	var logLevel string

	agent, span, ctx := getTracing()

	root := &cobra.Command{
		Use:           fmt.Sprintf("%s COMMAND [ARG...]", config.GetBinaryName()),
		Short:         "Manage cloud dev environments",
		SilenceErrors: true,
		PersistentPreRun: func(ccmd *cobra.Command, args []string) {
			log.SetLevel(logLevel)
			ccmd.SilenceUsage = true
		},
	}

	root.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "warn", "amount of information outputted (debug, info, warn, error)")
	root.AddCommand(cmd.Analytics())
	root.AddCommand(cmd.Version())
	root.AddCommand(cmd.Login())
	root.AddCommand(cmd.Create(ctx))
	root.AddCommand(cmd.Delete(ctx))
	root.AddCommand(namespace.Namespace(ctx))
	root.AddCommand(cmd.Init())
	root.AddCommand(cmd.Up())
	root.AddCommand(cmd.Down())
	root.AddCommand(cmd.Exec())
	root.AddCommand(cmd.Restart())

	err := root.Execute()

	if agent != nil {
		span.Finish()
		agent.Flush()
	}

	if err != nil {
		log.Fail(err.Error())
		if uErr, ok := err.(errors.UserError); ok {
			if len(uErr.Hint) > 0 {
				log.Hint("    %s", uErr.Hint)
			}
		}

		os.Exit(1)
	}
}

func getTracing() (*agent.Agent, opentracing.Span, context.Context) {
	ctx := context.Background()

	if apiKey, ok := os.LookupEnv("SCOPE_APIKEY"); ok {
		scope, err := agent.NewAgent(agent.WithApiKey(apiKey))
		if err != nil {
			log.Errorf("couldn't instantiate scope agent: %s", err)
			os.Exit(1)
		}

		operationName := filepath.Base(os.Args[0])
		if len(os.Args) > 1 {
			operationName = fmt.Sprintf("%s %s", filepath.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
		}

		span := process.StartSpan(operationName)
		span.SetTag("okteto.version", config.VersionString)
		ctx = opentracing.ContextWithSpan(ctx, span)
		log.Info("scope agent configured")
		return scope, span, ctx
	}

	return nil, nil, ctx
}
