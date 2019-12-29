package auctionhouse

import (
	"auctionauth"
	"database/sql"
	"net/http"
	"time"
)

const regionString = "{region}"
const localeString = "{locale}"
const tokenString = "{token}"

// Database Types
// ---------------------Database Types--------------------

type DBInfo struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBname    string `json:"dbname"`
	connected bool
}
type aucDB struct {
	db     *sql.DB
	auth   OAuthResponse
	client http.Client
}

// ---------------------OAuth Types--------------------

type Credentials struct {
	Client string `json:"cid"`
	Secret string `json:"csecret"`
}
type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ---------------------Token Types--------------------
type Token struct {
	Client        string
	Secret        string
	token         string `json:"access_token"`
	LastModified  int
	LastUpdated   time.Time
	tokenURL      string
	checkTokenURL string
}

// ---------------------Realm Types--------------------

type Realms struct {
	Realms []Realm `json:"realms"`
}
type Realm struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	URL  string
}

type Region struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}
type RealmType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
type RealmData struct {
	ID           int       `json:"id"`
	Region       Region    `json:"region"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Locale       string    `json:"locale"`
	Timezone     string    `json:"timezone"`
	Type         RealmType `json:"type"`
	IsTournament bool      `json:"is_tournament"`
	Slug         string    `json:"slug"`
}
type Daemon struct {
	Token  auctionauth.Token
	ID     int
	API    map[string]string
	Region string
	Locale string
	Realms []Realm
}

// ---------------------Auction Types--------------------

type AuctionHandler struct {
	Realm       Realm
	Auctions    chan Auction
	LastChecked time.Time
	Insert      string
	Token       auctionauth.Token
	URL         string
	db          *sql.DB
	dbInfo      DBInfo
}
type Auctions struct {
	Auctions []Auction `json:"auctions"`
}

type Auction struct {
	AuctionID int    `json:"auc"`
	Item      int    `json:"item"`
	ORealm    string `json:"ownerRealm"`
	Bid       int    `json:"bid"`
	Buyout    int    `json:"buyout"`
	Quantity  int    `json:"Quantity"`
	TimeLeft  string `json:"timeLeft"`
}

type Files struct {
	Files []AuctionURL `json:"files"`
}

type AuctionURL struct {
	URL string `json:"url"`
}

type AuctionMeta struct {
	URL          string `json:"url"`
	LastModified int    `json:"lastModified"`
}

// ---------------------Item Types--------------------
type ItemManager struct {
	api            map[string]string
	toAdd          chan Item
	toQueryAPI     chan Item
	Items          map[int]int
	db             *sql.DB
	dbInfo         DBInfo
	InsertString   *sql.Stmt
	QueryStatement *sql.Stmt
}
type Modifiers struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}
type Mods struct {
	Modifiers []Modifiers
}
type Pet struct {
	Species   int `json:"petSpecies"`
	BreedID   int `json:"petBreedId"`
	Level     int `json:"petLevel"`
	QualityID int `json:"petQualityId"`
}
type Quality struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
type Media struct {
	Key Key `json:"key"`
	ID  int `json:"id"`
}
type Key struct {
	HREF string `json:"href"`
}
type ItemClass struct {
	Key  Key    `json:"key"`
	Name string `json:"name"`
	id   string `json:"id"`
}
type ItemSubClass struct {
	Key  Key    `json:"key"`
	Name string `json:"name"`
	id   string `json:"id"`
}
type InventoryType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
type Item struct {
	ItemAPICode   int
	IconAPICode   int
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Quality       Quality       `json:"quality"`
	Level         int           `json:"level"`
	ReqLevel      int           `json:"required_level"`
	Media         Media         `json:"media"`
	Class         ItemClass     `json:"item_class"`
	Subclass      ItemSubClass  `json:"item_subclass"`
	InventoryType InventoryType `json:"inventory_type"`
	PurchasePrice int           `json:"purchase_price"`
	MaxCount      int           `json:"max_count"`
	IsEquippable  bool          `json:"is_equippable"`
	IsStackable   bool          `json:"is_stackable"`
	Icon          Icon
}
type Icon struct {
	Asset []Asset `json:"assets"`
}
type Asset struct {
	Key  string `json:"key"`
	HREF string `json:"value"`
}
type AuctionFiles struct {
	Info []AuctionMeta `json:"files"`
}
