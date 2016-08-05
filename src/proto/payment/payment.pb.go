// Code generated by protoc-gen-go.
// source: payment/payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment/payment.proto

It has these top-level messages:
	CreateOrderRequest
	CreateOrderReply
	BuyOrderRequest
	BuyOrderReply
	PaymentButton
	PaymentNotificationMessage
*/
package payment

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

type Currency int32

const (
	Currency_RUB Currency = 0
	Currency_USD Currency = 1
)

var Currency_name = map[int32]string{
	0: "RUB",
	1: "USD",
}
var Currency_value = map[string]int32{
	"RUB": 0,
	"USD": 1,
}

func (x Currency) String() string {
	return proto.EnumName(Currency_name, int32(x))
}
func (Currency) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Errors int32

const (
	Errors_OK Errors = 0
	// internal errors
	Errors_INVALID_DATA Errors = 1
	Errors_DB_FAILED    Errors = 2
	// external errors
	Errors_INIT_FAILED Errors = 127
	Errors_PAY_FAILED  Errors = 128
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
}
var Errors_value = map[string]int32{
	"OK":           0,
	"INVALID_DATA": 1,
	"DB_FAILED":    2,
	"INIT_FAILED":  127,
	"PAY_FAILED":   128,
}

func (x Errors) String() string {
	return proto.EnumName(Errors_name, int32(x))
}
func (Errors) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CreateOrderRequest struct {
	Amount         uint64   `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency       Currency `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId         uint64   `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	UserId         uint64   `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ConversationId uint64   `protobuf:"varint,5,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	ShopCardNumber string   `protobuf:"bytes,6,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateOrderReply struct {
	Id    uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error Errors `protobuf:"varint,3,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type BuyOrderRequest struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	LeadId uint64 `protobuf:"varint,2,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Ip     string `protobuf:"bytes,3,opt,name=ip" json:"ip,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type BuyOrderReply struct {
	RedirectUrl string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error       Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// chat messages
type PaymentButton struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
}

func (m *PaymentButton) Reset()                    { *m = PaymentButton{} }
func (m *PaymentButton) String() string            { return proto.CompactTextString(m) }
func (*PaymentButton) ProtoMessage()               {}
func (*PaymentButton) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type PaymentNotificationMessage struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
}

func (m *PaymentNotificationMessage) Reset()                    { *m = PaymentNotificationMessage{} }
func (m *PaymentNotificationMessage) String() string            { return proto.CompactTextString(m) }
func (*PaymentNotificationMessage) ProtoMessage()               {}
func (*PaymentNotificationMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func init() {
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*PaymentButton)(nil), "payment.PaymentButton")
	proto.RegisterType((*PaymentNotificationMessage)(nil), "payment.PaymentNotificationMessage")
	proto.RegisterEnum("payment.Currency", Currency_name, Currency_value)
	proto.RegisterEnum("payment.Errors", Errors_name, Errors_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for PaymentService service

type PaymentServiceClient interface {
	CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderReply, error)
	BuyOrder(ctx context.Context, in *BuyOrderRequest, opts ...grpc.CallOption) (*BuyOrderReply, error)
}

type paymentServiceClient struct {
	cc *grpc.ClientConn
}

func NewPaymentServiceClient(cc *grpc.ClientConn) PaymentServiceClient {
	return &paymentServiceClient{cc}
}

func (c *paymentServiceClient) CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderReply, error) {
	out := new(CreateOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/CreateOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) BuyOrder(ctx context.Context, in *BuyOrderRequest, opts ...grpc.CallOption) (*BuyOrderReply, error) {
	out := new(BuyOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/BuyOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PaymentService service

type PaymentServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderReply, error)
	BuyOrder(context.Context, *BuyOrderRequest) (*BuyOrderReply, error)
}

func RegisterPaymentServiceServer(s *grpc.Server, srv PaymentServiceServer) {
	s.RegisterService(&_PaymentService_serviceDesc, srv)
}

func _PaymentService_CreateOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CreateOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/CreateOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CreateOrder(ctx, req.(*CreateOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_BuyOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuyOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).BuyOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/BuyOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).BuyOrder(ctx, req.(*BuyOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PaymentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler:    _PaymentService_CreateOrder_Handler,
		},
		{
			MethodName: "BuyOrder",
			Handler:    _PaymentService_BuyOrder_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("payment/payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 486 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x53, 0xdd, 0x6e, 0xd3, 0x4c,
	0x10, 0xad, 0xdd, 0xc6, 0x49, 0x26, 0x8d, 0xe3, 0x6f, 0xa5, 0xf6, 0x33, 0x81, 0x0b, 0xb0, 0x04,
	0x54, 0x95, 0x28, 0x52, 0x7a, 0x8f, 0xe4, 0xc4, 0x05, 0x59, 0x94, 0xb4, 0x38, 0x0d, 0x52, 0xaf,
	0x2c, 0xd7, 0x1e, 0xc0, 0x52, 0x62, 0x9b, 0xf1, 0xba, 0x52, 0xae, 0xe0, 0x35, 0x78, 0x3b, 0x1e,
	0x85, 0x5d, 0xff, 0x24, 0x0d, 0xb4, 0x5c, 0x79, 0xe7, 0x9c, 0xb3, 0xb3, 0xe7, 0xec, 0xac, 0xe1,
	0x20, 0x0b, 0x56, 0x4b, 0x4c, 0xf8, 0xeb, 0xfa, 0x7b, 0x92, 0x51, 0xca, 0x53, 0xd6, 0xae, 0x4b,
	0xeb, 0x97, 0x02, 0x6c, 0x42, 0x18, 0x70, 0xbc, 0xa0, 0x08, 0xc9, 0xc3, 0x6f, 0x05, 0xe6, 0x9c,
	0x1d, 0x82, 0x16, 0x2c, 0xd3, 0x22, 0xe1, 0xa6, 0xf2, 0x54, 0x39, 0xda, 0xf3, 0xea, 0x8a, 0xbd,
	0x82, 0x4e, 0x58, 0x10, 0x61, 0x12, 0xae, 0x4c, 0x55, 0x30, 0xfa, 0xe8, 0xbf, 0x93, 0xa6, 0xf3,
	0xa4, 0x26, 0xbc, 0xb5, 0x84, 0xfd, 0x0f, 0xed, 0x05, 0x06, 0x91, 0x1f, 0x47, 0xe6, 0x6e, 0xd5,
	0x47, 0x96, 0x6e, 0x24, 0x89, 0x22, 0x47, 0x92, 0xc4, 0x5e, 0x45, 0xc8, 0x52, 0x10, 0x2f, 0x61,
	0x10, 0xa6, 0xc9, 0x2d, 0x52, 0x1e, 0xf0, 0x38, 0x4d, 0xa4, 0xa0, 0x55, 0x0a, 0xf4, 0xbb, 0xb0,
	0x10, 0x1e, 0x81, 0x91, 0x7f, 0x4d, 0x33, 0x3f, 0x0c, 0x28, 0xf2, 0x93, 0x62, 0x79, 0x83, 0x64,
	0x6a, 0x42, 0xd9, 0xf5, 0x74, 0x89, 0x4f, 0x04, 0x3c, 0x2d, 0x51, 0xcb, 0x05, 0x63, 0x2b, 0x61,
	0xb6, 0x58, 0x31, 0x1d, 0x54, 0xd1, 0xb9, 0xca, 0x26, 0x56, 0xec, 0x39, 0xb4, 0x90, 0x28, 0xa5,
	0xd2, 0xa6, 0x3e, 0x1a, 0xac, 0x43, 0x9d, 0x49, 0x34, 0xf7, 0x2a, 0xd6, 0xfa, 0x08, 0x83, 0x71,
	0xb1, 0xda, 0xba, 0xa9, 0x03, 0xd0, 0x84, 0xd6, 0x5f, 0x77, 0x6b, 0x89, 0xaa, 0x0a, 0xd8, 0x24,
	0x57, 0xb7, 0x92, 0xcb, 0x93, 0xb3, 0xf2, 0x98, 0xae, 0x38, 0x39, 0xb3, 0xae, 0xa1, 0xbf, 0x69,
	0x29, 0xad, 0x3d, 0x83, 0x7d, 0xc2, 0x28, 0x26, 0x0c, 0xb9, 0x5f, 0xd0, 0xa2, 0x6c, 0xdb, 0xf5,
	0x7a, 0x0d, 0x36, 0xa7, 0xc5, 0xc6, 0xad, 0xfa, 0x4f, 0xb7, 0x2f, 0xa0, 0x7f, 0x59, 0x11, 0xe3,
	0x82, 0xf3, 0x34, 0x79, 0xc0, 0xab, 0x75, 0x0a, 0xc3, 0x5a, 0x37, 0x4d, 0x79, 0xfc, 0x39, 0x0e,
	0xcb, 0x3b, 0xfe, 0x80, 0x79, 0x1e, 0x7c, 0xc1, 0x07, 0x36, 0x1d, 0x3f, 0x81, 0x4e, 0x33, 0x70,
	0xd6, 0x86, 0x5d, 0x6f, 0x3e, 0x36, 0x76, 0xe4, 0x62, 0x3e, 0x73, 0x0c, 0xe5, 0x78, 0x06, 0x5a,
	0xe5, 0x85, 0x69, 0xa0, 0x5e, 0xbc, 0x17, 0x94, 0x01, 0xfb, 0xee, 0xf4, 0x93, 0x7d, 0xee, 0x3a,
	0xbe, 0x63, 0x5f, 0xd9, 0x86, 0xc2, 0xfa, 0xd0, 0x75, 0xc6, 0xfe, 0x5b, 0xdb, 0x3d, 0x3f, 0x73,
	0x0c, 0x95, 0x0d, 0xa0, 0xe7, 0x4e, 0xdd, 0xab, 0x06, 0xf8, 0x2e, 0x00, 0xb8, 0xb4, 0xaf, 0x9b,
	0xfa, 0x87, 0x32, 0xfa, 0xa9, 0x80, 0x5e, 0x1b, 0x9d, 0x21, 0xdd, 0xc6, 0x21, 0xb2, 0x77, 0xd0,
	0xbb, 0x33, 0x5b, 0xf6, 0x78, 0xf3, 0x18, 0xff, 0x7a, 0xd3, 0xc3, 0x47, 0xf7, 0x93, 0xe2, 0xce,
	0xad, 0x1d, 0xf6, 0x06, 0x3a, 0xcd, 0x18, 0x98, 0xb9, 0x16, 0xfe, 0x31, 0xec, 0xe1, 0xe1, 0x3d,
	0x4c, 0xb9, 0xff, 0x46, 0x2b, 0xff, 0xab, 0xd3, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8a, 0xae,
	0x9e, 0xa6, 0x70, 0x03, 0x00, 0x00,
}