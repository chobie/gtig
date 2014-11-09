package twitter

import (
	"github.com/garyburd/go-oauth/oauth"
	"net/url"
	"net/http"
	"fmt"
	"encoding/json"
	"bufio"

	"io/ioutil"
)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
	TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
}


type Twitter struct {
	oauth oauth.Client
	token *oauth.Credentials
}

type Stream struct {
	res *http.Response
	reader *bufio.Reader
}

type User struct {
	ProfileSidebarFilColor string `json:"profile_sidebar_fill_color"`
	ProfileSidebarBorderColor string `json:"profile_sidebar_border_color"`
	ProfileBackgroundTile bool `json:"profile_background_tile"`
	Name string `json:name`
	ProfileImageUrl string `json:"profile_image_url"`
	CreatedAt string `json:"created_at"`
	Location string `json:"location"`
	FollowRequestSent bool `json:"follow_request_sent"`
	ProfileLinkColor string `json:"profile_link_color"`
	IsTranslator bool `json:"is_translator"`
	IsTranslationEnabled bool `json:"is_translation_enabled"`
	// Entities
	DefaultProfile bool `json:"default_profile"`
	ContributorsEnabled bool `json:"contributors_enabled"`
	FavouritesCount int `json:"favourites_count"`
	Url string `json:"url"`
	ProfileImageUrlHttps string `json:"profile_image_url_https"`
	UtcOffset int `json:"utc_offset"`
	Id int64 `json:"id"`
	ProfileUseBackgroundImage bool `json:"profile_use_background_image"`
	ListedCount int `json:"listed_count'`
	ProfileTextColor string `json:"profile_text_color"`
	Lang string `json:"lang"`
	FollowersCount int `json:"followers_count"`
	Protected bool `json:"protected"`
	Notifications bool `json:"notifications"`
	ProfileBackgroundImageUrlHttps string `json:"profile_background_image_url_https"`
	ProfileBackgroundColor string `json:"profile_background_color"`
	Verified bool `json:"verified"`
	GeoEnabled bool `json:"geo_enabled"`
	TimeZOne string `json:"time_zone"`
	Description string `json:"description"`
	DefaultProfileImage bool `json:"default_profile_image"`
	ProfileBackgroundImageUrl string `json:"profile_background_image_url"`
	StatusesCount int `json:"statuses_count"`
	FriendsCount int `json:"friends_count"`
	Following bool `json:"following"`
	ShowAllInlineMedia bool `json:"show_all_inline_media"`
	ScreenName string `json:"screen_name"`
}

type Url struct {
	ExpandedUrl string `json:"expanded_url"`
	Url string `json:"url"`
	Indices []int `json:"indices"`
	DisplayUrl string `json:"display_url"`
}


type HashTag struct {
	Text string `json:"text"`
	Indices []int `json:"indices"`
}

type Size struct {
	W int `json:w`
	H int `json:h`
	Resize string `json:"resize"`
}

type Media struct {
	Id int64 `json:"id"`
	Indices []int `json:"indices"`
	MediaUrl string `json:"media_url"`
	MediaUrlHttps string `json:"media_url_https"`
	Url string `json:"url"`
	DisplayUrl string `json:"display_url"`
	ExpandedUrl string `json:"expanded_url"`
	Type string `json:"type"`
	Sizes map[string]Size `json:"size"`
}

type Mention struct {
	ScreenName string `json:"screen_name"`
	Name string `json:"Name"`
	Id int64 `json:"id"`
	Indices []int `json:"indices"`
}

type Entity struct {
	Urls []Url `json:"urls"`
	Hashtags []HashTag `json:"hashtags"`
	UserMentions []Mention `json:"user_mentions"`
	//Symbols
	//UserMentionds
	//Urls
	Media []Media `json:"media"`
}

type Geo struct {
	Type string `json:"type"`
	//Coordinates []float64 `json:"coordinates"`
}

type Place struct {
	Id string `json:"id"`
	Url string `json:"url"`
	PlaceType string `json:"place_type"`
	Name string `json:"name"`
	FullName string `json:"full_name"`
	CountryCode string `json:"country_code"`
	Country string `json:"country"`
	BoundingBox Geo `json:"bounding_box"`
}

type Tweet struct {
	// Coordinates bool `json:"coordinates"`
	Favorited bool `json:"favorited"`
	Truncated bool `json:"truncated"`
	CreatedAt string `json:"created_at"`

	Entities Entity `json:"entities"`
	ExpandedEntities Entity `json:"expanded_entities"`

	//InReplyToUserIdStr string `json:"in_reply_to_user_id_str"`
	Contributors interface{} `json:"contributors"`
	RetweetedStatus *Tweet `json:"retweeted_status"`

	Text string `json:"text"`
	RetweetCount int `json:"retweet_count"`
	//InReplyToIdStr string `json:"in_reply_to_id_str"`
	Id int64 `json:"id"`
	Geo *Geo `json:"geo"`
	Coordinates *Geo `json:"coordinates"`
	Retweeted bool `json:"retweeted"`
	PossiblySensitive bool `json:"possibly_sensitive"`
	InReplyToUserId int64 `json:"in_reply_to_user_id"`
	Place *Place `json:"place"`

	User *User `json:"user"`
	InReplyToScreenName string `json:"in_reply_to_screen_name"`

	Source interface{} `json:"source"`
	InReplyToStatusId int64 `json:"in_reply_to_status_id"`
}

func NewClient(oauth_token, oauth_secret string) *Twitter{
	oauthClient.Credentials.Token = "uoSgZWThDlCDJA1G5GNZg"
	oauthClient.Credentials.Secret = "3nrp5n4evnJBOiT0ssPvtz7LZXaw8W5jFtBtBKUwG4"

	token := &oauth.Credentials{
		oauth_token,
		oauth_secret,
	}
	tw := &Twitter {
		oauth: oauthClient,
		token: token,
	}
	return tw
}

func (self *Twitter) Update(status string, opt map[string]string) {
	url_ := "https://api.twitter.com/1.1/statuses/update.json"

	param := make(url.Values)
	param.Set("status", status)

	for k, v := range opt {
		param.Set(k, v)
	}
	oauthClient.SignParam(self.token, "POST", url_, param)
	res, _ := http.PostForm(url_, url.Values(param))

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return
	}
}

func (self *Twitter) Favorite(id string) {
	url_ := "https://api.twitter.com/1.1/favorites/create.json"

	param := make(url.Values)
	param.Set("id", id)

	oauthClient.SignParam(self.token, "POST", url_, param)
	res, _ := http.PostForm(url_, url.Values(param))
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return
	}
}

type SearchResult struct {
	Statuses []Tweet `json:"statuses"`
}

func (self *Twitter) Search(query string) (*SearchResult, error) {
	url_ := "https://api.twitter.com/1.1/search/tweets.json"

	param := make(url.Values)
	param.Set("q", query)

	oauthClient.SignParam(self.token, "GET", url_, param)
	url_ = url_ + "?" + param.Encode()
	res, err := http.Get(url_)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
	}

	var tweets SearchResult
	err = json.NewDecoder(res.Body).Decode(&tweets)
	if err != nil {
		panic(err)
	}

	return &tweets, nil
}

func (self *Twitter) Retweet(id string) {
	url_ := fmt.Sprintf("https://api.twitter.com/1.1/statuses/retweet/%s.json", id)

	param := make(url.Values)
	oauthClient.SignParam(self.token, "POST", url_, param)
	res, _ := http.PostForm(url_, param)

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return
	}
}



func (self *Twitter) Statuses() {
	url_ := "https://api.twitter.com/1.1/statuses/user_timeline.json"
	opt := map[string]string{"screen_name": "chobi_e"}

	param := make(url.Values)
	for k, v := range opt {
		param.Set(k, v)
	}

	oauthClient.SignParam(self.token, "GET", url_, param)
	url_ = url_ + "?" + param.Encode()

	res, err := http.Get(url_)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return
	}

	var tweets []Tweet
	json.NewDecoder(res.Body).Decode(&tweets)
	out, _ := json.MarshalIndent(tweets, "", "  ")
	fmt.Printf("%s\n", out)
}

func (self *Twitter) Tweet(message string) {
	self.Update(message, nil)
}

type DeleteEventStatus struct {
	Id int64 `json:"id"`
	UserId int64 `json:"user_id"`
}

type DeleteEvent struct {
	Status DeleteEventStatus
}

type LimitEvent struct {
	Track int
}


type StatusWithheldEvent struct {
	Id int64 `json:"id"`
	UserId int64 `json:"user_id"`
	WithheldInCountries []string `json:"withhel_in_countries"`
}

type UserWithheldEvent struct {
	Id int64 `json:"id"`
	WithheldInCountries []string `json:"withhel_in_countries"`
}

type DisconnectEvent struct {
	Code int `json:"code"`
	StreamName string `json:"stream_name"`
	Reason string `json:"reason"`
}

type StallWarningEvent struct {
	Code string `json:"code"`
	Message string `json:"message"`
	PercentFull int `json:"percent_full"`
}

type Event struct {
	CreatedAt string `json:"created_at"`
	Event string `json:"event"`
}

type TweetEvent struct {
	Tweet
}

type StreamEvent struct {
	Tweet

	Friends []int64 `json:"friends"`
	Delete *DeleteEvent `json:"delete_event"`
	Limit *LimitEvent `json:"limit"`
	StatusWithheld *StatusWithheldEvent `json:"status_withheld"`
	UserWithheld *UserWithheldEvent `json:"user_withheld"`
	Disconnect *DisconnectEvent `json:"disconnect_event"`
	Warning *StallWarningEvent `json:"warning"`
	Event string `json:"event"`

	Target interface{} `json:"target"`
	Source interface{} `json:"source"`
	TargetObject interface{} `json:"target_object"`
}

type EventType int

const (
	TWEET_EVENT EventType = iota
	DELETE_EVENT
	LIMIT_EVENT
	STATUSWITHHELD_EVENT
	USERWITHHELD_EVENT
	DISCONNECT_EVENT
	KEEPALIVE_EVENT
	WARNING_EVENT
	FRIENDLIST_EVENT
	FAVORITE_EVENT
)

func (self *StreamEvent) EventType() EventType {
	if self.Text != "" {
		return TWEET_EVENT
	} else if self.Delete != nil {
		return DELETE_EVENT
	} else if self.Limit != nil {
		return LIMIT_EVENT
	} else if self.StatusWithheld != nil {
		return STATUSWITHHELD_EVENT
	} else if self.UserWithheld != nil {
		return USERWITHHELD_EVENT
	} else if self.Disconnect != nil {
		return DISCONNECT_EVENT
	} else if self.Warning != nil {
		return WARNING_EVENT
	} else if self.Event == "favorite" {
		return FAVORITE_EVENT
	} else {
		return KEEPALIVE_EVENT
	}
}

func (self *Twitter) UserStream() *Stream {
	url_ := "https://userstream.twitter.com/1.1/user.json"
	param := make(url.Values)

	oauthClient.SignParam(self.token, "GET", url_, param)
	url_ = url_ + "?" + param.Encode()

	res, err := http.Get(url_)
	if err != nil {
		return nil
	}

	s := &Stream{
		res: res,
		reader: bufio.NewReaderSize(res.Body, 1024 * 1024 * 16),
	}

	return s
}

func (self *Stream) Close() {
	self.res.Body.Close()
}

func (self *Stream) Next() (*StreamEvent, error) {
	var ev StreamEvent
	line, _, err := self.reader.ReadLine()
	if err != nil {
		panic(err)
		return nil, err
	}

	fmt.Printf("> %s\n", line)
	if len(line) == 0 {
		return &ev, nil
	}

	err = json.Unmarshal(line, &ev)
	if err != nil {
		fmt.Printf(">>> %s <<<\n", line)
	}

	fmt.Printf("%#v\n", ev)
	return &ev, err
}

func (self *Twitter) Debug() (*StreamEvent, error){
	line, _ := ioutil.ReadFile("data")

	var ev StreamEvent
	err := json.Unmarshal(line, &ev)
	if err != nil {
		fmt.Printf(">>> %s <<<\n", line)
		fmt.Printf("%T %s\n", err, err)
	}

	fmt.Printf("%#v\n", ev)

	fmt.Printf("%s > %s\n",
		ev.TargetObject.(map[string]interface{})["user"].(map[string]interface{})["screen_name"].(string),
		ev.TargetObject.(map[string]interface{})["text"].(string))

	return nil, nil
}
