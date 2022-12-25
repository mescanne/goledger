package web

import (
	"embed"
	"fmt"
	"github.com/mescanne/goledger/cmd/accounts"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/currencies"
	"github.com/mescanne/goledger/cmd/register"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

type WebConfig struct {
	Port  int
	Host  string
	Local string
}

type WebApp struct {
	app.App
	Report   *reports.TransactionReport
	Register *register.RegisterReport
}

//go:embed static/*.js static/*.css index.html
var webData embed.FS

const web_long = `Web reporting

... stuff here.
`

const DEFAULT_PORT = 8080

func Add(root *cobra.Command, webcfg *WebConfig, app *WebApp) {
	ncmd := &cobra.Command{
		Use:               "web",
		Short:             "Run as a website",
		Long:              web_long,
		DisableAutoGenTag: true,
	}

	if webcfg.Port == 0 {
		webcfg.Port = DEFAULT_PORT
	}

	ncmd.Flags().IntVar(&webcfg.Port, "port", webcfg.Port, "port for webserver")
	ncmd.Flags().StringVar(&webcfg.Host, "host", webcfg.Host, "host for serving")
	ncmd.Flags().StringVar(&webcfg.Local, "local", webcfg.Local, "Serve local files from directory rather than compiled")
	ncmd.MarkFlagDirname("local")
	ncmd.Args = cobra.NoArgs

	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return webcfg.run(app, cmd, args)
	}

	root.AddCommand(ncmd)
}

func (webcfg *WebConfig) run(app *WebApp, cmd *cobra.Command, args []string) error {
	fmt.Printf("Web server on port %d\n", webcfg.Port)
	http.Handle("/goledger", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		webcfg.handle(app, resp, req)
	}))

	if webcfg.Local != "" {
		http.Handle("/", http.FileServer(http.Dir(webcfg.Local)))
	} else {
		http.Handle("/", http.FileServer(http.FS(webData)))
	}

	return http.ListenAndServe(fmt.Sprintf("%s:%d", webcfg.Host, webcfg.Port), nil)
}

func (webcfg *WebConfig) handle(app *WebApp, resp http.ResponseWriter, req *http.Request) {
	args := []string{}
	v := req.URL.Query()
	u_args, ok := v["arg"]
	if ok {
		args = u_args
	}

	// NOTE: This should be JSON if it's rendering in JSON!
	resp.Header()["Content-Type"] = []string{"text/html"}
	resp.WriteHeader(http.StatusOK)
	err := app.Execute(args, resp)
	if err != nil {
		fmt.Fprintf(resp, "Error: %v\n", err)
	}
}

// Execute command line program
func (app *WebApp) Execute(args []string, out io.Writer) error {

	// Set the output
	app.Output = out
	app.Colour = false

	// Load core application
	appCmd := app.LoadCommand()

	// Add sub-commands
	// As they are webified
	accounts.Add(appCmd, &app.App)
	reports.Add(appCmd, &app.App, app.Report)
	register.Add(appCmd, &app.App, app.Register)
	currencies.Add(appCmd, &app.App)

	// Set the arguments
	appCmd.SetArgs(args)

	// Run core app
	return appCmd.Execute()
}
