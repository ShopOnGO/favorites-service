package favorites

import (
	"errors"

	"github.com/ShopOnGO/ShopOnGO/pkg/db"

	"gorm.io/gorm"
)

type FavoriteRepository struct {
	Db *db.Db
}

func NewFavoriteRepository(db *db.Db) *FavoriteRepository {
	return &FavoriteRepository{
		Db: db,
	}
}

func (r *FavoriteRepository) Add(userID, variantID uint) error {
	fav := &Favorite{UserID: userID, ProductVariantID: variantID}
	err := r.Db.Create(fav).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (r *FavoriteRepository) Delete(userID, variantID uint) error {
	err := r.Db.Where("user_id = ? AND product_variant_id = ?", userID, variantID).
		Delete(&Favorite{}).
		Error
	return err
}

func (r *FavoriteRepository) ListByUser(userID uint) ([]Favorite, error) {
	var list []Favorite
	err := r.Db.Where("user_id = ?", userID).Find(&list).Error
	return list, err
}