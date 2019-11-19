package download

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type MonzoDownload struct {
	ClientID     string
	ClientSecret string
	Port         int
}

const DAYS_AGO = 89

func (m *MonzoDownload) Add(root *cobra.Command) {
	if m.Port == 0 {
		m.Port = 12345
	}

	ncmd := &cobra.Command{
		Use:               "monzo <statefile.json[.gz]>",
		Short:             "Download data from Monzo",
		Long:              "Download data from Monzo",
		DisableAutoGenTag: true,
	}
	ncmd.Flags().StringVar(&m.ClientID, "clientid", m.ClientID, "Client ID for application")
	ncmd.Flags().StringVar(&m.ClientSecret, "clientsecret", m.ClientSecret, "Client Secret for application")
	ncmd.Flags().IntVarP(&m.Port, "port", "p", m.Port, "Port for OAuth2 loop-back on localhost (must match app config)")
	ncmd.Args = cobra.ExactArgs(1)
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		client, err := m.NewMonzoClient(args[0])
		if err != nil {
			return err
		}

		if err = client.Sync(); err != nil {
			return err
		}

		return nil
	}
	root.AddCommand(ncmd)
}

type MonzoData struct {
	Token        *oauth2.Token           `json:"token"`
	Transactions map[string]*Transaction `json:"transactions"`
	Accounts     map[string]*Account     `json:"accounts"`
}

type MonzoClient struct {
	client *http.Client
	file   string
	data   *MonzoData
	config *MonzoDownload
}

func (m *MonzoClient) Sync() error {

	// Verify authentication
	if err := m.VerifyAuthenticated(); err != nil {
		return err
	}

	type Accounts struct {
		Accounts []*Account
	}
	accts := &Accounts{}

	// Get accounts
	if err := fetchFromURL(m.client, "https://api.monzo.com/accounts", accts); err != nil {
		return fmt.Errorf("failed getting accounts: %s", err)
	}

	if m.data.Accounts == nil {
		m.data.Accounts = make(map[string]*Account)
	}
	for _, v := range accts.Accounts {
		m.data.Accounts[v.ID] = v
	}

	startTrans := len(m.data.Transactions)
	for _, acct := range m.data.Accounts {
		if err := m.GetAllTransactions(acct.ID); err != nil {
			return fmt.Errorf("failed getting transactions: %v", err)
		}
	}
	newTrans := len(m.data.Transactions) - startTrans

	fmt.Printf("Added %d new transactions (total %d)\n", newTrans, len(m.data.Transactions))

	if err := utils.SaveToFile(m.file, &m.data); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	return nil
}

func (m *MonzoDownload) NewMonzoClient(file string) (*MonzoClient, error) {
	var config MonzoData
	if err := utils.LoadFromFile(file, &config); err != nil {
		return nil, err
	}

	// Configure Monzo endpoints
	conf := oauth2.Config{
		ClientID:     m.ClientID,
		ClientSecret: m.ClientSecret,

		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://auth.monzo.com/",
			TokenURL:  "https://api.monzo.com/oauth2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/", m.Port),
		Scopes:      []string{},
	}
	oauth2.RegisterBrokenAuthHeaderProvider(conf.Endpoint.TokenURL)

	ctx := context.Background()
	if config.Token == nil {
		fmt.Printf("Getting a brand new authorization token..\n")

		rand.Seed(time.Now().UnixNano())
		state := fmt.Sprintf("%d", rand.Int())

		url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
		fmt.Printf("Visit the URL for the auth dialog: %v\n", url)

		code, istate, err := CaptureCode(m.Port)
		if err != nil {
			return nil, fmt.Errorf("Failed capturing code: %s", err)
		}
		if istate != state {
			return nil, fmt.Errorf("State %s doesn't match %s!", state, istate)
		}

		tok, err := conf.Exchange(ctx, code)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Done getting token.\n")
		config.Token = tok

		if err := utils.SaveToFile(file, &config); err != nil {
			return nil, err
		}

		// Success, but we can't do anything until authorised on app as well.
		fmt.Printf("Approve in Monzo app -- retry after authorised\n")
		return nil, fmt.Errorf("pending authorisation")
	}

	// Refresh token if needed
	if config.Token.Expiry.Before(time.Now()) {
		src := conf.TokenSource(ctx, config.Token)
		newToken, err := src.Token() // Renew the token
		if err != nil {
			return nil, fmt.Errorf("failed renewing token: %w", err)
		}

		// Capture new token
		if newToken.AccessToken != config.Token.AccessToken {
			fmt.Printf("Refreshed access token.\n")

			config.Token = newToken

			// Save it.
			if err := utils.SaveToFile(file, &config); err != nil {
				return nil, fmt.Errorf("failed saving: %w", err)
			}
		}
	}

	// New client
	return &MonzoClient{
		client: conf.Client(ctx, config.Token),
		file:   file,
		data:   &config,
		config: m,
	}, nil
}

//
// STRUCTURES
//

type Account struct {
	ID   string
	Data interface{}
}

func (acc *Account) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &acc.Data); err != nil {
		return err
	}
	var err error
	if acc.ID, err = utils.GetStringValue(acc.Data, "id"); err != nil {
		return err
	}
	return nil
}

func (acc *Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(&acc.Data)
}

type Transaction struct {
	ID      string
	Created string
	Data    interface{}
}

func (t *Transaction) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Data); err != nil {
		return err
	}
	var err error
	if t.ID, err = utils.GetStringValue(t.Data, "id"); err != nil {
		return err
	}
	if t.Created, err = utils.GetStringValue(t.Data, "created"); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}

//
// FETCHERS
//

func (m *MonzoClient) VerifyAuthenticated() error {
	type WhoAmIResponse struct {
		Authenticated bool   `json:"authenticated"`
		ClientID      string `json:"client_id"`
		UserID        string `json:"user_id"`
	}

	var resp *WhoAmIResponse = &WhoAmIResponse{}

	if err := fetchFromURL(m.client, "https://api.monzo.com/ping/whoami", resp); err != nil {
		return fmt.Errorf("failed whoami: %v", err)
	}

	if !resp.Authenticated {
		return fmt.Errorf("failed whoami: not authenticated")
	}

	return nil
}

const LIMIT = 100

func (m *MonzoClient) GetAllTransactions(accountID string) error {
	if m.data.Transactions == nil {
		m.data.Transactions = make(map[string]*Transaction)
	}

	var since = ""
	for {
		fmt.Printf("Downloading transactions for account %s... ", accountID)

		trans, err := m.getTransactions(accountID, since, LIMIT)
		if since == "" && errors.Is(err, forbidden) {
			fmt.Printf("only fetching last %d days... ", DAYS_AGO)
			since = time.Now().Add(-1 * time.Hour * 24 * DAYS_AGO).Format(time.RFC3339)
			trans, err = m.getTransactions(accountID, since, LIMIT)
		}

		if err != nil {
			fmt.Printf("\n")
			return err
		}

		fmt.Printf("downloaded %d\n", len(trans.Trans))

		// Received none -- really done!
		if len(trans.Trans) == 0 {
			break
		}

		// Map new transactions in
		var created time.Time
		for _, t := range trans.Trans {
			m.data.Transactions[t.ID] = t
			thisTime, err := time.Parse(time.RFC3339, t.Created)
			if err != nil {
				return fmt.Errorf("failed parsing timestamp '%s': %w", t.Created, err)
			}
			if thisTime.After(created) {
				created = thisTime
			}
		}

		// If we didn't retrieve the full limit, we're done
		if len(trans.Trans) < LIMIT {
			break
		}

		// Use the the time of a minute before the most recent transaction --
		// Bug in Monzo API, unfortunately.
		since = created.Add(-1 * time.Second * 60).Format(time.RFC3339)

	}

	return nil
}

type Transactions struct {
	Trans []*Transaction `json:"transactions"`
}

func (m *MonzoClient) getTransactions(accountID string, since string, limit uint) (*Transactions, error) {
	vals := make(url.Values)
	vals.Add("since", since)
	vals.Add("limit", fmt.Sprintf("%d", limit))
	vals.Add("account_id", accountID)
	vals.Add("expand[]", "merchant")
	url := fmt.Sprintf("https://api.monzo.com/transactions?%s", vals.Encode())
	resp := &Transactions{}
	if err := fetchFromURL(m.client, url, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
