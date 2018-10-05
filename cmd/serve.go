package cmd

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/percona/pmm-agent/errs"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		errs := &errs.Safe{}
		wg := &sync.WaitGroup{}

		// Create context which cancels on termination signals.
		ctx := contextWithCancelOnSignal(syscall.SIGTERM, syscall.SIGINT)

		// Start all services.
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Serve(ctx); err != nil {
				errs.Add(err)
			}
		}()

		// Wait for all services to finish.
		wg.Wait()

		return errs.Err()
	},
}

func init() {
	server.Flags(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func contextWithCancelOnSignal(sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

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
