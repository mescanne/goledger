package web

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/accounts"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/currencies"
	"github.com/mescanne/goledger/cmd/register"
	"github.com/mescanne/goledger/cmd/reports"
	"io"
	//"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"net/http"
)

type WebConfig struct {
	Port  int
	Host  string
	Local bool
}

type WebApp struct {
	app.App
	Report   *reports.TransactionReport
	Register *register.RegisterReport
}

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
	ncmd.Flags().BoolVar(&webcfg.Local, "local", webcfg.Local, "Serve local files from /assets rather than compiled")
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
	http.Handle("/", http.FileServer(FS(webcfg.Local)))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", webcfg.Host, webcfg.Port), nil)
}

func (webcfg *WebConfig) handle(app *WebApp, resp http.ResponseWriter, req *http.Request) {
	//req.Method, req.URL, req.Header, req.Body
	//// , req.Form, req.PostForm
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//resp.Header

	// Need:
	// - Simple way to send in through forms, multipart forms, or JSON POST configuration parameters.
	// - Simple way to send in command line arguments (?) through GET parameters perhaps?
	// - Mechanism to send plain text back (normal) or JSON or HTML depending on the content type of webcfg.

	args := []string{}

	v := req.URL.Query()
	u_args, ok := v["arg"]
	if ok {
		args = u_args
	}

	resp.WriteHeader(http.StatusOK)
	err := app.Execute(args, resp)
	if err != nil {
		fmt.Printf("got error: %v\b", err)
	}

	//fmt.Fprintf(resp, "Hello world")

	// possibly can re-execute the command as-is multiple times.

	// Then we need:
	// A suite of files embedded that can be used to drive the command line from the web.
	// This allows configuration of what returns, getting JSON, building webcfg, etc.

	// Order to start with:
	// - A simple HTML form, static page, that can pass in command line arguments?
	// - Or a JSON form? Maybe just command line arguments is enough.

	// Order to start:
	// Develop a gen-embed approach with some HTML. Get that working.
	// Custom the HTML to have a command line input at the top.
	// Send in a request to the web to process it.
}

//
// Execute command line program
//
func (app *WebApp) Execute(args []string, out io.Writer) error {

	// Load core application
	appCmd := app.LoadCommand()

	// Add sub-commands
	// As they are webified
	accounts.Add(appCmd, &app.App)
	reports.Add(appCmd, &app.App, app.Report)
	register.Add(appCmd, &app.App, app.Register)
	currencies.Add(appCmd, &app.App)

	// Set the
	appCmd.SetOut(out)
	appCmd.SetArgs(args)

	// Run core app
	return appCmd.Execute()
}
