// Code generated by protoc-gen-go.
// source: payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment.proto

It has these top-level messages:
	CreateOrderRequest
	CreateOrderReply
	BuyOrderRequest
	BuyOrderReply
	ChatMessageNewOrder
	ChatMessagePaymentFinished
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
	Errors_INVALID_DATA  Errors = 1
	Errors_DB_FAILED     Errors = 2
	Errors_ALREADY_PAYED Errors = 3
	// external errors
	Errors_INIT_FAILED Errors = 127
	Errors_PAY_FAILED  Errors = 128
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	3:   "ALREADY_PAYED",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
}
var Errors_value = map[string]int32{
	"OK":            0,
	"INVALID_DATA":  1,
	"DB_FAILED":     2,
	"ALREADY_PAYED": 3,
	"INIT_FAILED":   127,
	"PAY_FAILED":    128,
}

func (x Errors) String() string {
	return proto.EnumName(Errors_name, int32(x))
}
func (Errors) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Direction int32

const (
	Direction_CLIENT_PAYS Direction = 0
	Direction_CLIENT_RECV Direction = 1
)

var Direction_name = map[int32]string{
	0: "CLIENT_PAYS",
	1: "CLIENT_RECV",
}
var Direction_value = map[string]int32{
	"CLIENT_PAYS": 0,
	"CLIENT_RECV": 1,
}

func (x Direction) String() string {
	return proto.EnumName(Direction_name, int32(x))
}
func (Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type CreateOrderRequest struct {
	Amount         uint64    `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency       Currency  `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId         uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	UserId         uint64    `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ConversationId uint64    `protobuf:"varint,5,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	Direction      Direction `protobuf:"varint,6,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ShopCardNumber string    `protobuf:"bytes,7,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateOrderReply struct {
	Id    uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type BuyOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	// parameters to check correctness
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
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
type ChatMessageNewOrder struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount uint64 `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ChatMessagePaymentFinished struct {
	PayId     uint64    `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Success   bool      `protobuf:"varint,3,opt,name=success" json:"success,omitempty"`
	Amount    uint64    `protobuf:"varint,4,opt,name=amount" json:"amount,omitempty"`
	Direction Direction `protobuf:"varint,6,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *ChatMessagePaymentFinished) Reset()                    { *m = ChatMessagePaymentFinished{} }
func (m *ChatMessagePaymentFinished) String() string            { return proto.CompactTextString(m) }
func (*ChatMessagePaymentFinished) ProtoMessage()               {}
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func init() {
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*ChatMessageNewOrder)(nil), "payment.ChatMessageNewOrder")
	proto.RegisterType((*ChatMessagePaymentFinished)(nil), "payment.ChatMessagePaymentFinished")
	proto.RegisterEnum("payment.Currency", Currency_name, Currency_value)
	proto.RegisterEnum("payment.Errors", Errors_name, Errors_value)
	proto.RegisterEnum("payment.Direction", Direction_name, Direction_value)
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

func init() { proto.RegisterFile("payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 566 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x54, 0xdf, 0x6f, 0xd2, 0x50,
	0x18, 0x5d, 0x0b, 0x14, 0xf8, 0x36, 0xa0, 0xbb, 0xc6, 0x59, 0xd1, 0x07, 0x6d, 0x62, 0x5c, 0x48,
	0xb6, 0x98, 0xf9, 0x6e, 0x52, 0x5a, 0x66, 0x1a, 0x91, 0x2d, 0x17, 0x58, 0xc2, 0x53, 0xd3, 0xb5,
	0x57, 0x69, 0x02, 0x6d, 0xbd, 0x6d, 0x67, 0x78, 0xd2, 0xc4, 0xff, 0xc0, 0x17, 0xe3, 0x7f, 0xeb,
	0xbd, 0xfd, 0x09, 0x3a, 0x92, 0xf9, 0xc6, 0x77, 0xce, 0xb9, 0xdf, 0x8f, 0x93, 0x53, 0xa0, 0x13,
	0xda, 0x9b, 0x35, 0xf1, 0xe3, 0xf3, 0x90, 0x06, 0x71, 0x80, 0x9a, 0x79, 0xa9, 0xfe, 0x14, 0x01,
	0xe9, 0x94, 0xd8, 0x31, 0xb9, 0xa2, 0x2e, 0xa1, 0x98, 0x7c, 0x49, 0x48, 0x14, 0xa3, 0x13, 0x90,
	0xec, 0x75, 0x90, 0xf8, 0xb1, 0x22, 0xbc, 0x10, 0x4e, 0xeb, 0x38, 0xaf, 0xd0, 0x19, 0xb4, 0x9c,
	0x84, 0x52, 0xe2, 0x3b, 0x1b, 0x45, 0x64, 0x4c, 0xf7, 0xe2, 0xf8, 0xbc, 0xe8, 0xac, 0xe7, 0x04,
	0x2e, 0x25, 0xe8, 0x09, 0x34, 0x57, 0xc4, 0x76, 0x2d, 0xcf, 0x55, 0x6a, 0x59, 0x1f, 0x5e, 0x9a,
	0x2e, 0x27, 0x92, 0x88, 0x50, 0x4e, 0xd4, 0x33, 0x82, 0x97, 0x8c, 0x78, 0x0d, 0x3d, 0x27, 0xf0,
	0xef, 0x08, 0x8d, 0xec, 0xd8, 0x0b, 0x7c, 0x2e, 0x68, 0xa4, 0x82, 0xee, 0x36, 0xcc, 0x84, 0x6f,
	0xa0, 0xed, 0x7a, 0x94, 0x38, 0xbc, 0x54, 0xa4, 0x74, 0x15, 0x54, 0xae, 0x62, 0x14, 0x0c, 0xae,
	0x44, 0xe8, 0x14, 0xe4, 0x68, 0x19, 0x84, 0x96, 0x63, 0x53, 0xd7, 0xf2, 0x93, 0xf5, 0x2d, 0xa1,
	0x4a, 0x93, 0x3d, 0x6c, 0xe3, 0x2e, 0xc7, 0x75, 0x06, 0x4f, 0x52, 0x54, 0x35, 0x41, 0xde, 0xf1,
	0x24, 0x5c, 0x6d, 0x50, 0x17, 0x44, 0xb6, 0x4b, 0xe6, 0x06, 0xfb, 0x85, 0x5e, 0x41, 0x83, 0x50,
	0x1a, 0xd0, 0xdc, 0x86, 0x5e, 0x39, 0x7b, 0xc4, 0xd1, 0x08, 0x67, 0xac, 0xfa, 0x43, 0x80, 0xde,
	0x30, 0xd9, 0xec, 0x98, 0xfb, 0x18, 0x24, 0x26, 0xb6, 0xca, 0x76, 0x0d, 0x56, 0xb1, 0x8b, 0xf8,
	0x84, 0x30, 0x6d, 0xd7, 0x66, 0x13, 0xc2, 0xfd, 0xe6, 0xed, 0x9c, 0x5e, 0x7f, 0xc0, 0xe9, 0xea,
	0x02, 0x3a, 0xd5, 0x12, 0xfc, 0x9a, 0x97, 0x70, 0x44, 0x49, 0xc6, 0x5b, 0x09, 0x5d, 0xa5, 0x8b,
	0xb4, 0xf1, 0x61, 0x81, 0xcd, 0xe9, 0xea, 0xa1, 0x07, 0x1a, 0xf0, 0x48, 0x5f, 0xda, 0xf1, 0x47,
	0x12, 0x45, 0xf6, 0x67, 0x32, 0x21, 0x5f, 0xd3, 0x29, 0xfb, 0x6e, 0xac, 0x72, 0x25, 0x6e, 0xe7,
	0x4a, 0xfd, 0x25, 0x40, 0x7f, 0xab, 0xcd, 0x75, 0x36, 0xea, 0xd2, 0xf3, 0xbd, 0x68, 0x49, 0xdc,
	0x7d, 0xdd, 0x14, 0x68, 0x46, 0x89, 0xe3, 0xb0, 0x37, 0xa9, 0x43, 0x2d, 0x5c, 0x94, 0x5b, 0x73,
	0xea, 0x3b, 0xf9, 0xfd, 0xef, 0xd4, 0x0c, 0x9e, 0x43, 0xab, 0x08, 0x36, 0x6a, 0x42, 0x0d, 0xcf,
	0x87, 0xf2, 0x01, 0xff, 0x31, 0x9f, 0x1a, 0xb2, 0x30, 0xf8, 0x04, 0x52, 0x66, 0x07, 0x92, 0x40,
	0xbc, 0xfa, 0xc0, 0x28, 0x19, 0x8e, 0xcc, 0xc9, 0x8d, 0x36, 0x36, 0x0d, 0xcb, 0xd0, 0x66, 0x9a,
	0x2c, 0xa0, 0x0e, 0xb4, 0x8d, 0xa1, 0x75, 0xa9, 0x99, 0xe3, 0x91, 0x21, 0x8b, 0xe8, 0x18, 0x3a,
	0xda, 0x18, 0x8f, 0x34, 0x63, 0x61, 0x5d, 0x6b, 0x0b, 0x06, 0xd5, 0x50, 0x0f, 0x0e, 0xcd, 0x89,
	0x39, 0x2b, 0x34, 0xdf, 0x18, 0x00, 0x8c, 0x2b, 0xea, 0xef, 0xc2, 0xe0, 0x8c, 0xf5, 0x28, 0x83,
	0xcc, 0xe4, 0xfa, 0xd8, 0x1c, 0x4d, 0x66, 0xbc, 0xc1, 0x94, 0xcd, 0xac, 0x00, 0x3c, 0xd2, 0x6f,
	0x64, 0xe1, 0xe2, 0xb7, 0x00, 0xdd, 0xdc, 0xc3, 0x29, 0xa1, 0x77, 0x9e, 0x43, 0xd0, 0x7b, 0xa6,
	0xa9, 0x32, 0x8d, 0x9e, 0x55, 0x9f, 0xed, 0x3f, 0x5f, 0x7f, 0xff, 0xe9, 0xfd, 0x24, 0x0b, 0x8e,
	0x7a, 0x80, 0xde, 0x41, 0xab, 0xc8, 0x12, 0x52, 0x4a, 0xe1, 0x5f, 0x19, 0xef, 0x9f, 0xdc, 0xc3,
	0xa4, 0xef, 0x6f, 0xa5, 0xf4, 0x1f, 0xe8, 0xed, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc7, 0x49,
	0x63, 0x4d, 0x92, 0x04, 0x00, 0x00,
}
