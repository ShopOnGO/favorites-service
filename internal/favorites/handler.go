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
