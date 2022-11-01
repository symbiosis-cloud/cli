/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/symbiosis-go"
)

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func createListener() (l net.Listener, close func()) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l, func() {
		_ = l.Close()
	}
}

type LoginCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *LoginCommand) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Prompts you to login to Symbiosis and store login details",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {

			l, close := createListener()
			defer close()
			http.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				if !r.URL.Query().Has("token") {
					log.Printf("Could not retrieve token, please contact Symbiosis Support")
					close()
				}

				viper.Set("auth.method", "token")
				viper.Set("auth.token", r.URL.Query().Get("token"))
				viper.Set("auth.team_id", r.URL.Query().Get("teamId"))
				viper.WriteConfig()

				http.Redirect(rw, r, "https://app.symbiosis.host/oauth/done", http.StatusTemporaryRedirect)

				// make sure the connection isn't closed before we complete the redirect
				go func(ctx context.Context) {
					// TODO: find out why this is neccesary and fix it
					time.Sleep(time.Second * 2)
					for {
						select {
						case <-ctx.Done():
							close()
						case <-time.After(time.Second * 30):
							close()
						}
					}
				}(r.Context())

				log.Println("Successfully initialised")

			}))

			localUrl := fmt.Sprintf("http://localhost:%d", l.Addr().(*net.TCPAddr).Port)
			oauthUrl := fmt.Sprintf("https://app.symbiosis.host/oauth?redirect=%s", url.QueryEscape(localUrl))

			log.Printf("Opening your browser to %s\n", oauthUrl)

			openbrowser(oauthUrl)

			http.Serve(l, nil)

			return nil
		},
	}
}

func (c *LoginCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
