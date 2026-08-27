package service

import (
	"encoding/base64"
	"strconv"

	productv1 "github.com/sartim/storemesh-product-service/gen/storemesh/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func pageParameters(req *productv1.ListProductsRequest) (int, int, error) {
	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	if pageSize < 0 || pageSize > maxPageSize {
		return 0, 0, status.Errorf(codes.InvalidArgument, "page_size must be between 1 and %d", maxPageSize)
	}

	offset := 0
	if token := req.GetPageToken(); token != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			return 0, 0, status.Error(codes.InvalidArgument, "page_token is invalid")
		}
		offset, err = strconv.Atoi(string(decoded))
		if err != nil || offset < 0 {
			return 0, 0, status.Error(codes.InvalidArgument, "page_token is invalid")
		}
	}
	return pageSize, offset, nil
}

func nextPageToken(offset, pageSize, total int) string {
	next := offset + pageSize
	if next >= total {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(next)))
}
