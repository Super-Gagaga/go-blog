package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/repository"
)

var ErrTagExists = errors.New("tag already exists")
var ErrTagNotFound = errors.New("tag not found")

type TagService struct {
	tags *repository.TagRepository
}

type CreateTagReq struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type UpdateTagReq struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

func NewTagService(tags *repository.TagRepository) *TagService {
	return &TagService{tags: tags}
}

func (s *TagService) List(ctx context.Context, keyword string) ([]model.Tag, error) {
	return s.tags.List(strings.TrimSpace(keyword))
}

func (s *TagService) Create(ctx context.Context, req CreateTagReq) (*model.Tag, error) {
	name := strings.TrimSpace(req.Name)
	existing, err := s.tags.FindByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTagExists
	}

	tag := &model.Tag{Name: name}
	if err := s.tags.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Update(ctx context.Context, id int64, req UpdateTagReq) (*model.Tag, error) {
	tag, err := s.tags.FindByID(id)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, ErrTagNotFound
	}

	name := strings.TrimSpace(req.Name)
	existing, err := s.tags.FindByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, ErrTagExists
	}

	tag.Name = name
	if err := s.tags.Update(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Delete(ctx context.Context, id int64) error {
	tag, err := s.tags.FindByID(id)
	if err != nil {
		return err
	}
	if tag == nil {
		return ErrTagNotFound
	}
	return s.tags.Delete(id)
}
