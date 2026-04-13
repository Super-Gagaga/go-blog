package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/slug"
	"github.com/yourname/go-blog/internal/repository"
)

var ErrCategoryExists = errors.New("category already exists")
var ErrCategoryNotFound = errors.New("category not found")

type CategoryService struct {
	categories *repository.CategoryRepository
}

type CreateCategoryReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateCategoryReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	return s.categories.List()
}

func (s *CategoryService) Create(ctx context.Context, req CreateCategoryReq) (*model.Category, error) {
	name := strings.TrimSpace(req.Name)
	categorySlug := slug.Make(name)

	existing, err := s.categories.FindBySlug(categorySlug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCategoryExists
	}

	category := &model.Category{Name: name, Slug: categorySlug}
	if err := s.categories.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id int64, req UpdateCategoryReq) (*model.Category, error) {
	category, err := s.categories.FindByID(id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}

	name := strings.TrimSpace(req.Name)
	categorySlug := slug.Make(name)
	existing, err := s.categories.FindBySlug(categorySlug)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, ErrCategoryExists
	}

	category.Name = name
	category.Slug = categorySlug
	if err := s.categories.Update(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	category, err := s.categories.FindByID(id)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrCategoryNotFound
	}
	return s.categories.Delete(id)
}
