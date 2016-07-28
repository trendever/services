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
	Amount             uint64   `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency           Currency `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId             uint64   `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	ShopCardNumber     string   `protobuf:"bytes,4,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
	CustomerCardNumber string   `protobuf:"bytes,5,opt,name=customer_card_number,json=customerCardNumber" json:"customer_card_number,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateOrderReply struct {
	Id          uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	RedirectUrl string `protobuf:"bytes,2,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error       Errors `protobuf:"varint,3,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type BuyOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
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

func init() {
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
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
	// 436 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x52, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xed, 0xba, 0x8d, 0x93, 0x4c, 0x5a, 0xc7, 0x8c, 0x68, 0x31, 0x85, 0x03, 0x58, 0x42, 0xaa,
	0x2a, 0x51, 0x50, 0xb8, 0x70, 0x42, 0x72, 0xe2, 0x82, 0x2c, 0xaa, 0xb4, 0x72, 0x1a, 0xa4, 0x9e,
	0x2c, 0xd7, 0x5e, 0x89, 0x48, 0xfe, 0x62, 0x6c, 0x23, 0xf9, 0x04, 0x7f, 0x83, 0x7f, 0xc5, 0x4f,
	0x62, 0xfd, 0x19, 0xd2, 0x46, 0x39, 0xed, 0xcc, 0x7b, 0x6f, 0x76, 0xdf, 0xcc, 0x0e, 0x1c, 0x27,
	0x6e, 0x11, 0xf2, 0x28, 0x7b, 0xd7, 0x9c, 0x17, 0x09, 0xc5, 0x59, 0x8c, 0xfd, 0x26, 0xd5, 0xff,
	0x32, 0xc0, 0x19, 0x71, 0x37, 0xe3, 0xd7, 0xe4, 0x73, 0xb2, 0xf9, 0x8f, 0x9c, 0xa7, 0x19, 0x9e,
	0x80, 0xec, 0x86, 0x71, 0x1e, 0x65, 0x1a, 0x7b, 0xc5, 0xce, 0x0e, 0xec, 0x26, 0xc3, 0xb7, 0x30,
	0xf0, 0x72, 0x22, 0x1e, 0x79, 0x85, 0x26, 0x09, 0x46, 0x99, 0x3c, 0xb9, 0x68, 0x6f, 0x9e, 0x35,
	0x84, 0xdd, 0x49, 0xf0, 0x19, 0xf4, 0x03, 0xee, 0xfa, 0xce, 0xca, 0xd7, 0xf6, 0xeb, 0x7b, 0xca,
	0xd4, 0xf2, 0xf1, 0x0c, 0xd4, 0xf4, 0x7b, 0x9c, 0x38, 0x9e, 0x4b, 0xbe, 0x13, 0xe5, 0xe1, 0x3d,
	0x27, 0xed, 0x40, 0x28, 0x86, 0xb6, 0x52, 0xe2, 0x33, 0x01, 0xcf, 0x2b, 0x14, 0xdf, 0xc3, 0x53,
	0x2f, 0x4f, 0xb3, 0x38, 0xe4, 0xb4, 0xa1, 0xee, 0x55, 0x6a, 0x6c, 0xb9, 0x75, 0x85, 0x1e, 0x80,
	0xba, 0xd1, 0x51, 0x12, 0x14, 0xa8, 0x80, 0x24, 0x3c, 0xd4, 0xbd, 0x88, 0x08, 0x5f, 0xc3, 0x21,
	0x71, 0x7f, 0x45, 0xdc, 0xcb, 0x9c, 0x9c, 0x82, 0xaa, 0x97, 0xa1, 0x3d, 0x6a, 0xb1, 0x25, 0x05,
	0xf8, 0x06, 0x7a, 0x9c, 0x28, 0xa6, 0xca, 0xb9, 0x32, 0x19, 0x77, 0x7d, 0x5e, 0x96, 0x68, 0x6a,
	0xd7, 0xac, 0xfe, 0x11, 0xc6, 0xd3, 0xbc, 0xd8, 0x18, 0xde, 0x31, 0xc8, 0x42, 0xeb, 0x74, 0x0f,
	0xf6, 0x44, 0x26, 0x7a, 0x2e, 0x3d, 0x24, 0xcd, 0x4b, 0x22, 0xd2, 0xef, 0xe0, 0x68, 0x5d, 0x59,
	0x9a, 0x7c, 0x68, 0x8a, 0xed, 0x30, 0x25, 0xed, 0x32, 0x75, 0xfe, 0x12, 0x06, 0xed, 0x6f, 0x60,
	0x1f, 0xf6, 0xed, 0xe5, 0x54, 0xdd, 0x2b, 0x83, 0xe5, 0xc2, 0x54, 0xd9, 0xf9, 0x02, 0xe4, 0x5a,
	0x8e, 0x32, 0x48, 0xd7, 0x5f, 0x05, 0xa5, 0xc2, 0xa1, 0x35, 0xff, 0x66, 0x5c, 0x59, 0xa6, 0x63,
	0x1a, 0xb7, 0x86, 0xca, 0xf0, 0x08, 0x86, 0xe6, 0xd4, 0xf9, 0x6c, 0x58, 0x57, 0x97, 0xa6, 0x2a,
	0xe1, 0x18, 0x46, 0xd6, 0xdc, 0xba, 0x6d, 0x81, 0x5f, 0x02, 0x80, 0x1b, 0xe3, 0xae, 0xcd, 0x7f,
	0xb3, 0xc9, 0x1f, 0x06, 0xca, 0x4d, 0x6d, 0x66, 0xc1, 0xe9, 0xe7, 0xca, 0xe3, 0xf8, 0x05, 0x46,
	0xff, 0x7d, 0x04, 0xbe, 0x58, 0x6f, 0xca, 0xa3, 0x85, 0x3b, 0x7d, 0xbe, 0x9d, 0x14, 0x63, 0xd1,
	0xf7, 0xf0, 0x13, 0x0c, 0xda, 0x49, 0xa1, 0xd6, 0x09, 0x1f, 0x8c, 0xfd, 0xf4, 0x64, 0x0b, 0x53,
	0xd5, 0xdf, 0xcb, 0xd5, 0xd2, 0x7f, 0xf8, 0x17, 0x00, 0x00, 0xff, 0xff, 0xab, 0x0c, 0x7a, 0x41,
	0x0d, 0x03, 0x00, 0x00,
}
