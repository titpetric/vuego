package tour

import (
	"context"
	"flag"
	"log"

	"github.com/titpetric/platform"
	"github.com/titpetric/vuego/server/tour"
)

// Run executes the tour command with the given arguments.
func Run(args []string) error {
	fs := flag.NewFlagSet("tour", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "HTTP server address")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return Serve(*addr)
}

// Serve starts the tour server using the platform.
func Serve(addr string) error {
	log.Print("Serving embedded tour")

	opts := platform.NewOptions()
	opts.ServerAddr = addr

	p := platform.New(opts)
	p.Register(tour.NewModule())

	if err := p.Start(context.Background()); err != nil {
		return err
	}

	p.Wait()
	return nil
}

// Usage returns the usage string for the tour command.
func Usage() string {
	return `vuego tour [options]

Start the vuego tour server.

Options:
  -addr string    HTTP server address (default ":8080")

Examples:
  vuego tour                    # Start tour on default port
  vuego tour -addr :3000        # Start tour on port 3000`
}
