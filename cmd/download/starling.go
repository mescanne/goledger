package download

import (
	"encoding/json"
	"fmt"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

const STARLING_ENDPOINT = "https://api.starlingbank.com"

type StarlingDownload struct {
	PersonalAccessToken string
	AllData             bool
}

const STARLING_DAYS_AGO = 90

func (m *StarlingDownload) Add(root *cobra.Command) {
	ncmd := &cobra.Command{
		Use:               "starling <statefile.json[.gz]>",
		Short:             "Download data from Starling",
		Long:              "Download data from Starling",
		DisableAutoGenTag: true,
	}
	ncmd.Flags().StringVar(&m.PersonalAccessToken, "personalAccessToken", m.PersonalAccessToken, "Personal Access Token for Starling")
	downloadAllMsg := fmt.Sprintf("Download all data (default is %d days prior to most recent transaction)", STARLING_DAYS_AGO)
	ncmd.Flags().BoolVar(&m.AllData, "alldata", m.AllData, downloadAllMsg)
	ncmd.Args = cobra.ExactArgs(1)
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		client, err := m.NewStarlingClient(args[0])
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

type StarlingAccountData struct {
	Account     *StarlingAccount     `json:"account"`
	Identifiers interface{}          `json:"identifiers"`
	FeedItems   map[string]*FeedItem `json:"feeditems"`
}

type StarlingData struct {
	Accounts map[string]*StarlingAccountData `json:"accounts"`
}

type StarlingClient struct {
	client *http.Client
	file   string
	data   *StarlingData
	config *StarlingDownload
}

func (m *StarlingDownload) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: m.PersonalAccessToken,
	}, nil
}

func (m *StarlingDownload) NewStarlingClient(file string) (*StarlingClient, error) {
	var config StarlingData
	if err := utils.LoadFromFile(file, &config); err != nil {
		return nil, err
	}

	// New client
	return &StarlingClient{
		client: oauth2.NewClient(context.Background(), m),
		file:   file,
		data:   &config,
		config: m,
	}, nil
}

func (m *StarlingClient) Sync() error {

	type Accounts struct {
		Accounts []*StarlingAccount
	}
	accts := &Accounts{}

	// Get accounts
	if err := fetchFromURL(m.client, STARLING_ENDPOINT+"/api/v2/accounts", accts); err != nil {
		return fmt.Errorf("failed getting accounts: %s", err)
	}

	// Get Account Map
	if m.data.Accounts == nil {
		m.data.Accounts = make(map[string]*StarlingAccountData)
	}
	for _, v := range accts.Accounts {
		m.data.Accounts[v.AccountUid].Account = v
	}

	// Iterate the accounts
	for id, acct := range m.data.Accounts {

		// Download each account's identifier
		if err := fetchFromURL(m.client, STARLING_ENDPOINT+"/api/v2/accounts/"+id+"/identifiers", &acct.Identifiers); err != nil {
			return fmt.Errorf("failed getting account identifier: %w", err)
		}

		// Start-of-time for the account
		sinceTime := acct.Account.CreatedAt

		// Download each account
		if acct.FeedItems == nil {
			acct.FeedItems = make(map[string]*FeedItem)
		} else if !m.config.AllData {
			recentTime := ""
			for _, item := range acct.FeedItems {
				if recentTime == "" || item.TransactionTime > recentTime {
					recentTime = item.TransactionTime
				}
			}

			// Subtract DAYS AGO
			ttime, err := time.Parse(time.RFC3339, recentTime)
			if err != nil {
				return fmt.Errorf("failed parsing time '%s': %w", recentTime, err)
			}
			adjustedRecentTime := ttime.AddDate(0, 0, -1*STARLING_DAYS_AGO).Format(time.RFC3339)

			// Use this time if it's more recent
			if adjustedRecentTime > sinceTime {
				sinceTime = adjustedRecentTime
			}
		}
		fmt.Printf("fetching since %s...", sinceTime)

		// Structure to fetch
		type FeedItems struct {
			FeedItems []*FeedItem
		}
		items := &FeedItems{}

		// Fetch
		reqUrl := fmt.Sprintf("%s/api/v2/feed/account/%s/category/%s?changesSince=%s", STARLING_ENDPOINT, id, acct.Account.DefaultCategory, sinceTime)
		if err := fetchFromURL(m.client, reqUrl, &items); err != nil {
			return fmt.Errorf("failed getting feeditems: %w", err)
		}

		// Update items
		updatedItems := 0
		newItems := 0
		for _, item := range items.FeedItems {
			curr, ok := acct.FeedItems[item.FeedItemUid]
			if !ok {
				newItems++
			} else if fmt.Sprintf("%v", curr) != fmt.Sprintf("%v", item) {
				updatedItems++
			}
			acct.FeedItems[item.FeedItemUid] = item
		}
		fmt.Printf("fetched %d transactions, %d new, %d updated.\n",
			len(items.FeedItems), newItems, updatedItems)
	}

	if err := utils.SaveToFile(m.file, &m.data); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	return nil
}

//
// STRUCTURES
//

type StarlingAccount struct {
	AccountUid      string
	DefaultCategory string
	CreatedAt       string
	Data            interface{}
}

func (acc *StarlingAccount) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &acc.Data); err != nil {
		return err
	}
	var err error
	if acc.AccountUid, err = utils.GetStringValue(acc.Data, "accountUid"); err != nil {
		return err
	}
	if acc.DefaultCategory, err = utils.GetStringValue(acc.Data, "defaultCategory"); err != nil {
		return err
	}
	if acc.CreatedAt, err = utils.GetStringValue(acc.Data, "createdAt"); err != nil {
		return err
	}
	return nil
}

func (acc *StarlingAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(&acc.Data)
}

type FeedItem struct {
	FeedItemUid     string
	TransactionTime string
	Data            interface{}
}

func (t *FeedItem) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Data); err != nil {
		return err
	}
	var err error
	if t.FeedItemUid, err = utils.GetStringValue(t.Data, "feedItemUid"); err != nil {
		return err
	}
	if t.TransactionTime, err = utils.GetStringValue(t.Data, "transactionTime"); err != nil {
		return err
	}
	return nil
}

func (t *FeedItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}
