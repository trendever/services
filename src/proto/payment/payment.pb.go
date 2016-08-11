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

type Direction int32

const (
	Direction_TO_CLIENT   Direction = 0
	Direction_FROM_CLIENT Direction = 1
)

var Direction_name = map[int32]string{
	0: "TO_CLIENT",
	1: "FROM_CLIENT",
}
var Direction_value = map[string]int32{
	"TO_CLIENT":   0,
	"FROM_CLIENT": 1,
}

func (x Direction) String() string {
	return proto.EnumName(Direction_name, int32(x))
}
func (Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type CreateOrderRequest struct {
	Amount         uint64    `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency       Currency  `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId         uint64    `protobuf:"varint,3,opt,name=lead_id" json:"lead_id,omitempty"`
	UserId         uint64    `protobuf:"varint,4,opt,name=user_id" json:"user_id,omitempty"`
	ConversationId uint64    `protobuf:"varint,5,opt,name=conversation_id" json:"conversation_id,omitempty"`
	Direction      Direction `protobuf:"varint,6,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ShopCardNumber string    `protobuf:"bytes,7,opt,name=shop_card_number" json:"shop_card_number,omitempty"`
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
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	// parameters to check correctness
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type BuyOrderReply struct {
	RedirectUrl string `protobuf:"bytes,1,opt,name=redirect_url" json:"redirect_url,omitempty"`
	Error       Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// chat messages
type PaymentButton struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id" json:"pay_id,omitempty"`
}

func (m *PaymentButton) Reset()                    { *m = PaymentButton{} }
func (m *PaymentButton) String() string            { return proto.CompactTextString(m) }
func (*PaymentButton) ProtoMessage()               {}
func (*PaymentButton) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type PaymentNotificationMessage struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id" json:"pay_id,omitempty"`
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
	// 473 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x53, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xad, 0xd3, 0xd4, 0x89, 0xa7, 0x4d, 0x62, 0x56, 0x08, 0x4c, 0x40, 0x80, 0x8c, 0x90, 0x50,
	0x40, 0x3d, 0x94, 0x7b, 0x25, 0x27, 0x4e, 0x91, 0x45, 0x9a, 0x54, 0x49, 0x8a, 0xc4, 0xc9, 0x72,
	0xed, 0x01, 0x2c, 0x25, 0xb6, 0x19, 0xef, 0x56, 0xca, 0x09, 0x7e, 0x83, 0x0f, 0xe2, 0xbf, 0x58,
	0x6f, 0x6c, 0x97, 0xa6, 0x45, 0xdc, 0xbc, 0xef, 0xcd, 0x9b, 0x99, 0x37, 0x33, 0x86, 0x4e, 0x16,
	0x6c, 0xd6, 0x98, 0xf0, 0xe3, 0x8c, 0x52, 0x9e, 0xb2, 0x56, 0xf9, 0xb4, 0x7f, 0x6b, 0xc0, 0x46,
	0x84, 0x01, 0xc7, 0x19, 0x45, 0x48, 0x73, 0xfc, 0x2e, 0x30, 0xe7, 0xac, 0x0b, 0x7a, 0xb0, 0x4e,
	0x45, 0xc2, 0x2d, 0xed, 0xa5, 0xf6, 0xa6, 0xc9, 0x5e, 0x41, 0x3b, 0x14, 0x44, 0x98, 0x84, 0x1b,
	0xab, 0x21, 0x91, 0xee, 0xc9, 0x83, 0xe3, 0x2a, 0xe3, 0xa8, 0x24, 0x58, 0x0f, 0x5a, 0x2b, 0x0c,
	0x22, 0x3f, 0x8e, 0xac, 0x7d, 0xa5, 0x92, 0x80, 0xc8, 0x91, 0x0a, 0xa0, 0xa9, 0x80, 0xc7, 0xd0,
	0x0b, 0xd3, 0xe4, 0x1a, 0x29, 0x0f, 0x78, 0x9c, 0x26, 0x05, 0x71, 0xa0, 0x88, 0xd7, 0x60, 0x44,
	0x31, 0x61, 0x58, 0xa0, 0x96, 0xae, 0x0a, 0xb0, 0xba, 0x80, 0x5b, 0x31, 0xcc, 0x02, 0x33, 0xff,
	0x96, 0x66, 0x7e, 0x18, 0x50, 0xe4, 0x27, 0x62, 0x7d, 0x85, 0x64, 0xb5, 0x64, 0xb4, 0x61, 0x9f,
	0x82, 0x79, 0xcb, 0x46, 0xb6, 0xda, 0x30, 0x80, 0x86, 0x2c, 0xb0, 0x35, 0xf0, 0x1c, 0x0e, 0x90,
	0x28, 0xa5, 0xb2, 0xfb, 0x5e, 0x9d, 0x7c, 0x5c, 0xa0, 0xb9, 0x8d, 0xd0, 0x1b, 0x8a, 0xcd, 0xee,
	0x0c, 0x64, 0x90, 0x5f, 0xa7, 0x28, 0xd2, 0x65, 0x4a, 0x6f, 0xdc, 0xb5, 0x7a, 0xcb, 0x40, 0xf3,
	0x5f, 0x06, 0xec, 0x31, 0x74, 0x6e, 0xca, 0x14, 0x3d, 0x3e, 0x84, 0x23, 0xc2, 0xad, 0xd2, 0x17,
	0xb4, 0x52, 0xa5, 0x8c, 0xff, 0x76, 0xfb, 0x02, 0x3a, 0x17, 0x5b, 0x64, 0x28, 0x38, 0x97, 0x83,
	0xd9, 0xe9, 0xd5, 0x7e, 0x07, 0xfd, 0x32, 0x60, 0x9a, 0xf2, 0xf8, 0x4b, 0x1c, 0xaa, 0x79, 0x9f,
	0x63, 0x9e, 0x07, 0x5f, 0x71, 0x37, 0x7a, 0xf0, 0x0c, 0xda, 0xf5, 0x12, 0x5b, 0xb0, 0x3f, 0xbf,
	0x1c, 0x9a, 0x7b, 0xc5, 0xc7, 0xe5, 0xc2, 0x35, 0xb5, 0xc1, 0x02, 0xf4, 0x6d, 0x59, 0xa6, 0x43,
	0x63, 0xf6, 0x51, 0x52, 0x26, 0x1c, 0x79, 0xd3, 0x4f, 0xce, 0xc4, 0x73, 0x7d, 0xd7, 0x59, 0x3a,
	0xa6, 0xc6, 0x3a, 0x60, 0xb8, 0x43, 0xff, 0xcc, 0xf1, 0x26, 0x63, 0xd7, 0x6c, 0xc8, 0xf1, 0x1c,
	0x7a, 0x53, 0x6f, 0x59, 0x01, 0x3f, 0x24, 0x00, 0x17, 0xce, 0xe7, 0xea, 0xfd, 0x53, 0x1b, 0xbc,
	0x95, 0x82, 0x7a, 0xad, 0x52, 0xbd, 0x9c, 0xf9, 0xa3, 0x89, 0x37, 0x9e, 0x2e, 0x65, 0x7a, 0xa9,
	0x3e, 0x9b, 0xcf, 0xce, 0x2b, 0x40, 0x3b, 0xf9, 0xa5, 0x41, 0xb7, 0xb4, 0xb3, 0x40, 0xba, 0x8e,
	0x43, 0x64, 0x1f, 0xe0, 0xf0, 0xaf, 0x7d, 0xb3, 0xa7, 0x37, 0xd7, 0x78, 0xe7, 0x98, 0xfb, 0x4f,
	0xee, 0x27, 0xe5, 0xf8, 0xed, 0x3d, 0x76, 0x0a, 0xed, 0x6a, 0x23, 0xcc, 0xaa, 0x03, 0x77, 0x6e,
	0xa1, 0xff, 0xe8, 0x1e, 0x46, 0xe9, 0xaf, 0x74, 0xf5, 0x43, 0xbd, 0xff, 0x13, 0x00, 0x00, 0xff,
	0xff, 0x2f, 0x74, 0xd1, 0x33, 0x61, 0x03, 0x00, 0x00,
}
