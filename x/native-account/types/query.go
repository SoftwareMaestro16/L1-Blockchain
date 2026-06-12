package types

import (
	"encoding/json"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type QueryAccountRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

type QueryAccountResponse struct {
	Found		bool	`protobuf:"varint,1,opt,name=found,proto3" json:"found,omitempty"`
	Virtual		bool	`protobuf:"varint,2,opt,name=virtual,proto3" json:"virtual,omitempty"`
	AddressUser	string	`protobuf:"bytes,3,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw	string	`protobuf:"bytes,4,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status		string	`protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
	AccountJSON	string	`protobuf:"bytes,6,opt,name=account_json,json=accountJSON,proto3" json:"account_json,omitempty"`
}

type QueryAccountByRawRequest struct {
	AddressRaw string `protobuf:"bytes,1,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
}

type QueryVirtualAccountRequest struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
}

type QueryVirtualAccountResponse struct {
	AddressUser		string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw		string	`protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status			string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Persistent		bool	`protobuf:"varint,4,opt,name=persistent,proto3" json:"persistent,omitempty"`
	StorageRentActive	bool	`protobuf:"varint,5,opt,name=storage_rent_active,json=storageRentActive,proto3" json:"storage_rent_active,omitempty"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	DefaultQueryLimit	uint64	`protobuf:"varint,1,opt,name=default_query_limit,json=defaultQueryLimit,proto3" json:"default_query_limit,omitempty"`
	MaxQueryLimit		uint64	`protobuf:"varint,2,opt,name=max_query_limit,json=maxQueryLimit,proto3" json:"max_query_limit,omitempty"`
}

type QueryAccountStatusRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

type QueryAccountStatusResponse struct {
	AddressUser		string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw		string	`protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status			string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Persistent		bool	`protobuf:"varint,4,opt,name=persistent,proto3" json:"persistent,omitempty"`
	StorageRentActive	bool	`protobuf:"varint,5,opt,name=storage_rent_active,json=storageRentActive,proto3" json:"storage_rent_active,omitempty"`
	StorageRentDebt		uint64	`protobuf:"varint,6,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
}

func NewQueryAccountResponse(account Account, found, virtual bool) QueryAccountResponse {
	resp := QueryAccountResponse{Found: found, Virtual: virtual}
	if found {
		resp.AddressUser = account.AddressUser
		resp.AddressRaw = account.AddressRaw
		resp.Status = account.Status
		if bz, err := json.Marshal(account); err == nil {
			resp.AccountJSON = string(bz)
		}
	}
	return resp
}

func NewQueryVirtualAccountResponse(view VirtualAccountView) QueryVirtualAccountResponse {
	return QueryVirtualAccountResponse{
		AddressUser:		view.AddressUser,
		AddressRaw:		view.AddressRaw,
		Status:			view.Status,
		Persistent:		view.Persistent,
		StorageRentActive:	view.StorageRentActive,
	}
}

func NewQueryParamsResponse(params prototype.Params) QueryParamsResponse {
	return QueryParamsResponse{DefaultQueryLimit: params.DefaultQueryLimit, MaxQueryLimit: params.MaxQueryLimit}
}
