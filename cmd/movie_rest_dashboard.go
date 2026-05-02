// movie_rest_dashboard.go — JSON endpoints powering the HTML dashboard's
// filter sidebar, paginated card list, and per-item modal details.
package cmd

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// ----- /api/dashboard/filters -----------------------------------------------

type dashboardFilters struct {
	Tags       []db.TagCount  `json:"tags"`
	Types      map[string]int `json:"types"`
	Genres     []string       `json:"genres"`
	RatingMax  float64        `json:"rating_max"`
	YearMin    int            `json:"year_min"`
	YearMax    int            `json:"year_max"`
	TotalItems int            `json:"total_items"`
}

func handleDashboardFilters(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items, err := database.ListAllMedia()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tags, _ := database.GetAllTagCounts()
	writeJSON(w, buildDashboardFilters(items, tags))
}

func buildDashboardFilters(items []db.Media, tags []db.TagCount) dashboardFilters {
	out := dashboardFilters{Tags: tags, Types: map[string]int{}, TotalItems: len(items)}
	genreSet := map[string]struct{}{}
	for i := range items {
		collectFilterFacets(&items[i], &out, genreSet)
	}
	out.Genres = sortedKeys(genreSet)
	return out
}

func collectFilterFacets(m *db.Media, out *dashboardFilters, genres map[string]struct{}) {
	out.Types[m.Type]++
	for _, g := range splitGenres(m.Genre) {
		genres[g] = struct{}{}
	}
	if m.Year > 0 && (out.YearMin == 0 || m.Year < out.YearMin) {
		out.YearMin = m.Year
	}
	if m.Year > out.YearMax {
		out.YearMax = m.Year
	}
	if m.TmdbRating > out.RatingMax {
		out.RatingMax = m.TmdbRating
	}
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ----- /api/dashboard/list --------------------------------------------------

type dashboardListQuery struct {
	Type      string
	Genre     string
	Tag       string
	Search    string
	Sort      string
	MinRating float64
	YearFrom  int
	YearTo    int
	Limit     int
	Offset    int
}

type dashboardListResponse struct {
	Items  []dashboardCard `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type dashboardCard struct {
	Title         string   `json:"title"`
	Type          string   `json:"type"`
	Genre         string   `json:"genre"`
	GenreList     []string `json:"genre_list"`
	Director      string   `json:"director"`
	CastList      string   `json:"cast_list"`
	Description   string   `json:"description"`
	Tagline       string   `json:"tagline"`
	ThumbnailPath string   `json:"thumbnail_path"`
	Tags          []string `json:"tags"`
	ID            int64    `json:"id"`
	Year          int      `json:"year"`
	Runtime       int      `json:"runtime"`
	TmdbID        int      `json:"tmdb_id"`
	TmdbRating    float64  `json:"tmdb_rating"`
	Watched       bool     `json:"watched"`
}

func handleDashboardList(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := parseDashboardListQuery(r)
	items, err := database.ListAllMedia()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cards := buildDashboardCards(database, items)
	cards = applyDashboardFilters(cards, q)
	sortDashboardCards(cards, q.Sort)
	writeJSON(w, paginateDashboardCards(cards, q))
}

func parseDashboardListQuery(r *http.Request) dashboardListQuery {
	v := r.URL.Query()
	q := dashboardListQuery{
		Type: v.Get("type"), Genre: v.Get("genre"), Tag: v.Get("tag"),
		Search: strings.ToLower(v.Get("q")), Sort: v.Get("sort"),
	}
	q.MinRating, _ = strconv.ParseFloat(v.Get("min_rating"), 64)
	q.YearFrom, _ = strconv.Atoi(v.Get("year_from"))
	q.YearTo, _ = strconv.Atoi(v.Get("year_to"))
	q.Limit = parseLimitDefault(v.Get("limit"), 200, 2000)
	q.Offset, _ = strconv.Atoi(v.Get("offset"))
	return q
}

func parseLimitDefault(raw string, def, ceiling int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > ceiling {
		return ceiling
	}
	return n
}

func buildDashboardCards(database *db.DB, items []db.Media) []dashboardCard {
	out := make([]dashboardCard, 0, len(items))
	for i := range items {
		out = append(out, mediaToCard(database, &items[i]))
	}
	return out
}

func mediaToCard(database *db.DB, m *db.Media) dashboardCard {
	tags, _ := database.GetTagsByMediaID(int(m.ID))
	return dashboardCard{
		ID: m.ID, Title: m.Title, Year: m.Year, Type: m.Type, Runtime: m.Runtime,
		TmdbID: m.TmdbID, TmdbRating: m.TmdbRating, Genre: m.Genre,
		GenreList: splitGenres(m.Genre), Director: m.Director, CastList: m.CastList,
		Description: m.Description, Tagline: m.Tagline,
		ThumbnailPath: m.ThumbnailPath, Tags: tags, Watched: containsTag(tags, "watched"),
	}
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func applyDashboardFilters(cards []dashboardCard, q dashboardListQuery) []dashboardCard {
	out := cards[:0:0]
	for i := range cards {
		if cardMatches(&cards[i], q) {
			out = append(out, cards[i])
		}
	}
	return out
}

func cardMatches(c *dashboardCard, q dashboardListQuery) bool {
	if !typeMatches(c, q.Type) {
		return false
	}
	if q.Genre != "" && !strings.Contains(strings.ToLower(c.Genre), strings.ToLower(q.Genre)) {
		return false
	}
	if q.Tag != "" && !containsTag(c.Tags, q.Tag) {
		return false
	}
	if q.MinRating > 0 && c.TmdbRating < q.MinRating {
		return false
	}
	return yearMatches(c, q) && searchMatches(c, q.Search)
}

func typeMatches(c *dashboardCard, want string) bool {
	if want == "" || want == "all" {
		return true
	}
	if want == "watched" {
		return c.Watched
	}
	return c.Type == want
}

func yearMatches(c *dashboardCard, q dashboardListQuery) bool {
	if q.YearFrom > 0 && c.Year < q.YearFrom {
		return false
	}
	if q.YearTo > 0 && c.Year > q.YearTo {
		return false
	}
	return true
}

func searchMatches(c *dashboardCard, q string) bool {
	if q == "" {
		return true
	}
	hay := strings.ToLower(c.Title + " " + c.Director + " " + c.CastList)
	return strings.Contains(hay, q)
}

func sortDashboardCards(cards []dashboardCard, mode string) {
	sort.SliceStable(cards, func(i, j int) bool {
		return cardLess(&cards[i], &cards[j], mode)
	})
}

func cardLess(a, b *dashboardCard, mode string) bool {
	switch mode {
	case "title-desc":
		return a.Title > b.Title
	case "rating-desc":
		return a.TmdbRating > b.TmdbRating
	case "rating-asc":
		return a.TmdbRating < b.TmdbRating
	case "year-desc":
		return a.Year > b.Year
	case "year-asc":
		return a.Year < b.Year
	}
	return a.Title < b.Title
}

func paginateDashboardCards(cards []dashboardCard, q dashboardListQuery) dashboardListResponse {
	total := len(cards)
	start := q.Offset
	if start > total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}
	return dashboardListResponse{
		Items: cards[start:end], Total: total, Limit: q.Limit, Offset: q.Offset,
	}
}

// ----- /api/media/{id}/details ---------------------------------------------
