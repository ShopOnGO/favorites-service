package favorites

import (
	"fmt"
)

type FavoriteService struct {
	FavoriteRepository *FavoriteRepository
}

func NewFavoriteService(FavoriteRepo *FavoriteRepository) *FavoriteService {
	return &FavoriteService{FavoriteRepository: FavoriteRepo}
}

func (s *FavoriteService) AddFavorite(userID, variantID uint) error {
	if userID == 0 || variantID == 0 {
		return fmt.Errorf("invalid user_id or variant_id")
	}
	return s.FavoriteRepository.Add(userID, variantID)
}

func (s *FavoriteService) DeleteFavorite(userID, variantID uint) error {
	if userID == 0 || variantID == 0 {
		return fmt.Errorf("invalid user_id or variant_id")
	}
	return s.FavoriteRepository.Delete(userID, variantID)
}

func (s *FavoriteService) ListFavorites(userID uint) ([]Favorite, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user_id")
	}
	return s.FavoriteRepository.ListByUser(userID)
}
