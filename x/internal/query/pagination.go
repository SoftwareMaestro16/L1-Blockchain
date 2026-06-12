package query

import (
	"bytes"
	"errors"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
)

var ErrInvalidPagination = errors.New("invalid pagination")

type PageBounds struct {
	Start	[]byte
	End	[]byte
	Limit	int
}

func ForwardPageBounds(req *sdkquery.PageRequest, prefix []byte, defaultLimit, maxLimit uint64) (PageBounds, error) {
	if defaultLimit == 0 || maxLimit == 0 || defaultLimit > maxLimit {
		return PageBounds{}, fmt.Errorf("%w: invalid bounds", ErrInvalidPagination)
	}
	if req == nil {
		req = &sdkquery.PageRequest{}
	}
	if len(req.Key) > 0 && req.Offset > 0 {
		return PageBounds{}, fmt.Errorf("%w: key and offset cannot both be set", ErrInvalidPagination)
	}
	if req.Offset > 0 {
		return PageBounds{}, fmt.Errorf("%w: offset is not supported; use next_key", ErrInvalidPagination)
	}
	if req.CountTotal {
		return PageBounds{}, fmt.Errorf("%w: count_total is not supported", ErrInvalidPagination)
	}
	if req.Reverse {
		return PageBounds{}, fmt.Errorf("%w: reverse is not supported", ErrInvalidPagination)
	}

	limit := req.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		return PageBounds{}, fmt.Errorf("%w: limit %d exceeds max %d", ErrInvalidPagination, limit, maxLimit)
	}

	end := storetypes.PrefixEndBytes(prefix)
	start := prefix
	if len(req.Key) > 0 {
		if !bytes.HasPrefix(req.Key, prefix) {
			return PageBounds{}, fmt.Errorf("%w: key is outside query prefix", ErrInvalidPagination)
		}
		if bytes.Compare(req.Key, end) >= 0 {
			return PageBounds{}, fmt.Errorf("%w: key is outside query prefix", ErrInvalidPagination)
		}
		start = req.Key
	}

	limitInt := int(limit)
	return PageBounds{Start: start, End: end, Limit: limitInt}, nil
}

func PageResponse(nextKey []byte) *sdkquery.PageResponse {
	if len(nextKey) == 0 {
		return &sdkquery.PageResponse{}
	}
	return &sdkquery.PageResponse{NextKey: append([]byte(nil), nextKey...)}
}
