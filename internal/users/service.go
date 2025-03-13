package users

import (
	"context"
	"time"

	"github.com/bluele/gcache"
)

type Service struct {
	repo  *Repository
	cache gcache.Cache
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:  repo,
		cache: gcache.New(100).LRU().Expiration(time.Minute * 5).Build(),
	}
}

func (s *Service) GetUserInfos(ctx context.Context, id string) (*UserRequest, error) {
	if cached, err := s.cache.Get(id); err == nil {
		if user, ok := cached.(*UserRequest); ok {
			return user, nil
		}
		return nil, err
	}

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetWithExpire(id, user, time.Hour); err != nil {
		return nil, err
	}

	return user, nil
}
