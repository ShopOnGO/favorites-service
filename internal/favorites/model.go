package favorites

import "gorm.io/gorm"

type Favorite struct {
	gorm.Model
	UserID    			uint `gorm:"not null;index" json:"user_id"`
	ProductVariantID 	uint `gorm:"not null;index" json:"product_id"`
}
