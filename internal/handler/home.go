package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type HomeHandler struct {
	articles        *service.ArticleService
	categories      *service.CategoryService
	recommendations *service.RecommendationService
}

func NewHomeHandler(
	articles *service.ArticleService,
	categories *service.CategoryService,
	recommendations *service.RecommendationService,
) *HomeHandler {
	return &HomeHandler{
		articles:        articles,
		categories:      categories,
		recommendations: recommendations,
	}
}

// Home returns aggregated data for the homepage in a single request.
func (h *HomeHandler) Home(c *gin.Context) {
	req := service.ListArticleReq{
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 8),
		Status:   "published",
	}

	type homeResponse struct {
		Articles       interface{} `json:"articles"`
		Categories     interface{} `json:"categories"`
		TrendingTopics interface{} `json:"trending_topics"`
		RecommendedAuthors interface{} `json:"recommended_authors"`
		WeeklyPicks    interface{} `json:"weekly_picks"`
	}

	var (
		resp homeResponse
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex
	)

	addErr := func(err error) {
		mu.Lock()
		errs = append(errs, err)
		mu.Unlock()
	}

	wg.Add(4)
	go func() {
		defer wg.Done()
		result, err := h.articles.ListArticles(c.Request.Context(), req)
		if err != nil {
			addErr(err)
			return
		}
		resp.Articles = result
	}()
	go func() {
		defer wg.Done()
		categories, err := h.categories.List(c.Request.Context())
		if err != nil {
			addErr(err)
			return
		}
		resp.Categories = categories
	}()
	go func() {
		defer wg.Done()
		topics, err := h.recommendations.GetTrendingTags(c.Request.Context(), 12)
		if err != nil {
			addErr(err)
			return
		}
		resp.TrendingTopics = topics
	}()
	go func() {
		defer wg.Done()
		authors, err := h.recommendations.GetRecommendedAuthors(c.Request.Context(), 5)
		if err != nil {
			addErr(err)
			return
		}
		resp.RecommendedAuthors = authors
	}()
	wg.Wait()

	// weekly picks as a separate sync call (uses article service)
	picks, err := h.articles.Ranking(c.Request.Context(), "week", 5)
	if err != nil {
		// non-fatal
	} else {
		resp.WeeklyPicks = picks
	}

	if len(errs) > 0 && resp.Articles == nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}

	app.Success(c, resp)
}
