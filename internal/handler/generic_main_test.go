// Package handler 的通用路由测试:缓存类路由跑 miniredis,数据库类路由跑
// sqlmock 幽灵库(经 cachemanager.SetGlobalDB 注入)。
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var testMR *miniredis.Miniredis

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	mr, err := miniredis.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "miniredis.Run failed: %v\n", err)
		os.Exit(1)
	}
	testMR = mr

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cachemanager.SetGlobalRedis(client)

	code := m.Run()

	client.Close()
	mr.Close()
	cachemanager.SetGlobalRedis(nil)
	os.Exit(code)
}

// newSM builds a ServiceManager over the shared test model.
func newSM() *cachemanager.ServiceManager[testutil.TestUser] {
	return cachemanager.NewServiceManager(testutil.TestUser{})
}

// injectDB swaps the global DB for a fresh sqlmock-backed gorm instance and
// returns the mock so the test can program expectations.
func injectDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		// 测试中的 gorm 实例同样走 slog 适配器,保证输出可控
		Logger: cachemanager.NewSlogGormLogger(0),
	})
	require.NoError(t, err)
	cachemanager.SetGlobalDB(gormDB)
	t.Cleanup(func() { sqlDB.Close() })
	return mock
}

func flushRedis() {
	testMR.FlushAll()
}

// seedCache writes raw key→JSON entries straight into miniredis.
func seedCache(data map[string]string) {
	ctx := context.Background()
	rdb := cachemanager.GetRedis()
	for k, v := range data {
		if err := rdb.Set(ctx, k, v, 0).Err(); err != nil {
			panic(fmt.Sprintf("seedCache %s: %v", k, err))
		}
	}
}

// doJSON performs a JSON request and decodes the JSON response body.
func doJSON(t *testing.T, r *gin.Engine, method, path string, body interface{}) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, decodeBody(w)
}

func doGet(t *testing.T, r *gin.Engine, path string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, decodeBody(w)
}

func decodeBody(w *httptest.ResponseRecorder) map[string]interface{} {
	resp := map[string]interface{}{}
	if w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
	}
	return resp
}

// userRow builds one sqlmock row shaped for testutil.TestUser.
func userRows(rows ...testutil.TestUser) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "username", "email", "age", "status"})
	for _, u := range rows {
		r.AddRow(u.ID, u.Username, u.Email, u.Age, u.Status)
	}
	return r
}

func countRows(n int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count(*)"}).AddRow(n)
}

// ==================== 路由引擎构造 ====================

func newWriteRouter() *gin.Engine {
	r := gin.New()
	g := r.Group("/api/v1/users")
	NewWriteRouterGroup(g, newSM()).RegisterRoutes("")
	return r
}

func newQueryRouter() *gin.Engine {
	r := gin.New()
	g := r.Group("/api/v1/users")
	qrg := NewQueryRouterGroup(g, newSM())
	qrg.RegisterCommonMethods(20)
	qrg.RegisterRoutes("")
	return r
}

func newWritedownRouter() *gin.Engine {
	r := gin.New()
	g := r.Group("/api/v1/users")
	NewWritedownRouterGroup(g, newSM()).RegisterRoutes("/cache")
	return r
}

func newLookupRouter() *gin.Engine {
	r := gin.New()
	g := r.Group("/api/v1/users")
	lrg := NewLookupRouterGroup(g, newSM())
	lrg.SetDefaults("user:*", time.Hour)
	lrg.RegisterRoutes("/lookup")
	return r
}
