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

const STARLING_DAYS_AGO = 1

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
	Account     *StarlingAccount                 `json:"account"`
	Spaces      *Spaces                          `json:"spaces"`
	Identifiers interface{}                      `json:"identifiers"`
	FeedItems   map[string]map[string]*FeedItem  `json:"feeditems"`
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

func (m *StarlingClient) syncFeed(acct *StarlingAccountData, name string, categoryUid string) error {

	sinceTime := acct.Account.CreatedAt
	accountId := acct.Account.AccountUid

	if acct.FeedItems == nil {
		acct.FeedItems = make(map[string]map[string]*FeedItem)
	}
	if acct.FeedItems[categoryUid] == nil {
		acct.FeedItems[categoryUid] = make(map[string]*FeedItem)
	}
	feed := acct.FeedItems[categoryUid]

	// Find most recent time minus N days
	if !m.config.AllData && len(feed) > 0 {
		recentTime := ""
		for _, item := range feed {
			if recentTime == "" || item.UpdatedAt > recentTime {
				recentTime = item.UpdatedAt
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

	fmt.Printf("Downloading category %s (%s) since %s\n", name, categoryUid, sinceTime)

	// Structure to fetch
	type FeedItems struct {
		FeedItems []*FeedItem
	}
	items := &FeedItems{}

	// Fetch
	reqUrl := fmt.Sprintf("%s/api/v2/feed/account/%s/category/%s?changesSince=%s", STARLING_ENDPOINT, accountId, categoryUid, sinceTime)
	if err := fetchFromURL(m.client, reqUrl, &items); err != nil {
		return fmt.Errorf("failed getting feeditems: %w", err)
	}

	// Update items
	updatedItems := 0
	newItems := 0
	for _, item := range items.FeedItems {
		curr, ok := feed[item.FeedItemUid]
		if !ok {
			newItems++
		} else if fmt.Sprintf("%v", curr) != fmt.Sprintf("%v", item) {
			updatedItems++
		}
		feed[item.FeedItemUid] = item
	}
	fmt.Printf("Downloaded %d transactions(s), %d new, %d updated.\n",
		len(items.FeedItems), newItems, updatedItems)

	return nil
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
		_, ok := m.data.Accounts[v.AccountUid]
		if !ok {
			m.data.Accounts[v.AccountUid] = &StarlingAccountData{}
		}
		m.data.Accounts[v.AccountUid].Account = v
	}

	// Iterate the accounts
	for id, acct := range m.data.Accounts {

		// Download each account's identifier
		if err := fetchFromURL(m.client, STARLING_ENDPOINT+"/api/v2/accounts/"+id+"/identifiers", &acct.Identifiers); err != nil {
			return fmt.Errorf("failed getting account identifier: %w", err)
		}

		// Download default category feed
		if err := m.syncFeed(acct, "main account", acct.Account.DefaultCategory); err != nil {
			return err
		}

		// Fetch spaces
		if err := fetchFromURL(m.client, STARLING_ENDPOINT+"/api/v2/account/"+id+"/spaces", &acct.Spaces); err != nil {
			return fmt.Errorf("failed getting spaces: %w", err)
		}

		// Download  Saving Spaces
		for _, savingsSpace := range acct.Spaces.SavingsGoals {
			if err := m.syncFeed(acct, savingsSpace.Name, savingsSpace.SavingsGoalUid); err != nil {
				return err
			}
		}

		// Download Spending Spaces
		for _, spendingSpace := range acct.Spaces.SpendingSpaces {
			if err := m.syncFeed(acct, spendingSpace.Name, spendingSpace.SpaceUid); err != nil {
				return err
			}
		}
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

type Spaces struct {
	SavingsGoals   []SavingsGoal   `json:"savingsGoals"`
	SpendingSpaces []SpendingSpace `json:"spendingSpaces"`
}

type SavingsGoal struct {
	SavingsGoalUid string
	Name string
	Data interface{}
}

func (t *SavingsGoal) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Data); err != nil {
		return err
	}
	var err error
	if t.SavingsGoalUid, err = utils.GetStringValue(t.Data, "savingsGoalUid"); err != nil {
		return err
	}
	if t.Name, err = utils.GetStringValue(t.Data, "name"); err != nil {
		return err
	}
	return nil
}

func (t *SavingsGoal) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}

type SpendingSpace struct {
	SpaceUid string
	Name string
	Data interface{}
}

func (t *SpendingSpace) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Data); err != nil {
		return err
	}
	var err error
	if t.SpaceUid, err = utils.GetStringValue(t.Data, "spaceUid"); err != nil {
		return err
	}
	if t.Name, err = utils.GetStringValue(t.Data, "name"); err != nil {
		return err
	}
	return nil
}

func (t *SpendingSpace) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}

type FeedItem struct {
	FeedItemUid     string
	TransactionTime string
	UpdatedAt       string
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
	if t.UpdatedAt, err = utils.GetStringValue(t.Data, "updatedAt"); err != nil {
		return err
	}
	return nil
}

func (t *FeedItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}
