package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/gocolly/colly"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
	"google.golang.org/api/option"
)

type CoinValuesResponse struct {
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
		Notice       string    `json:"notice"`
	} `json:"status"`
	Data map[string][]CoinData `json:"data"`
}

type CoinData struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Symbol         string    `json:"symbol"`
	Slug           string    `json:"slug"`
	NumMarketPairs int       `json:"num_market_pairs"`
	DateAdded      time.Time `json:"date_added"`
	Tags           []struct {
		Slug     string `json:"slug"`
		Name     string `json:"name"`
		Category string `json:"category"`
	} `json:"tags"`
	MaxSupply                     int         `json:"max_supply"`
	CirculatingSupply             float64     `json:"circulating_supply"`
	TotalSupply                   float64     `json:"total_supply"`
	IsActive                      int         `json:"is_active"`
	InfiniteSupply                bool        `json:"infinite_supply"`
	Platform                      interface{} `json:"platform"`
	CmcRank                       int         `json:"cmc_rank"`
	IsFiat                        int         `json:"is_fiat"`
	SelfReportedCirculatingSupply interface{} `json:"self_reported_circulating_supply"`
	SelfReportedMarketCap         interface{} `json:"self_reported_market_cap"`
	TVLRatio                      interface{} `json:"tvl_ratio"`
	LastUpdated                   time.Time   `json:"last_updated"`
	Quote                         struct {
		USD struct {
			Price                 float64     `json:"price"`
			Volume24h             float64     `json:"volume_24h"`
			VolumeChange24h       float64     `json:"volume_change_24h"`
			PercentChange1h       float64     `json:"percent_change_1h"`
			PercentChange24h      float64     `json:"percent_change_24h"`
			PercentChange7d       float64     `json:"percent_change_7d"`
			PercentChange30d      float64     `json:"percent_change_30d"`
			MarketCap             float64     `json:"market_cap"`
			MarketCapDominance    float64     `json:"market_cap_dominance"`
			FullyDilutedMarketCap float64     `json:"fully_diluted_market_cap"`
			TVL                   interface{} `json:"tvl"`
			LastUpdated           time.Time   `json:"last_updated"`
		} `json:"USD"`
	} `json:"quote"`
}

type CoinSentiment struct {
	id        int
	createdAt time.Time
	sentiment int // -1: negative, 0: neutral, 1: positive
	coin      string
	source    string
}

// create struct for usb_conversions table
type CoinConversion struct {
	createdAt time.Time
	value     float64
	coin      string
}

type GhostPosts struct {
	Posts []GhostPost `json:"posts"`
}

type GhostPost struct {
	Slug  string `json:"slug"`
	ID    string `json:"id"`
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	HTML  string `json:"html"`
	//Lexical      string `json:"lexical"`
	FeatureImage string `json:"feature_image"`
	Featured     bool   `json:"featured"`
	Status       string `json:"status"`
	Visibility   string `json:"visibility"` // members, public
}

type RandomUnSplashResponse struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	AlternativeSlugs struct {
		En string `json:"en"`
		Es string `json:"es"`
		Ja string `json:"ja"`
		Fr string `json:"fr"`
		It string `json:"it"`
		Ko string `json:"ko"`
		De string `json:"de"`
		Pt string `json:"pt"`
	} `json:"alternative_slugs"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	PromotedAt     interface{}   `json:"promoted_at"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	Color          string        `json:"color"`
	BlurHash       string        `json:"blur_hash"`
	Description    string        `json:"description"`
	AltDescription string        `json:"alt_description"`
	Breadcrumbs    []interface{} `json:"breadcrumbs"`
	Urls           struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
		SmallS3 string `json:"small_s3"`
	} `json:"urls"`
	Links struct {
		Self             string `json:"self"`
		HTML             string `json:"html"`
		Download         string `json:"download"`
		DownloadLocation string `json:"download_location"`
	} `json:"links"`
	Likes                  int           `json:"likes"`
	LikedByUser            bool          `json:"liked_by_user"`
	CurrentUserCollections []interface{} `json:"current_user_collections"`
	Sponsorship            interface{}   `json:"sponsorship"`
	TopicSubmissions       struct {
	} `json:"topic_submissions"`
	AssetType string `json:"asset_type"`
	User      struct {
		ID              string      `json:"id"`
		UpdatedAt       time.Time   `json:"updated_at"`
		Username        string      `json:"username"`
		Name            string      `json:"name"`
		FirstName       string      `json:"first_name"`
		LastName        interface{} `json:"last_name"`
		TwitterUsername interface{} `json:"twitter_username"`
		PortfolioURL    interface{} `json:"portfolio_url"`
		Bio             interface{} `json:"bio"`
		Location        interface{} `json:"location"`
		Links           struct {
			Self      string `json:"self"`
			HTML      string `json:"html"`
			Photos    string `json:"photos"`
			Likes     string `json:"likes"`
			Portfolio string `json:"portfolio"`
			Following string `json:"following"`
			Followers string `json:"followers"`
		} `json:"links"`
		ProfileImage struct {
			Small  string `json:"small"`
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"profile_image"`
		InstagramUsername   interface{} `json:"instagram_username"`
		TotalCollections    int         `json:"total_collections"`
		TotalLikes          int         `json:"total_likes"`
		TotalPhotos         int         `json:"total_photos"`
		TotalPromotedPhotos int         `json:"total_promoted_photos"`
		AcceptedTos         bool        `json:"accepted_tos"`
		ForHire             bool        `json:"for_hire"`
		Social              struct {
			InstagramUsername interface{} `json:"instagram_username"`
			PortfolioURL      interface{} `json:"portfolio_url"`
			TwitterUsername   interface{} `json:"twitter_username"`
			PaypalEmail       interface{} `json:"paypal_email"`
		} `json:"social"`
	} `json:"user"`
	Exif struct {
		Make         string `json:"make"`
		Model        string `json:"model"`
		Name         string `json:"name"`
		ExposureTime string `json:"exposure_time"`
		Aperture     string `json:"aperture"`
		FocalLength  string `json:"focal_length"`
		Iso          int    `json:"iso"`
	} `json:"exif"`
	Location struct {
		Name     interface{} `json:"name"`
		City     interface{} `json:"city"`
		Country  interface{} `json:"country"`
		Position struct {
			Latitude  interface{} `json:"latitude"`
			Longitude interface{} `json:"longitude"`
		} `json:"position"`
	} `json:"location"`
	Meta struct {
		Index bool `json:"index"`
	} `json:"meta"`
	PublicDomain bool `json:"public_domain"`
	Tags         []struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Source struct {
			Ancestry struct {
				Type struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"type"`
				Category struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"category"`
				Subcategory struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"subcategory"`
			} `json:"ancestry"`
			Title           string `json:"title"`
			Subtitle        string `json:"subtitle"`
			Description     string `json:"description"`
			MetaTitle       string `json:"meta_title"`
			MetaDescription string `json:"meta_description"`
			CoverPhoto      struct {
				ID               string `json:"id"`
				Slug             string `json:"slug"`
				AlternativeSlugs struct {
					En string `json:"en"`
					Es string `json:"es"`
					Ja string `json:"ja"`
					Fr string `json:"fr"`
					It string `json:"it"`
					Ko string `json:"ko"`
					De string `json:"de"`
					Pt string `json:"pt"`
				} `json:"alternative_slugs"`
				CreatedAt      time.Time   `json:"created_at"`
				UpdatedAt      time.Time   `json:"updated_at"`
				PromotedAt     interface{} `json:"promoted_at"`
				Width          int         `json:"width"`
				Height         int         `json:"height"`
				Color          string      `json:"color"`
				BlurHash       string      `json:"blur_hash"`
				Description    interface{} `json:"description"`
				AltDescription string      `json:"alt_description"`
				Breadcrumbs    []struct {
					Slug  string `json:"slug"`
					Title string `json:"title"`
					Index int    `json:"index"`
					Type  string `json:"type"`
				} `json:"breadcrumbs"`
				Urls struct {
					Raw     string `json:"raw"`
					Full    string `json:"full"`
					Regular string `json:"regular"`
					Small   string `json:"small"`
					Thumb   string `json:"thumb"`
					SmallS3 string `json:"small_s3"`
				} `json:"urls"`
				Links struct {
					Self             string `json:"self"`
					HTML             string `json:"html"`
					Download         string `json:"download"`
					DownloadLocation string `json:"download_location"`
				} `json:"links"`
				Likes                  int           `json:"likes"`
				LikedByUser            bool          `json:"liked_by_user"`
				CurrentUserCollections []interface{} `json:"current_user_collections"`
				Sponsorship            interface{}   `json:"sponsorship"`
				TopicSubmissions       struct {
				} `json:"topic_submissions"`
				AssetType string `json:"asset_type"`
				User      struct {
					ID              string      `json:"id"`
					UpdatedAt       time.Time   `json:"updated_at"`
					Username        string      `json:"username"`
					Name            string      `json:"name"`
					FirstName       string      `json:"first_name"`
					LastName        string      `json:"last_name"`
					TwitterUsername string      `json:"twitter_username"`
					PortfolioURL    string      `json:"portfolio_url"`
					Bio             interface{} `json:"bio"`
					Location        string      `json:"location"`
					Links           struct {
						Self      string `json:"self"`
						HTML      string `json:"html"`
						Photos    string `json:"photos"`
						Likes     string `json:"likes"`
						Portfolio string `json:"portfolio"`
						Following string `json:"following"`
						Followers string `json:"followers"`
					} `json:"links"`
					ProfileImage struct {
						Small  string `json:"small"`
						Medium string `json:"medium"`
						Large  string `json:"large"`
					} `json:"profile_image"`
					InstagramUsername   string `json:"instagram_username"`
					TotalCollections    int    `json:"total_collections"`
					TotalLikes          int    `json:"total_likes"`
					TotalPhotos         int    `json:"total_photos"`
					TotalPromotedPhotos int    `json:"total_promoted_photos"`
					AcceptedTos         bool   `json:"accepted_tos"`
					ForHire             bool   `json:"for_hire"`
					Social              struct {
						InstagramUsername string      `json:"instagram_username"`
						PortfolioURL      string      `json:"portfolio_url"`
						TwitterUsername   string      `json:"twitter_username"`
						PaypalEmail       interface{} `json:"paypal_email"`
					} `json:"social"`
				} `json:"user"`
			} `json:"cover_photo"`
		} `json:"source,omitempty"`
	} `json:"tags"`
	TagsPreview []struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Source struct {
			Ancestry struct {
				Type struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"type"`
				Category struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"category"`
				Subcategory struct {
					Slug       string `json:"slug"`
					PrettySlug string `json:"pretty_slug"`
				} `json:"subcategory"`
			} `json:"ancestry"`
			Title           string `json:"title"`
			Subtitle        string `json:"subtitle"`
			Description     string `json:"description"`
			MetaTitle       string `json:"meta_title"`
			MetaDescription string `json:"meta_description"`
			CoverPhoto      struct {
				ID               string `json:"id"`
				Slug             string `json:"slug"`
				AlternativeSlugs struct {
					En string `json:"en"`
					Es string `json:"es"`
					Ja string `json:"ja"`
					Fr string `json:"fr"`
					It string `json:"it"`
					Ko string `json:"ko"`
					De string `json:"de"`
					Pt string `json:"pt"`
				} `json:"alternative_slugs"`
				CreatedAt      time.Time   `json:"created_at"`
				UpdatedAt      time.Time   `json:"updated_at"`
				PromotedAt     interface{} `json:"promoted_at"`
				Width          int         `json:"width"`
				Height         int         `json:"height"`
				Color          string      `json:"color"`
				BlurHash       string      `json:"blur_hash"`
				Description    interface{} `json:"description"`
				AltDescription string      `json:"alt_description"`
				Breadcrumbs    []struct {
					Slug  string `json:"slug"`
					Title string `json:"title"`
					Index int    `json:"index"`
					Type  string `json:"type"`
				} `json:"breadcrumbs"`
				Urls struct {
					Raw     string `json:"raw"`
					Full    string `json:"full"`
					Regular string `json:"regular"`
					Small   string `json:"small"`
					Thumb   string `json:"thumb"`
					SmallS3 string `json:"small_s3"`
				} `json:"urls"`
				Links struct {
					Self             string `json:"self"`
					HTML             string `json:"html"`
					Download         string `json:"download"`
					DownloadLocation string `json:"download_location"`
				} `json:"links"`
				Likes                  int           `json:"likes"`
				LikedByUser            bool          `json:"liked_by_user"`
				CurrentUserCollections []interface{} `json:"current_user_collections"`
				Sponsorship            interface{}   `json:"sponsorship"`
				TopicSubmissions       struct {
				} `json:"topic_submissions"`
				AssetType string `json:"asset_type"`
				User      struct {
					ID              string      `json:"id"`
					UpdatedAt       time.Time   `json:"updated_at"`
					Username        string      `json:"username"`
					Name            string      `json:"name"`
					FirstName       string      `json:"first_name"`
					LastName        string      `json:"last_name"`
					TwitterUsername string      `json:"twitter_username"`
					PortfolioURL    string      `json:"portfolio_url"`
					Bio             interface{} `json:"bio"`
					Location        string      `json:"location"`
					Links           struct {
						Self      string `json:"self"`
						HTML      string `json:"html"`
						Photos    string `json:"photos"`
						Likes     string `json:"likes"`
						Portfolio string `json:"portfolio"`
						Following string `json:"following"`
						Followers string `json:"followers"`
					} `json:"links"`
					ProfileImage struct {
						Small  string `json:"small"`
						Medium string `json:"medium"`
						Large  string `json:"large"`
					} `json:"profile_image"`
					InstagramUsername   string `json:"instagram_username"`
					TotalCollections    int    `json:"total_collections"`
					TotalLikes          int    `json:"total_likes"`
					TotalPhotos         int    `json:"total_photos"`
					TotalPromotedPhotos int    `json:"total_promoted_photos"`
					AcceptedTos         bool   `json:"accepted_tos"`
					ForHire             bool   `json:"for_hire"`
					Social              struct {
						InstagramUsername string      `json:"instagram_username"`
						PortfolioURL      string      `json:"portfolio_url"`
						TwitterUsername   string      `json:"twitter_username"`
						PaypalEmail       interface{} `json:"paypal_email"`
					} `json:"social"`
				} `json:"user"`
			} `json:"cover_photo"`
		} `json:"source,omitempty"`
	} `json:"tags_preview"`
	Views     int           `json:"views"`
	Downloads int           `json:"downloads"`
	Topics    []interface{} `json:"topics"`
}

type MarketForecast struct {
	Description string
	Currency    string
	Price       float64
	Chatter     []string
	OneWeek     float64
	OneMonth    float64
	ThreeMonths float64
	Change24h   float64
}

var forecastTemplate = `
<!DOCTYPE html>
<html>
<body>
	<p>{{.Description}}</p>
	<table border="1">
		<tr>
			<th>Currency</th>
			<th>Price (USD)</th>
			<th>Change 24h (%)</th>
		</tr>
		<tr>
			<td>{{.Currency}}</td>
			<td>{{.Price}}</td>
			<td>{{.Change24h}}</td>
		</tr>
	</table>
	<h2>Forecast</h2>
	<ul>
		<li>1 Week: {{.OneWeek}}</li>
		<li>1 Month: {{.OneMonth}}</li>
		<li>3 Months: {{.ThreeMonths}}</li>
	</ul>
	<h2>Chatter</h2>
	<ul>
		{{range .Chatter}}
		<li>{{.}}</li>
		{{end}}
	</ul>
</body>
</html>
`

var db *pgxpool.Pool

func main() {
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to the database
	db, err = pgxpool.New(context.Background(), os.Getenv("databaseUrl"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}

	// market watcher job
	_, err = s.NewJob(
		gocron.DurationJob(
			15*time.Minute,
		),
		gocron.NewTask(
			func() {
				log.Println("Checking market values")
				getAllCoinValues()
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// predictions job
	_, err = s.NewJob(
		gocron.DurationJob(
			60*time.Minute*24, // every 24 hours
		),
		gocron.NewTask(
			func() {
				if time.Now().Weekday() == time.Sunday {
					log.Println("Creating predictions")
					postPredictions()
				}
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// rss job
	_, err = s.NewJob(
		gocron.DurationJob(
			1*time.Minute,
		),
		gocron.NewTask(
			func() {
				log.Println("Checking rss")
				checkFeedsAndPost()
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	s.Start()

	// create a web server to keep the program running and to provide a health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	http.ListenAndServe(":8080", nil)
}

func postPredictions() {
	dailyForecast("BTC")
	dailyForecast("ETH")
	dailyForecast("LTC")
}

func checkFeedsAndPost() {
	feeds := []string{
		"https://www.coindesk.com/feed",
		"https://cointelegraph.com/rss",
		"https://cryptopotato.com/feed",
		"https://cryptoslate.com/feed",
		"https://cryptonews.com/feed",
		"https://cryptobriefing.com/feed",
		"https://cryptocurrencynews.com/feed",
		"https://cryptoslate.com/feed",
	}

	for _, feed := range feeds {
		log.Println("Parsing feed: ", feed)

		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(feed)
		if err != nil {
			log.Println(err)
		}

		for _, item := range feed.Items {
			time.Sleep(30 * time.Second)
			log.Println("Parsing article: ", item.Title)

			if isArticleAlreadyParaphrased(item.Link) {
				log.Println("Article already paraphrased")
				continue
			}

			text := getTextFromArticle(item.Link)
			pContent, pTitle, err := paraphrase(text[0], item.Title)
			if err != nil {
				log.Println(err)
				continue
			}
			determineHeadlineSetiment(item.Title, "BTC", item.Link)

			standardPost(pContent, pTitle, item.Link)
		}
	}
}

func getAllCoinValues() {
	getCoinValue("BTC")
	time.Sleep(10 * time.Second)

	getCoinValue("ETH")
	time.Sleep(10 * time.Second)

	getCoinValue("LTC")
	time.Sleep(10 * time.Second)

	getCoinValue("DOGE")
	time.Sleep(10 * time.Second)

	getCoinValue("SHIB")
	time.Sleep(10 * time.Second)

	getCoinValue("LINK")
	time.Sleep(10 * time.Second)

	getCoinValue("XMR")
	time.Sleep(10 * time.Second)

	getCoinValue("SOL")
	time.Sleep(10 * time.Second)

	getCoinValue("USDT")
	time.Sleep(10 * time.Second)

	getCoinValue("XTZ")
	time.Sleep(10 * time.Second)
}

var disclaimer = "This is not financial advice. This is for entertainment purposes only. Do your own research before making any investment. The author is not responsible for any losses incurred. The information on this page is simply opinion based on publicly available data"

func isArticleAlreadyParaphrased(url string) bool {
	// check if the article is already in the database, table rss_posts
	rows, err := db.Query(context.Background(), "SELECT * FROM rss_posts WHERE url = $1", url)
	if err != nil {
		log.Printf("Error querying database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		return true
	}

	return false
}

func standardPost(content string, title string, source string) {
	createPost(GhostPost{
		Title:        title,
		HTML:         content + "<br><br><a href='" + source + "'>Source</a>",
		FeatureImage: fetchUnsplashImage("cryptocurrency").Urls.Small,
		Featured:     false,
		Status:       "published",
		Visibility:   "public",
	})
}

func dailyForecast(coin string) {
	curr := getCoinValue(coin)
	week, _ := strconv.ParseFloat(forecast(coin, "1 week"), 64)
	month, _ := strconv.ParseFloat(forecast(coin, "1 month"), 64)
	threeMonths, _ := strconv.ParseFloat(forecast(coin, "3 months"), 64)
	//desc := generateForecastDescription(coin, curr, week, month, threeMonths)

	forecastData := MarketForecast{
		Description: disclaimer,
		Currency:    coin,
		Price:       curr,
		Chatter:     sentimentUrlList(),
		OneWeek:     week,
		OneMonth:    month,
		ThreeMonths: threeMonths,
		Change24h:   getPercentChange24h(coin),
	}

	tmpl, err := template.New("forecast").Parse(forecastTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// write to string
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, forecastData)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	createPost(GhostPost{
		Title:        "Weekly " + coin,
		HTML:         tpl.String(),
		FeatureImage: fetchUnsplashImage("cryptocurrency").Urls.Small,
		Featured:     true,
		Status:       "published",
		Visibility:   "public",
	})
}

func getPercentChange24h(c string) float64 {
	// get the 24h percent change of a cryptocurrency from a Postgres database only. Do not use API
	change := 0.0

	// get the values from the database
	rows, err := db.Query(context.Background(), "SELECT * FROM exchange_rates WHERE coin = $1 ORDER BY created_at DESC LIMIT 2", c)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return change
	}
	defer rows.Close()

	var values []CoinConversion
	for rows.Next() {
		var value CoinConversion
		err = rows.Scan(&value.createdAt, &value.coin, &value.value)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		values = append(values, value)
	}

	if len(values) == 2 {
		change = (values[0].value - values[1].value) / values[1].value
		change = change * 100
	}

	return change
}

func getCoinValue(c string) float64 {
	// get the value of a cryptocurrency from a Postgres database. If the value is not found, or is older than 4 hours, get it from an API.
	// If the API is down, return the most recent value from the database.
	localVal, err := getValueFromPostgres(c)
	if err != nil {
		log.Println(err)
		return 0
	}
	if localVal.createdAt.Before(time.Now().Add(-4 * time.Hour)) {
		log.Println("Value is older than a 4 hours, getting from API -" + c)

		// get the value from the API
		values, err := getCoinValuesFromAPI(c)
		if err != nil {
			log.Println(err)
			return localVal.value
		}
		saveCoinValuesToPostgres([]CoinConversion{
			{
				value: values.Data[c][0].Quote.USD.Price,
				coin:  c,
			},
		})
		return values.Data[c][0].Quote.USD.Price
	}

	log.Println("Value loaded from database")
	return localVal.value
}

// wrapper function to prioritize getting values from Postgres database
func getCoinValuesTimeRange(start int64, coin string) []CoinConversion {
	// the API does not provide historical data, so we must rely on the database for this information.
	// get all values from the database from start to now
	values := []CoinConversion{}

	// get the values from the database
	rows, err := db.Query(context.Background(), "SELECT * FROM exchange_rates WHERE coin = ? AND created_at > ? ORDER BY created_at DESC", coin, time.Unix(start, 0))
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return values
	}
	defer rows.Close()

	for rows.Next() {
		var value CoinConversion
		err = rows.Scan(&value.createdAt, &value.coin, &value.value)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		values = append(values, value)
	}

	return values
}

// get the value of a cryptocurrency from a Postgres database
func getValueFromPostgres(c string) (CoinConversion, error) {
	rows, err := db.Query(context.Background(), "SELECT * FROM exchange_rates WHERE coin = $1 ORDER BY created_at DESC LIMIT 1", c)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return CoinConversion{}, err
	}
	defer rows.Close()

	var value CoinConversion
	for rows.Next() {
		err = rows.Scan(&value.createdAt, &value.coin, &value.value)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
	}

	return value, nil
}

// add up all the sentiment values for the coin from the past h hours
func getCoinSentiment(c string, h int) int {
	rows, err := db.Query(context.Background(), "SELECT * FROM sentiments WHERE coin = ? AND created_at > ? ORDER BY created_at DESC", c, time.Now().Add(-1*time.Hour*time.Duration(h)))
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return 0
	}
	defer rows.Close()

	var value int
	for rows.Next() {
		var sentiment CoinSentiment
		err = rows.Scan(&sentiment.id, &sentiment.createdAt, &sentiment.coin, &sentiment.sentiment, &sentiment.source)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		value += sentiment.sentiment
	}

	return value
}

// save the sentiment of a cryptocurrency to a Postgres database
func saveCoinSentimentToPostgres(sentiments []CoinSentiment) {
	for _, sentiment := range sentiments {
		_, err := db.Exec(context.Background(), "INSERT INTO sentiments (sentiment, coin, source) VALUES ($1, $2, $3)", sentiment.sentiment, sentiment.coin, sentiment.source)
		if err != nil {
			log.Println(err)
		}
	}
}

// save the value of a cryptocurrency to a Postgres database
func saveCoinValuesToPostgres(values []CoinConversion) {
	for _, value := range values {
		_, err := db.Exec(context.Background(), "INSERT INTO exchange_rates (value, coin) VALUES ($1, $2)", value.value, value.coin)
		if err != nil {
			log.Println(err)
		}
	}
}

// get the value of a cryptocurrency from an API. coins is a comma-separated list of coin symbols
func getCoinValuesFromAPI(coin string) (CoinValuesResponse, error) {
	log.Println("Getting coin values from API")

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		return CoinValuesResponse{}, err
	}

	q := url.Values{}
	q.Add("symbol", coin)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", os.Getenv("coinmarketKey"))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to server")
		return CoinValuesResponse{}, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body")
		return CoinValuesResponse{}, err
	}

	// convert response to struct
	var coinValues CoinValuesResponse
	err = json.Unmarshal(respBody, &coinValues)
	if err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return coinValues, err
	}

	return coinValues, nil
}

// scrape body text from articles using colly v2
func getTextFromArticle(url string) []string {
	text := []string{}

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
	c.Limit(&colly.LimitRule{
		Delay: 5 * time.Second,
	})
	c.IgnoreRobotsTxt = true
	c.CacheDir = "./cache"

	// collect text from the body of the article
	c.OnHTML("p", func(e *colly.HTMLElement) {
		text = append(text, e.Text)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		log.Println("Finished", r.Request.URL)
	})

	c.Visit(url)

	return text
}

func fetchUnsplashImage(query string) RandomUnSplashResponse {
	// replace spaces with %20
	query = strings.ReplaceAll(query, " ", "%20")
	token := os.Getenv("unsplashBearer")
	url := "https://api.unsplash.com/photos/random?client_id=583846&query=" + query
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	var response RandomUnSplashResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
	}

	return response
}

func generateJwt() string {
	secret := os.Getenv("adminKey")
	id := os.Getenv("adminId")

	// Decode hexadecimal secret into original binary byte array
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		fmt.Println("Error decoding secret:", err)
	}

	// Create a new token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the key ID in the token header
	token.Header["kid"] = id

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	claims["iat"] = time.Now().Unix()
	claims["aud"] = "/admin/"

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(secretBytes)
	if err != nil {
		fmt.Println("Error creating token:", err)
	}

	return tokenString
}

func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func createLineChart() string {
	// returns html of a line chart
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "My first bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())

	// create writer to write to a string
	f := &strings.Builder{}

	bar.Render(f)

	htmlString := f.String()
	htmlString = "<!--kg-card-begin: html-->" + htmlString
	htmlString = htmlString + "<!--kg-card-end: html-->"

	return htmlString
}

func createPost(content GhostPost) {
	posts := GhostPosts{
		Posts: []GhostPost{
			content,
		},
	}

	url := "https://cryptobunker.org/ghost/api/admin/posts/?source=html"
	method := "POST"

	json, err := json.Marshal(posts)
	if err != nil {
		fmt.Println(err)
		return
	}
	payload := strings.NewReader(string(json))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Ghost "+generateJwt())
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
}

func paraphrase(text string, source string) (string, string, error) {
	// ensure text is less than 2048 characters
	if len(text) > 2048 {
		text = text[:2048]
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("geminiKey")))
	if err != nil {
		log.Println(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text("Paraphrase the following blog post from "+source+". Speak as if you're the one reporting the information and don't mention this information from elsewhere: "+text))
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	body := fmt.Sprint(resp.Candidates[0].Content.Parts[0])

	// now paraphrase the title
	resp, err = model.GenerateContent(ctx, genai.Text("Paraphrase the title of the following blog post from "+source+": "+text))
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	title := fmt.Sprint(resp.Candidates[0].Content.Parts[0])

	return body, title, nil
}

func determineHeadlineSetiment(text string, coin string, source string) (int, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("geminiKey")))
	if err != nil {
		log.Println(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text("return -1 if the following headline could have a negative impact on the market value of "+coin+", 1 for positive, 0 for no effect at all: "+text))
	if err != nil {
		log.Println(err)
		return 0, err
	}

	parsed, err := strconv.Atoi(fmt.Sprint(resp.Candidates[0].Content.Parts[0]))
	if err != nil {
		log.Println(err)
	}

	// save the sentiment to the database
	saveCoinSentimentToPostgres([]CoinSentiment{
		{
			createdAt: time.Now(),
			sentiment: parsed,
			coin:      coin,
			source:    source,
		},
	})

	// save source to rss_posts (id, created_at, url)
	_, err = db.Exec(ctx, "INSERT INTO rss_posts (url) VALUES ($1)", source)
	if err != nil {
		log.Println(err)
	}

	return parsed, nil
}

func forecast(coin string, timespan string) string {
	values := []int{}

	from := time.Now().Add(-1 * time.Hour * time.Duration(1000)).Unix()

	// get the values of the coin from the past h hours
	for _, value := range getCoinValuesTimeRange(from, coin) {
		values = append(values, int(value.value))
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("geminiKey")))
	if err != nil {
		log.Println(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text("You are a finicial consultant. It is required you give a best guess. Only provide a USD estimate and nothing else. Do not use special characters. just numbers. Forecast the price of "+coin+" "+timespan+" from now. Here is a list of recent values in USD from the last 168 hours "+fmt.Sprint(values)))
	if err != nil {
		log.Println(err)
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0])
}

func generateForecastDescription(coin string, current float64, week float64, month float64, months float64) string {
	// generate a description of the forecast
	prompt := "Speak objectively and do not speak in the first person. Return plain text without markdown or html, do not stylize. Based on the forecasted values of " + coin + " over the next week, month, and 3 months, provide a summary of the forecast." +
		"Current value: " + fmt.Sprint(current) + ", 1 week: " + fmt.Sprint(week) + ", 1 month: " + fmt.Sprint(month) + ", 3 months: " + fmt.Sprint(months)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("geminiKey")))
	if err != nil {
		log.Println(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Println(err)
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0])
}

// returns a string of html bullet points with sentiment analysis of recent news articles. Use the 10 most recent articles in the database. Does not use neutral sentiment.
func sentimentUrlList() []string {
	rows, err := db.Query(context.Background(), "SELECT * FROM sentiments WHERE sentiment <> 0 ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return []string{}
	}
	defer rows.Close()

	var sentimentsUrls []string
	for rows.Next() {
		var sentiment CoinSentiment
		err = rows.Scan(&sentiment.id, &sentiment.createdAt, &sentiment.coin, &sentiment.sentiment, &sentiment.source)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		sentimentsUrls = append(sentimentsUrls, sentiment.source)
	}

	return sentimentsUrls
}
