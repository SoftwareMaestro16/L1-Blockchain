package prototype

import (
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultAuthority	= "4:0000000000000000000000000000000000000000000000000000000000000001"
	DefaultQueryLimit	= uint64(50)
	MaxQueryLimit		= uint64(200)
	DefaultVersionGate	= "prototype-v1"
	NextMigrationVersion	= uint64(2)
	CurrentGenesisVersion	= uint64(1)
)

type Params struct {
	Enabled			bool
	TestnetProfile		bool
	ProductionVersionGate	string
	Authority		string
	DefaultQueryLimit	uint64
	MaxQueryLimit		uint64
}

type PageRequest struct {
	Offset	uint64
	Limit	uint64
}

type PageResponse struct {
	NextOffset uint64
}

func DefaultParams() Params {
	return Params{
		Enabled:		false,
		Authority:		DefaultAuthority,
		DefaultQueryLimit:	DefaultQueryLimit,
		MaxQueryLimit:		MaxQueryLimit,
	}
}

func TestnetParams() Params {
	params := DefaultParams()
	params.Enabled = true
	params.TestnetProfile = true
	return params
}

func ProductionEnabledParams(versionGate string) Params {
	params := DefaultParams()
	params.Enabled = true
	params.ProductionVersionGate = versionGate
	return params
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("prototype authority", p.Authority); err != nil {
		return err
	}
	if p.DefaultQueryLimit == 0 {
		return errors.New("prototype default query limit must be positive")
	}
	if p.MaxQueryLimit == 0 {
		return errors.New("prototype max query limit must be positive")
	}
	if p.DefaultQueryLimit > p.MaxQueryLimit {
		return errors.New("prototype default query limit must not exceed max query limit")
	}
	if p.MaxQueryLimit > MaxQueryLimit {
		return fmt.Errorf("prototype max query limit must be <= %d", MaxQueryLimit)
	}
	if p.Enabled && !p.TestnetProfile && p.ProductionVersionGate == "" {
		return errors.New("prototype production enablement requires software version gate")
	}
	return nil
}

func (p Params) RequireEnabled() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Enabled {
		return errors.New("prototype feature gate is disabled")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("prototype update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("prototype update requires governance authority")
	}
	return nil
}

func NormalizePage(req *PageRequest, params Params, total int) (start int, end int, res PageResponse, err error) {
	if err := params.Validate(); err != nil {
		return 0, 0, PageResponse{}, err
	}
	if req == nil {
		req = &PageRequest{}
	}
	limit := req.Limit
	if limit == 0 {
		limit = params.DefaultQueryLimit
	}
	if limit == 0 || limit > params.MaxQueryLimit {
		return 0, 0, PageResponse{}, errors.New("prototype query limit out of bounds")
	}
	if req.Offset > uint64(total) {
		return 0, 0, PageResponse{}, errors.New("prototype query offset out of bounds")
	}
	start = int(req.Offset)
	end64 := req.Offset + limit
	if end64 < req.Offset {
		return 0, 0, PageResponse{}, errors.New("prototype query offset overflow")
	}
	if end64 > uint64(total) {
		end64 = uint64(total)
	}
	end = int(end64)
	if end < total {
		res.NextOffset = uint64(end)
	}
	return start, end, res, nil
}
