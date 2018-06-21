package types

import (
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/logic/model"
)

type (
	// IsGrayArgs is_gray API parameters
	IsGrayArgs struct {
		Uri string `param:"<nozero> <rerr:400:Uri can not be empty>" protobuf:"bytes,1,opt,name=Uri,proto3" json:"Uri,omitempty"`
		Uid string `param:"<nozero> <rerr:400:Uid can not be empty>" protobuf:"bytes,2,opt,name=Uid,proto3" json:"Uid,omitempty"`
	}

	// IsGrayResult is_gray API result
	IsGrayResult struct {
		Gray bool `protobuf:"varint,1,opt,name=Gray,proto3" json:"Gray,omitempty"`
	}

	// GetArgs get API parameters
	GetArgs struct {
		Uri string `param:"<nonzero> <rerr:400:Uri can not be empty>" protobuf:"bytes,1,opt,name=Uri,proto3" json:"Uri,omitempty"`
	}

	// DeleteArgs delete API parameters
	DeleteArgs struct {
		Uri string `param:"<nonzero> <rerr:400:Uri can not be empty>" protobuf:"bytes,1,opt,name=Uri,proto3" json:"Uri,omitempty"`
	}

	// SetArgs set API parameters
	SetArgs struct {
		Uri    string `param:"<nonzero> <rerr:400:Uri can not be empty>" protobuf:"bytes,1,opt,name=Uri,proto3" json:"uri"`
		Regexp string `protobuf:"bytes,2,opt,name=Regexp,proto3" json:"regexp"`
	}
)

// GrayMatch
type GrayMatch = model.GrayMatch
