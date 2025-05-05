package favorites

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	pb "github.com/ShopOnGO/product-proto/pkg/product"
	"google.golang.org/grpc"

	"github.com/gorilla/mux"
)

type FavoriteHandlerDeps struct {
	Config          		*configs.Config
	FavoriteService 		*FavoriteService
}

type FavoriteHandler struct {
	Config          		*configs.Config
	FavoriteService 		*FavoriteService
	Client  				*GRPCClient
}

type GRPCClient struct {
	ProductVariantClient pb.ProductVariantServiceClient
}

func InitGRPCClient() *GRPCClient {
	conn, err := grpc.Dial("product_container:50053", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Errorf("Ошибка подключения к gRPC серверу: %v", err)
	}

	logger.Info("gRPC connected")
	ProductVariantClient := pb.NewProductVariantServiceClient(conn)

	return &GRPCClient{
		ProductVariantClient: ProductVariantClient,
	}
}


func NewFavoriteHandler(router *mux.Router, deps FavoriteHandlerDeps) {
	handler := &FavoriteHandler{
		FavoriteService: deps.FavoriteService,
		Client:  InitGRPCClient(),
	}

	router.Handle("/favorites/{product_variant_id}", middleware.IsAuthed(handler.AddFavorite(), deps.Config)).Methods("POST")
	router.Handle("/favorites/{product_variant_id}", middleware.IsAuthed(handler.DeleteFavorite(), deps.Config)).Methods("DELETE")
	router.Handle("/favorites", middleware.IsAuthed(handler.ListFavorites(), deps.Config)).Methods("GET")
}

// AddFavorite добавляет товар в избранное.
// @Summary      Добавление товара в избранное
// @Description  Добавляет указанный продукт (product_variant_id) в список избранного для авторизованного пользователя. Перед добавлением проверяется, что вариант продукта существует и активен.
// @Tags         favorites
// @Accept       json
// @Produce      json
// @Param        product_variant_id  path    uint64  true  "ID варианта товара"
// @Success      200  {object}  map[string]string  "status: added to favorites"
// @Failure      400  {string}  string  "invalid product_variant_id or user_id"
// @Failure      404  {string}  string  "product variant does not exist or inactive"
// @Failure      500  {string}  string  "failed to add to favorites"
// @Security     ApiKeyAuth
// @Router       /favorites/{product_variant_id} [post]
func (h *FavoriteHandler) AddFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["product_variant_id"]
		productVariantID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || productVariantID == 0 {
			http.Error(w, "invalid product_variant_id", http.StatusBadRequest)
			return
		}
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		if productVariantID == 0 || userID == 0 {
			http.Error(w, "product_variant_id and user_id are required", http.StatusBadRequest)
			return
		}

		checkResp, err := h.Client.ProductVariantClient.CheckProductVariantExists(context.Background(), &pb.CheckProductVariantRequest{
			ProductVariantId: uint32(productVariantID),
		})
		if err != nil {
			http.Error(w, "failed to verify product variant", http.StatusInternalServerError)
			return
		}
		if !checkResp.Exists || !checkResp.IsActive {
			http.Error(w, "product variant does not exist or inactive", http.StatusNotFound)
			return
		}

		if err := h.FavoriteService.AddFavorite(uint(userID), uint(productVariantID)); err != nil {
			http.Error(w, "failed to add to favorites: "+err.Error(), http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "added to favorites"}, http.StatusOK)
	}
}

// DeleteFavorite удаляет товар из избранного.
// @Summary      Удаление товара из избранного
// @Description  Удаляет указанный продукт (product_variant_id) из списка избранного для авторизованного пользователя.
// @Tags         favorites
// @Accept       json
// @Produce      json
// @Param        product_variant_id  path    uint64  true  "ID варианта товара"
// @Success      200  {object}  map[string]string  "status: removed from favorites"
// @Failure      400  {string}  string  "invalid product_variant_id or user_id"
// @Failure      500  {string}  string  "failed to remove from favorites"
// @Security     ApiKeyAuth
// @Router       /favorites/{product_variant_id} [delete]
func (h *FavoriteHandler) DeleteFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		if userID == 0 {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["product_variant_id"]
		productVariantID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || productVariantID == 0 {
			http.Error(w, "invalid product_variant_id", http.StatusBadRequest)
			return
		}

		if err := h.FavoriteService.DeleteFavorite(uint(userID), uint(productVariantID)); err != nil {
			http.Error(w, "failed to remove from favorites: "+err.Error(), http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "removed from favorites"}, http.StatusOK)
	}
}

// ListFavorites возвращает список избранного.
// @Summary      Получение списка избранных товаров
// @Description  Возвращает все варианты продуктов, добавленные в избранное авторизованным пользователем.
// @Tags         favorites
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "favorites: array of ProductVariant"
// @Failure      400  {string}  string  "invalid user_id"
// @Failure      500  {string}  string  "failed to get favorites"
// @Security     ApiKeyAuth
// @Router       /favorites [get]
func (h *FavoriteHandler) ListFavorites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		if userID == 0 {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		favorites, err := h.FavoriteService.ListFavorites(uint(userID))
		if err != nil {
			http.Error(w, "failed to get favorites: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var variantIDs []uint32
		for _, fav := range favorites {
			variantIDs = append(variantIDs, uint32(fav.ProductVariantID))
		}

		productsResp, err := h.Client.ProductVariantClient.GetProductVariants(context.Background(), &pb.GetProductVariantsRequest{
			ProductVariantIds: variantIDs,
		})
		if err != nil {
			http.Error(w, "failed to fetch product data", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]interface{}{
			"favorites": productsResp.ProductVariants,
		}, http.StatusOK)
	}
}
