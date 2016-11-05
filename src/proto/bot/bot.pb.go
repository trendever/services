// Code generated by protoc-gen-go.
// source: bot.proto
// DO NOT EDIT!

/*
Package bot is a generated protocol buffer package.

It is generated from these files:
	bot.proto

It has these top-level messages:
	RetrieveActivitiesRequest
	SendDirectRequest
	RetrieveActivitiesReply
	SendDirectReply
	Activity
	DirectMessageNotify
	SaveProductResult
	NotifyMessageRequest
	NotifyMessageResult
*/
package bot

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RetrieveActivitiesRequest struct {
	MentionName string   `protobuf:"bytes,1,opt,name=mention_name,json=mentionName" json:"mention_name,omitempty"`
	AfterId     int64    `protobuf:"varint,2,opt,name=after_id,json=afterId" json:"after_id,omitempty"`
	Type        []string `protobuf:"bytes,4,rep,name=type" json:"type,omitempty"`
	Limit       int64    `protobuf:"varint,5,opt,name=limit" json:"limit,omitempty"`
}

func (m *RetrieveActivitiesRequest) Reset()                    { *m = RetrieveActivitiesRequest{} }
func (m *RetrieveActivitiesRequest) String() string            { return proto.CompactTextString(m) }
func (*RetrieveActivitiesRequest) ProtoMessage()               {}
func (*RetrieveActivitiesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type SendDirectRequest struct {
	ActivityPk string `protobuf:"bytes,1,opt,name=activity_pk,json=activityPk" json:"activity_pk,omitempty"`
	Text       string `protobuf:"bytes,2,opt,name=text" json:"text,omitempty"`
}

func (m *SendDirectRequest) Reset()                    { *m = SendDirectRequest{} }
func (m *SendDirectRequest) String() string            { return proto.CompactTextString(m) }
func (*SendDirectRequest) ProtoMessage()               {}
func (*SendDirectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type RetrieveActivitiesReply struct {
	Result []*Activity `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *RetrieveActivitiesReply) Reset()                    { *m = RetrieveActivitiesReply{} }
func (m *RetrieveActivitiesReply) String() string            { return proto.CompactTextString(m) }
func (*RetrieveActivitiesReply) ProtoMessage()               {}
func (*RetrieveActivitiesReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RetrieveActivitiesReply) GetResult() []*Activity {
	if m != nil {
		return m.Result
	}
	return nil
}

type SendDirectReply struct {
}

func (m *SendDirectReply) Reset()                    { *m = SendDirectReply{} }
func (m *SendDirectReply) String() string            { return proto.CompactTextString(m) }
func (*SendDirectReply) ProtoMessage()               {}
func (*SendDirectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Activity struct {
	Id                int64  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Pk                string `protobuf:"bytes,2,opt,name=pk" json:"pk,omitempty"`
	MediaId           string `protobuf:"bytes,3,opt,name=media_id,json=mediaId" json:"media_id,omitempty"`
	MediaUrl          string `protobuf:"bytes,4,opt,name=media_url,json=mediaUrl" json:"media_url,omitempty"`
	UserId            int64  `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	UserName          string `protobuf:"bytes,6,opt,name=user_name,json=userName" json:"user_name,omitempty"`
	UserImageUrl      string `protobuf:"bytes,7,opt,name=user_image_url,json=userImageUrl" json:"user_image_url,omitempty"`
	MentionedUsername string `protobuf:"bytes,8,opt,name=mentioned_username,json=mentionedUsername" json:"mentioned_username,omitempty"`
	Type              string `protobuf:"bytes,9,opt,name=type" json:"type,omitempty"`
	Comment           string `protobuf:"bytes,10,opt,name=comment" json:"comment,omitempty"`
	CreatedAt         int64  `protobuf:"varint,11,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	DirectThreadId    string `protobuf:"bytes,12,opt,name=direct_thread_id,json=directThreadId" json:"direct_thread_id,omitempty"`
}

func (m *Activity) Reset()                    { *m = Activity{} }
func (m *Activity) String() string            { return proto.CompactTextString(m) }
func (*Activity) ProtoMessage()               {}
func (*Activity) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type DirectMessageNotify struct {
	ThreadId  string `protobuf:"bytes,1,opt,name=thread_id,json=threadId" json:"thread_id,omitempty"`
	MessageId string `protobuf:"bytes,2,opt,name=message_id,json=messageId" json:"message_id,omitempty"`
	// instagram id
	UserId uint64 `protobuf:"varint,3,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	Text   string `protobuf:"bytes,4,opt,name=text" json:"text,omitempty"`
	// id of possible related post
	// if it's aviable, it probably mean, that this message was comment to media share
	RelatedMedia string `protobuf:"bytes,5,opt,name=related_media,json=relatedMedia" json:"related_media,omitempty"`
}

func (m *DirectMessageNotify) Reset()                    { *m = DirectMessageNotify{} }
func (m *DirectMessageNotify) String() string            { return proto.CompactTextString(m) }
func (*DirectMessageNotify) ProtoMessage()               {}
func (*DirectMessageNotify) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type SaveProductResult struct {
	Id    int64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Retry bool  `protobuf:"varint,2,opt,name=retry" json:"retry,omitempty"`
}

func (m *SaveProductResult) Reset()                    { *m = SaveProductResult{} }
func (m *SaveProductResult) String() string            { return proto.CompactTextString(m) }
func (*SaveProductResult) ProtoMessage()               {}
func (*SaveProductResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type NotifyMessageRequest struct {
	Channel string `protobuf:"bytes,1,opt,name=channel" json:"channel,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *NotifyMessageRequest) Reset()                    { *m = NotifyMessageRequest{} }
func (m *NotifyMessageRequest) String() string            { return proto.CompactTextString(m) }
func (*NotifyMessageRequest) ProtoMessage()               {}
func (*NotifyMessageRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type NotifyMessageResult struct {
}

func (m *NotifyMessageResult) Reset()                    { *m = NotifyMessageResult{} }
func (m *NotifyMessageResult) String() string            { return proto.CompactTextString(m) }
func (*NotifyMessageResult) ProtoMessage()               {}
func (*NotifyMessageResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func init() {
	proto.RegisterType((*RetrieveActivitiesRequest)(nil), "bot.RetrieveActivitiesRequest")
	proto.RegisterType((*SendDirectRequest)(nil), "bot.SendDirectRequest")
	proto.RegisterType((*RetrieveActivitiesReply)(nil), "bot.RetrieveActivitiesReply")
	proto.RegisterType((*SendDirectReply)(nil), "bot.SendDirectReply")
	proto.RegisterType((*Activity)(nil), "bot.Activity")
	proto.RegisterType((*DirectMessageNotify)(nil), "bot.DirectMessageNotify")
	proto.RegisterType((*SaveProductResult)(nil), "bot.SaveProductResult")
	proto.RegisterType((*NotifyMessageRequest)(nil), "bot.NotifyMessageRequest")
	proto.RegisterType((*NotifyMessageResult)(nil), "bot.NotifyMessageResult")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for FetcherService service

type FetcherServiceClient interface {
	RetrieveActivities(ctx context.Context, in *RetrieveActivitiesRequest, opts ...grpc.CallOption) (*RetrieveActivitiesReply, error)
	SendDirect(ctx context.Context, in *SendDirectRequest, opts ...grpc.CallOption) (*SendDirectReply, error)
}

type fetcherServiceClient struct {
	cc *grpc.ClientConn
}

func NewFetcherServiceClient(cc *grpc.ClientConn) FetcherServiceClient {
	return &fetcherServiceClient{cc}
}

func (c *fetcherServiceClient) RetrieveActivities(ctx context.Context, in *RetrieveActivitiesRequest, opts ...grpc.CallOption) (*RetrieveActivitiesReply, error) {
	out := new(RetrieveActivitiesReply)
	err := grpc.Invoke(ctx, "/bot.FetcherService/RetrieveActivities", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fetcherServiceClient) SendDirect(ctx context.Context, in *SendDirectRequest, opts ...grpc.CallOption) (*SendDirectReply, error) {
	out := new(SendDirectReply)
	err := grpc.Invoke(ctx, "/bot.FetcherService/SendDirect", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for FetcherService service

type FetcherServiceServer interface {
	RetrieveActivities(context.Context, *RetrieveActivitiesRequest) (*RetrieveActivitiesReply, error)
	SendDirect(context.Context, *SendDirectRequest) (*SendDirectReply, error)
}

func RegisterFetcherServiceServer(s *grpc.Server, srv FetcherServiceServer) {
	s.RegisterService(&_FetcherService_serviceDesc, srv)
}

func _FetcherService_RetrieveActivities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveActivitiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FetcherServiceServer).RetrieveActivities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bot.FetcherService/RetrieveActivities",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FetcherServiceServer).RetrieveActivities(ctx, req.(*RetrieveActivitiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FetcherService_SendDirect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendDirectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FetcherServiceServer).SendDirect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bot.FetcherService/SendDirect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FetcherServiceServer).SendDirect(ctx, req.(*SendDirectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FetcherService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bot.FetcherService",
	HandlerType: (*FetcherServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RetrieveActivities",
			Handler:    _FetcherService_RetrieveActivities_Handler,
		},
		{
			MethodName: "SendDirect",
			Handler:    _FetcherService_SendDirect_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for SaveTrendService service

type SaveTrendServiceClient interface {
	SaveProduct(ctx context.Context, in *Activity, opts ...grpc.CallOption) (*SaveProductResult, error)
}

type saveTrendServiceClient struct {
	cc *grpc.ClientConn
}

func NewSaveTrendServiceClient(cc *grpc.ClientConn) SaveTrendServiceClient {
	return &saveTrendServiceClient{cc}
}

func (c *saveTrendServiceClient) SaveProduct(ctx context.Context, in *Activity, opts ...grpc.CallOption) (*SaveProductResult, error) {
	out := new(SaveProductResult)
	err := grpc.Invoke(ctx, "/bot.SaveTrendService/SaveProduct", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SaveTrendService service

type SaveTrendServiceServer interface {
	SaveProduct(context.Context, *Activity) (*SaveProductResult, error)
}

func RegisterSaveTrendServiceServer(s *grpc.Server, srv SaveTrendServiceServer) {
	s.RegisterService(&_SaveTrendService_serviceDesc, srv)
}

func _SaveTrendService_SaveProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Activity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SaveTrendServiceServer).SaveProduct(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bot.SaveTrendService/SaveProduct",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SaveTrendServiceServer).SaveProduct(ctx, req.(*Activity))
	}
	return interceptor(ctx, in, info, handler)
}

var _SaveTrendService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bot.SaveTrendService",
	HandlerType: (*SaveTrendServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SaveProduct",
			Handler:    _SaveTrendService_SaveProduct_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for TelegramService service

type TelegramServiceClient interface {
	NotifyMessage(ctx context.Context, in *NotifyMessageRequest, opts ...grpc.CallOption) (*NotifyMessageResult, error)
}

type telegramServiceClient struct {
	cc *grpc.ClientConn
}

func NewTelegramServiceClient(cc *grpc.ClientConn) TelegramServiceClient {
	return &telegramServiceClient{cc}
}

func (c *telegramServiceClient) NotifyMessage(ctx context.Context, in *NotifyMessageRequest, opts ...grpc.CallOption) (*NotifyMessageResult, error) {
	out := new(NotifyMessageResult)
	err := grpc.Invoke(ctx, "/bot.TelegramService/NotifyMessage", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TelegramService service

type TelegramServiceServer interface {
	NotifyMessage(context.Context, *NotifyMessageRequest) (*NotifyMessageResult, error)
}

func RegisterTelegramServiceServer(s *grpc.Server, srv TelegramServiceServer) {
	s.RegisterService(&_TelegramService_serviceDesc, srv)
}

func _TelegramService_NotifyMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TelegramServiceServer).NotifyMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bot.TelegramService/NotifyMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TelegramServiceServer).NotifyMessage(ctx, req.(*NotifyMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TelegramService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bot.TelegramService",
	HandlerType: (*TelegramServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NotifyMessage",
			Handler:    _TelegramService_NotifyMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("bot.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 649 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0x54, 0x4d, 0x4f, 0x1b, 0x4d,
	0x0c, 0x26, 0x24, 0x24, 0x59, 0x07, 0x02, 0x0c, 0xbc, 0x2f, 0x4b, 0xfa, 0x45, 0xb7, 0xad, 0xc4,
	0xa5, 0x1c, 0xd2, 0x5e, 0x2a, 0xf5, 0x50, 0xa4, 0x0a, 0x41, 0x25, 0x10, 0x5a, 0xc2, 0xa1, 0xa7,
	0x68, 0xc9, 0x1a, 0x18, 0xb1, 0x1f, 0xe9, 0xec, 0x24, 0xea, 0x9e, 0xfb, 0x53, 0xda, 0x9f, 0xd6,
	0x1f, 0xd2, 0xb1, 0x67, 0x96, 0xaf, 0x84, 0xdb, 0xf8, 0x79, 0x6c, 0x8f, 0xfd, 0xd8, 0x33, 0xe0,
	0x5d, 0xe4, 0x7a, 0x6f, 0xac, 0x72, 0x9d, 0x8b, 0xba, 0x39, 0x06, 0xbf, 0x6a, 0xb0, 0x1d, 0xa2,
	0x56, 0x12, 0xa7, 0xb8, 0x3f, 0xd2, 0x72, 0x2a, 0xb5, 0xc4, 0x22, 0xc4, 0x1f, 0x13, 0x2c, 0xb4,
	0x78, 0x0d, 0xcb, 0x29, 0x66, 0x5a, 0xe6, 0xd9, 0x30, 0x8b, 0x52, 0xf4, 0x6b, 0x3b, 0xb5, 0x5d,
	0x2f, 0xec, 0x38, 0xec, 0xc4, 0x40, 0x62, 0x1b, 0xda, 0xd1, 0xa5, 0x46, 0x35, 0x94, 0xb1, 0xbf,
	0x68, 0xe8, 0x7a, 0xd8, 0x62, 0xfb, 0x28, 0x16, 0x02, 0x1a, 0xba, 0x1c, 0xa3, 0xdf, 0xd8, 0xa9,
	0x9b, 0x28, 0x3e, 0x8b, 0x4d, 0x58, 0x4a, 0x64, 0x2a, 0xb5, 0xbf, 0xc4, 0xbe, 0xd6, 0x08, 0x0e,
	0x61, 0xfd, 0x0c, 0xb3, 0xf8, 0xab, 0x54, 0x38, 0xd2, 0xd5, 0xe5, 0xaf, 0xa0, 0x13, 0xd9, 0x8a,
	0xca, 0xe1, 0xf8, 0xc6, 0xdd, 0x0d, 0x15, 0x74, 0x7a, 0xc3, 0xf9, 0xf1, 0xa7, 0xe6, 0x6b, 0x29,
	0xbf, 0x39, 0x07, 0x5f, 0x60, 0x6b, 0x5e, 0x3b, 0xe3, 0xa4, 0x14, 0xef, 0xa0, 0xa9, 0xb0, 0x98,
	0x24, 0xda, 0xa4, 0xaa, 0xef, 0x76, 0xfa, 0x2b, 0x7b, 0xa4, 0x85, 0xf3, 0x2a, 0x43, 0x47, 0x06,
	0xeb, 0xb0, 0x7a, 0xbf, 0x16, 0x13, 0x19, 0xfc, 0x5d, 0x84, 0x76, 0xe5, 0x27, 0xba, 0xb0, 0x68,
	0x5a, 0xad, 0x71, 0xf9, 0xe6, 0x44, 0xb6, 0xa9, 0xce, 0xd6, 0x60, 0x4e, 0x24, 0x48, 0x8a, 0xb1,
	0x8c, 0x48, 0x90, 0x3a, 0xa3, 0x2d, 0xb6, 0x8d, 0x20, 0xcf, 0xc0, 0xb3, 0xd4, 0x44, 0x25, 0x46,
	0x15, 0xe2, 0xac, 0xef, 0xb9, 0x4a, 0xc4, 0x16, 0xb4, 0x26, 0x85, 0xd5, 0xd1, 0x6a, 0xd3, 0x24,
	0xd3, 0x46, 0x31, 0xc1, 0x13, 0x68, 0xda, 0x28, 0x02, 0x58, 0xfe, 0xb7, 0xd0, 0xb5, 0x51, 0x69,
	0x74, 0x85, 0x9c, 0xb7, 0xc5, 0x1e, 0xcb, 0x1c, 0x4c, 0x20, 0xe5, 0x7e, 0x0f, 0xc2, 0xcd, 0x0c,
	0xe3, 0x21, 0x31, 0x9c, 0xab, 0xcd, 0x9e, 0xeb, 0xb7, 0xcc, 0xb9, 0x23, 0x6e, 0x07, 0xe7, 0x39,
	0x61, 0x69, 0x70, 0x3e, 0xb4, 0x46, 0x79, 0x4a, 0xbe, 0x3e, 0xd8, 0xae, 0x9c, 0x29, 0x5e, 0x00,
	0x8c, 0x14, 0x46, 0xda, 0xa4, 0x8e, 0xb4, 0xdf, 0xe1, 0xda, 0x3d, 0x87, 0xec, 0x6b, 0xb1, 0x0b,
	0x6b, 0x31, 0x6b, 0x39, 0xd4, 0xd7, 0x06, 0x8c, 0xa9, 0xc1, 0x65, 0xce, 0xd0, 0xb5, 0xf8, 0x80,
	0xe1, 0xa3, 0x38, 0xf8, 0x5d, 0x83, 0x0d, 0x2b, 0xfb, 0x31, 0x16, 0x85, 0x29, 0xfd, 0x24, 0xd7,
	0xf2, 0xb2, 0x24, 0x01, 0xee, 0x42, 0xed, 0x1a, 0xb4, 0xb5, 0x0b, 0xa2, 0xdb, 0x53, 0xeb, 0x5d,
	0x6d, 0xa0, 0x17, 0x7a, 0x0e, 0x31, 0xf4, 0x3d, 0x55, 0x69, 0x18, 0x8d, 0x5b, 0x55, 0xab, 0xe5,
	0x69, 0xdc, 0x2d, 0x8f, 0x78, 0x03, 0x2b, 0x0a, 0x13, 0xee, 0x84, 0xc7, 0xc2, 0x83, 0x30, 0x5a,
	0x3a, 0xf0, 0x98, 0xb0, 0xe0, 0x93, 0xd9, 0xd5, 0x68, 0x8a, 0xa7, 0x2a, 0x8f, 0x27, 0xb4, 0x20,
	0xb4, 0x34, 0x33, 0x4b, 0x61, 0xd6, 0x5c, 0x99, 0x35, 0x2c, 0xb9, 0xa0, 0x76, 0x68, 0x8d, 0xe0,
	0x1b, 0x6c, 0xda, 0x96, 0x5c, 0x7f, 0xd5, 0xa6, 0x93, 0xb6, 0xd7, 0x51, 0x96, 0x61, 0xe2, 0xda,
	0xab, 0x4c, 0x62, 0x5c, 0x2f, 0xae, 0xb5, 0xca, 0x0c, 0xfe, 0x83, 0x8d, 0x47, 0xb9, 0xa8, 0x90,
	0xfe, 0x9f, 0x1a, 0x74, 0x0f, 0x50, 0x8f, 0xae, 0x51, 0x9d, 0xa1, 0x9a, 0xca, 0x11, 0x8a, 0x01,
	0x88, 0xd9, 0x27, 0x21, 0x5e, 0xf2, 0xf6, 0x3f, 0xf9, 0xf4, 0x7b, 0xcf, 0x9f, 0xe4, 0xe9, 0x45,
	0x2c, 0x88, 0xcf, 0x00, 0x77, 0xcf, 0x44, 0xfc, 0xcf, 0xde, 0x33, 0x6f, 0xb8, 0xb7, 0x39, 0x83,
	0x73, 0x74, 0xff, 0x10, 0xd6, 0x48, 0xc4, 0x81, 0x32, 0x4c, 0x55, 0xe7, 0x47, 0xe8, 0xdc, 0x13,
	0x56, 0x3c, 0x7c, 0x9e, 0x3d, 0x77, 0xc3, 0x63, 0xe5, 0xfb, 0xdf, 0x61, 0x75, 0x80, 0x09, 0x5e,
	0xa9, 0x28, 0xad, 0x12, 0x1d, 0xc0, 0xca, 0x03, 0x69, 0xc4, 0x36, 0xc7, 0xce, 0x93, 0xbe, 0xe7,
	0xcf, 0xa3, 0xf8, 0x1f, 0x58, 0xb8, 0x68, 0xf2, 0x3f, 0xf9, 0xe1, 0x5f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xae, 0x6b, 0x5e, 0x08, 0x34, 0x05, 0x00, 0x00,
}
