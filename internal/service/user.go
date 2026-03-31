package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/go-blog/internal/config"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/cache"
	"github.com/yourname/go-blog/internal/pkg/cachemanager"
	"github.com/yourname/go-blog/internal/pkg/hash"
	jwtpkg "github.com/yourname/go-blog/internal/pkg/jwt"
	"github.com/yourname/go-blog/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidRefresh    = errors.New("invalid refresh token")
	ErrUnsupportedAvatar = errors.New("unsupported avatar type")
	ErrAvatarTooLarge    = errors.New("avatar too large")
	ErrUserBanned        = errors.New("user is banned")
)

const userProfileCacheTTL = 10 * time.Minute

type UserService struct {
	users *repository.UserRepository
	redis *redis.Client
	jwt   *jwtpkg.Manager
	cfg   *config.Config
	// userCache 用户资料的缓存旁路;Password 字段 json:"-",不会进入缓存
	userCache *cachemanager.ServiceManager[model.User]
}

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileReq struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Bio      *string `json:"bio" binding:"omitempty,max=500"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewUserService(users *repository.UserRepository, redis *redis.Client, jwt *jwtpkg.Manager, cfg *config.Config, userCache *cachemanager.ServiceManager[model.User]) *UserService {
	return &UserService{
		users:     users,
		redis:     redis,
		jwt:       jwt,
		cfg:       cfg,
		userCache: userCache,
	}
}

func (s *UserService) Register(ctx context.Context, req RegisterReq) (*model.UserProfile, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	existing, err := s.users.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	existing, err = s.users.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		Status:   "active",
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	profile := model.NewUserProfile(user, true)
	return &profile, nil
}

func (s *UserService) Login(ctx context.Context, req LoginReq) (*TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil || !hash.CheckPassword(user.Password, req.Password) {
		return nil, ErrInvalidCredential
	}
	if user.Status == "banned" {
		return nil, ErrUserBanned
	}

	return s.issueTokens(ctx, user.ID)
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	key := refreshTokenKey(refreshToken)
	userID, err := s.redis.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) || userID != claims.UserID {
		return nil, ErrInvalidRefresh
	}
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, claims.UserID)
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*model.UserProfile, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	profile := model.NewUserProfile(user, true)
	return &profile, nil
}

func (s *UserService) GetPublicProfile(ctx context.Context, userID int64) (*model.UserProfile, error) {
	// 公开资料是热点读路径,走缓存旁路:未命中回源 DB 并异步回填
	user, err := s.userCache.LookupSingleWithFallback(
		ctx,
		cache.UserCacheKey(userID),
		func(db *gorm.DB) *gorm.DB { return db },
		userProfileCacheTTL,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	profile := model.NewUserProfile(user, false)
	return &profile, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileReq) (*model.UserProfile, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username != "" && username != user.Username {
			existing, err := s.users.FindByUsername(username)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != userID {
				return nil, ErrUserExists
			}
			user.Username = username
		}
	}
	if req.Bio != nil {
		user.Bio = strings.TrimSpace(*req.Bio)
	}

	if err := s.users.Update(user); err != nil {
		return nil, err
	}
	s.invalidateUserCache(ctx, userID)
	profile := model.NewUserProfile(user, true)
	return &profile, nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, filename string, size int64) (string, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	if size > 5*1024*1024 {
		return "", ErrAvatarTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", ErrUnsupportedAvatar
	}

	dir := filepath.Join("uploads", "avatars")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	path := filepath.Join(dir, name)
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	url := "/" + filepath.ToSlash(path)
	user.Avatar = url
	if err := s.users.Update(user); err != nil {
		return "", err
	}
	s.invalidateUserCache(ctx, userID)

	return url, nil
}

func (s *UserService) invalidateUserCache(ctx context.Context, userID int64) {
	_ = s.userCache.InvalidateSingleCache(ctx, cache.UserCacheKey(userID))
}

func (s *UserService) issueTokens(ctx context.Context, userID int64) (*TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	key := refreshTokenKey(refreshToken)
	if err := s.redis.Set(ctx, key, strconv.FormatInt(userID, 10), s.cfg.JWT.RefreshTTL).Err(); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

func refreshTokenKey(token string) string {
	return "auth:refresh:" + token
}
