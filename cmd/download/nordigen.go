package download

import (
	"encoding/json"
	"fmt"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
	"time"
)

const NORDIGEN_ENDPOINT = "https://ob.nordigen.com/api/v2"

type NordigenDownload struct {
	ClientID     string
	ClientSecret string
}

const NORDIGEN_DAYS_AGO = 89

func (m *NordigenDownload) Add(root *cobra.Command) {
	ncmd := &cobra.Command{
		Use:               "nordigen <statefile.json[.gz]>",
		Short:             "Download data from Nordigen",
		Long:              "Download data from Nordigen",
		DisableAutoGenTag: true,
	}
	ncmd.Flags().StringVar(&m.ClientID, "Client ID", m.ClientID, "Client Secret for Nordigen")
	ncmd.Flags().StringVar(&m.ClientSecret, "Client Secret", m.ClientSecret, "Client ID for Nordigen")
	ncmd.Args = cobra.ExactArgs(1)
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		client, err := m.NewNordigenClient(args[0])
		if err != nil {
			return err
		}

		err = client.Init()
		if err != nil {
			return err
		}

		err = client.Sync()
		if err != nil {
			return err
		}

		return nil
	}
	root.AddCommand(ncmd)
}

//type NordigenAccountData struct {
//	Account     *StarlingAccount                `json:"account"`
//	Spaces      *Spaces                         `json:"spaces"`
//	Identifiers interface{}                     `json:"identifiers"`
//	FeedItems   map[string]map[string]*FeedItem `json:"feeditems"`
//}

type NordigenData struct {
	Token       *NordigenToken
	Accounts    map[string]*NordigenAccount
	Institution *Institution
	Requisition *Requisition
}

type NordigenToken struct {
	Access         string
	AccessExpires  time.Time
	Refresh        string
	RefreshExpires time.Time
}

type NordigenClient struct {
	client *http.Client
	file   string
	data   *NordigenData
	config *NordigenDownload
}

func (c *NordigenClient) Token() (*oauth2.Token, error) {

	type NordigenJSONToken struct {
		Access         string `json:"access,omitempty"`
		AccessExpires  int    `json:"access_expires,omitempty"`
		Refresh        string `json:"refresh,omitempty"`
		RefreshExpires int    `json:"refresh_expires,omitempty"`
	}

	var hc http.Client

	// TODO: AccessExpires is prior to RefreshExpires.. do Refresh first..
	if c.data.Token == nil || time.Now().After(c.data.Token.AccessExpires) {
		nToken := &NordigenJSONToken{}
		newUrl := fmt.Sprintf("%s/token/new/", NORDIGEN_ENDPOINT)
		params := map[string][]string{
			"secret_id":  []string{c.config.ClientID},
			"secret_key": []string{c.config.ClientSecret},
		}
		if err := postToURL(&hc, newUrl, url.Values(params), nToken); err != nil {
			return nil, fmt.Errorf("failed fetching new token: %w", err)
		}
		c.data.Token = &NordigenToken{
			Access:         nToken.Access,
			AccessExpires:  time.Now().Add(time.Second * time.Duration(nToken.AccessExpires)),
			Refresh:        nToken.Refresh,
			RefreshExpires: time.Now().Add(time.Second * time.Duration(nToken.RefreshExpires)),
		}
	} else if time.Now().After(c.data.Token.RefreshExpires) {
		nToken := &NordigenJSONToken{}
		refreshUrl := fmt.Sprintf("%s/token/refresh/", NORDIGEN_ENDPOINT)
		params := map[string][]string{
			"refresh": []string{c.data.Token.Refresh},
		}
		if err := postToURL(&hc, refreshUrl, url.Values(params), nToken); err != nil {
			return nil, fmt.Errorf("failed refreshing token: %v", err)
		}
		c.data.Token.Refresh = nToken.Refresh
		c.data.Token.RefreshExpires = time.Now().Add(time.Duration(nToken.RefreshExpires) * time.Second)
	}

	// Return token
	return &oauth2.Token{
		AccessToken:  c.data.Token.Access,
		RefreshToken: c.data.Token.Refresh,
		Expiry:       c.data.Token.AccessExpires,
	}, nil
}

func (m *NordigenDownload) NewNordigenClient(file string) (*NordigenClient, error) {
	var ndata NordigenData
	if err := utils.LoadFromFile(file, &ndata); err != nil {
		return nil, err
	}

	// New client
	nclient := &NordigenClient{
		client: &http.Client{},
		file:   file,
		data:   &ndata,
		config: m,
	}

	// Adjust HTTP transport to use the bearer tokens
	nclient.client.Transport = &oauth2.Transport{
		Source: nclient,
	}

	return nclient, nil
}

func (m *NordigenClient) Init() error {

	// If institution is set, we're ready to go
	if m.data.Institution != nil {
		return nil
	}

	// Find the institution
	fmt.Printf("Enter country: ")
	var country string
	_, err := fmt.Scanf("%s", &country)
	if err != nil {
		return err
	}

	vals, err := m.ListInstitutions(country)
	if err != nil {
		return err
	}
	for i, v := range vals {
		fmt.Printf("%3d %s\n", i, v.Name)
	}

	fmt.Printf("Enter institution: ")
	var institution int
	_, err = fmt.Scanf("%d", &institution)
	if err != nil {
		return err
	}

	if institution < 0 || institution >= len(vals) {
		return fmt.Errorf("no valid institution selected: %d", institution)
	}

	var i Institution = vals[institution]

	fmt.Printf("Choose institution: %v\n", i.Data)

	// Create the requisition
	var req = &Requisition{}
	var params = map[string][]string{
		"redirect":       []string{"http://localhost:8080/"},
		"institution_id": []string{i.Id},
	}
	if err := postToURL(m.client, fmt.Sprintf("%s/requisitions/", NORDIGEN_ENDPOINT), url.Values(params), req); err != nil {
		return fmt.Errorf("failed creating requisition: %w", err)
	}

	fmt.Printf("Captured Requisition: %v\n", req.Data)

	// Capture requisition
	m.data.Institution = &i
	m.data.Requisition = req

	// Save to file
	fmt.Printf("Saving to file: %v\n", m.file)
	if err := utils.SaveToFile(m.file, &m.data); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	// Exit early!
	fmt.Printf("Please visit link:\n%s\n", req.Link)
	os.Exit(0)

	return nil
}

//
// STRUCTURES
//

type Institution struct {
	Id   string
	Name string
	Data interface{}
}

func (i *Institution) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &i.Data); err != nil {
		return err
	}
	var err error
	if i.Id, err = utils.GetStringValue(i.Data, "id"); err != nil {
		return err
	}
	if i.Name, err = utils.GetStringValue(i.Data, "name"); err != nil {
		return err
	}
	return nil
}

func (i *Institution) MarshalJSON() ([]byte, error) {
	return json.Marshal(&i.Data)
}

type Requisition struct {
	Id       string
	Link     string
	Accounts []string
	Data     interface{}
}

func (r *Requisition) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &r.Data); err != nil {
		return err
	}
	var err error
	if r.Id, err = utils.GetStringValue(r.Data, "id"); err != nil {
		return err
	}
	if r.Link, err = utils.GetStringValue(r.Data, "link"); err != nil {
		return err
	}
	accountList, err := utils.GetListValue(r.Data, "accounts")
	if err != nil {
		return err
	}
	if r.Accounts, err = utils.ListToStringList(accountList); err != nil {
		return err
	}
	fmt.Printf("r.Accounts: %v, source: %v\n", r.Accounts, r.Data)
	return nil
}

func (r *Requisition) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Data)
}

type NordigenTransaction struct {
	Id   string
	Data interface{}
}

func (t *NordigenTransaction) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Data); err != nil {
		return err
	}
	var err error
	if t.Id, err = utils.GetStringValue(t.Data, "transactionId"); err != nil {
		return err
	}
	return nil
}

func (t *NordigenTransaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(&t.Data)
}

//
// Methods
//

func (m *NordigenClient) Sync() error {

	if err := m.RefreshRequisition(); err != nil {
		return err
	}

	if err := m.RefreshAccounts(); err != nil {
		return err
	}

	fmt.Printf("Saving to file: %v\n", m.file)
	if err := utils.SaveToFile(m.file, &m.data); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	return nil
}

//
// Getters
//

func (c *NordigenClient) ListInstitutions(country string) ([]Institution, error) {
	resp := make([]Institution, 0)
	if err := fetchFromURL(c.client, fmt.Sprintf("%s/institutions?payments_enabled=true&country=%s", NORDIGEN_ENDPOINT, country), &resp); err != nil {
		return nil, fmt.Errorf("failed list institutions: %w", err)
	}

	return resp, nil
}

func (c *NordigenClient) RefreshRequisition() error {
	if err := fetchFromURL(c.client, fmt.Sprintf("%s/requisitions/%s/", NORDIGEN_ENDPOINT, c.data.Requisition.Id), c.data.Requisition); err != nil {
		return fmt.Errorf("failed getting requisition: %w", err)
	}

	return nil
}

type NordigenAccount struct {
	Basic        interface{}
	Details      interface{}
	Transactions map[string]*NordigenTransaction
}

func (c *NordigenClient) RefreshAccounts() error {

	if c.data.Accounts == nil {
		c.data.Accounts = make(map[string]*NordigenAccount)
	}

	for _, accountId := range c.data.Requisition.Accounts {

		fmt.Printf("accountId: %s\n", accountId)

		acct, ok := c.data.Accounts[accountId]
		if !ok {
			acct = &NordigenAccount{}
			c.data.Accounts[accountId] = acct
		}

		// Basic info
		fmt.Printf("Fetching basic information\n")
		if err := fetchFromURL(c.client, fmt.Sprintf("%s/accounts/%s/", NORDIGEN_ENDPOINT, accountId), &acct.Basic); err != nil {
			return fmt.Errorf("failed getting account basic: %w", err)
		}

		// Detailed info
		fmt.Printf("Fetching detailed information\n")
		if err := fetchFromURL(c.client, fmt.Sprintf("%s/accounts/%s/details", NORDIGEN_ENDPOINT, accountId), &acct.Details); err != nil {
			return fmt.Errorf("failed getting account details: %w", err)
		}

		type NordigenTransactions struct {
			Booked  []NordigenTransaction `json:"booked"`
			Pending []NordigenTransaction `json:"pending"`
		}
		type NordigenTransactionsResponse struct {
			Transactions NordigenTransactions `json:"transactions"`
		}

		// Transactions
		// date_from and date_to are relevant here
		// Need to merge based on TransactionId from booked. Skip pending.
		fmt.Printf("Fetching transactions\n")
		resp := &NordigenTransactionsResponse{}
		if err := fetchFromURL(c.client, fmt.Sprintf("%s/accounts/%s/transactions", NORDIGEN_ENDPOINT, accountId), &resp); err != nil {
			return fmt.Errorf("failed getting transactions: %w", err)
		}

		// Update all booked transactions
		if acct.Transactions == nil {
			acct.Transactions = make(map[string]*NordigenTransaction)
		}
		inserted := 0
		updated := 0
		for i, b := range resp.Transactions.Booked {
			_, ok := acct.Transactions[b.Id]
			if ok {
				updated += 1
			} else {
				inserted += 1
			}
			acct.Transactions[b.Id] = &resp.Transactions.Booked[i]
		}

		fmt.Printf("Updated %d transaction(s), inserted %d transaction(s)\n", updated, inserted)
	}

	return nil
}
