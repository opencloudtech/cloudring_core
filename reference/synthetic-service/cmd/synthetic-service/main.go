package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opencloudtech/CloudRING/reference/synthetic-service/controller"
)

const mockProviderMode = "mock-provider"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("synthetic-service", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	mode := flags.String("mode", mockProviderMode, "reference provider mode")
	check := flags.Bool("check", false, "run a synthetic reconciliation and exit")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse arguments: %w", err)
	}
	if *mode != mockProviderMode {
		return fmt.Errorf("unsupported mode %q", *mode)
	}

	status, _, _, err := controller.Reconcile(controller.Claim{
		Name:       "image-self-check",
		Namespace:  "synthetic-reference",
		ProjectRef: "synthetic-project",
		Plan:       "standard",
	}, "provision", time.Unix(0, 0).UTC())
	if err != nil {
		return fmt.Errorf("synthetic reconciliation: %w", err)
	}
	if status.Phase != controller.PhaseReady {
		return fmt.Errorf("synthetic reconciliation returned phase %q", status.Phase)
	}

	if *check {
		fmt.Fprintln(output, "synthetic_service_check_ok mode=mock-provider phase=ready")
		return nil
	}

	fmt.Fprintln(output, "synthetic_service_started mode=mock-provider")
	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return nil
}
