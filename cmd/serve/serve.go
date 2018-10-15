package serve

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/errs"
)

// New returns `serve` command.
func New(ctx context.Context, app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve accepts incoming connections and starts supervisor api.",
		RunE: func(cmd *cobra.Command, args []string) error {
			errs := &errs.Safe{}
			wg := &sync.WaitGroup{}

			// Create context which cancels on termination signals.
			ctx := contextWithCancelOnSignal(ctx, syscall.SIGTERM, syscall.SIGINT)

			// Start all services.
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := app.Server.Serve(ctx); err != nil {
					errs.Add(err)
				}
			}()

			// Wait for all services to finish.
			wg.Wait()

			return errs.Err()
		},
	}

	app.Server.Flags(cmd)
	return cmd
}

func contextWithCancelOnSignal(ctx context.Context, sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sig...)
	go func() {
		s := <-signals
		signal.Stop(signals)
		log.Printf("Got '%s (%d)' signal, shutting down...", s, s)
		cancel()
	}()

	return ctx
}
