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
	CancelOrderRequest
	CancelOrderReply
	ChatMessageNewOrder
	ChatMessagePaymentFinished
	ChatMessageOrderCancelled
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
	Currency_COP Currency = 2
)

var Currency_name = map[int32]string{
	0: "RUB",
	1: "USD",
	2: "COP",
}
var Currency_value = map[string]int32{
	"RUB": 0,
	"USD": 1,
	"COP": 2,
}

func (x Currency) String() string {
	return proto.EnumName(Currency_name, int32(x))
}
func (Currency) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Errors int32

const (
	Errors_OK Errors = 0
	// internal errors
	Errors_INVALID_DATA       Errors = 1
	Errors_DB_FAILED          Errors = 2
	Errors_ALREADY_PAYED      Errors = 3
	Errors_PAY_CANCELLED      Errors = 4
	Errors_ANOTHER_OPEN_ORDER Errors = 5
	Errors_UNKNOWN_ERROR      Errors = 126
	// external errors
	Errors_INIT_FAILED Errors = 127
	Errors_PAY_FAILED  Errors = 128
	Errors_CHAT_DOWN   Errors = 129
	Errors_COINS_DOWN  Errors = 130
	// commission source lacks funds
	Errors_CANT_PAY_FEE Errors = 131
	// realy bad one: refund fail after db fail, so commission was writed off, but pay wasn't created
	Errors_REFUND_ERROR Errors = 132
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	3:   "ALREADY_PAYED",
	4:   "PAY_CANCELLED",
	5:   "ANOTHER_OPEN_ORDER",
	126: "UNKNOWN_ERROR",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
	129: "CHAT_DOWN",
	130: "COINS_DOWN",
	131: "CANT_PAY_FEE",
	132: "REFUND_ERROR",
}
var Errors_value = map[string]int32{
	"OK":                 0,
	"INVALID_DATA":       1,
	"DB_FAILED":          2,
	"ALREADY_PAYED":      3,
	"PAY_CANCELLED":      4,
	"ANOTHER_OPEN_ORDER": 5,
	"UNKNOWN_ERROR":      126,
	"INIT_FAILED":        127,
	"PAY_FAILED":         128,
	"CHAT_DOWN":          129,
	"COINS_DOWN":         130,
	"CANT_PAY_FEE":       131,
	"REFUND_ERROR":       132,
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
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,8,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,9,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateOrderReply struct {
	Id           uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type BuyOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	// parameters to check permissions
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type BuyOrderReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type CancelOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	// parameters to check permissions
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	// userID just to log it
	UserId uint64 `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *CancelOrderRequest) Reset()                    { *m = CancelOrderRequest{} }
func (m *CancelOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderRequest) ProtoMessage()               {}
func (*CancelOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CancelOrderReply struct {
	Cancelled    bool   `protobuf:"varint,1,opt,name=cancelled" json:"cancelled,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CancelOrderReply) Reset()                    { *m = CancelOrderReply{} }
func (m *CancelOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderReply) ProtoMessage()               {}
func (*CancelOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// chat messages
type ChatMessageNewOrder struct {
	PayId    uint64   `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ChatMessagePaymentFinished struct {
	PayId     uint64    `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Success   bool      `protobuf:"varint,3,opt,name=success" json:"success,omitempty"`
	Failure   bool      `protobuf:"varint,4,opt,name=failure" json:"failure,omitempty"`
	Amount    uint64    `protobuf:"varint,5,opt,name=amount" json:"amount,omitempty"`
	Currency  Currency  `protobuf:"varint,6,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	Direction Direction `protobuf:"varint,7,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *ChatMessagePaymentFinished) Reset()                    { *m = ChatMessagePaymentFinished{} }
func (m *ChatMessagePaymentFinished) String() string            { return proto.CompactTextString(m) }
func (*ChatMessagePaymentFinished) ProtoMessage()               {}
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type ChatMessageOrderCancelled struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *ChatMessageOrderCancelled) Reset()                    { *m = ChatMessageOrderCancelled{} }
func (m *ChatMessageOrderCancelled) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageOrderCancelled) ProtoMessage()               {}
func (*ChatMessageOrderCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func init() {
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*CancelOrderRequest)(nil), "payment.CancelOrderRequest")
	proto.RegisterType((*CancelOrderReply)(nil), "payment.CancelOrderReply")
	proto.RegisterType((*ChatMessageNewOrder)(nil), "payment.ChatMessageNewOrder")
	proto.RegisterType((*ChatMessagePaymentFinished)(nil), "payment.ChatMessagePaymentFinished")
	proto.RegisterType((*ChatMessageOrderCancelled)(nil), "payment.ChatMessageOrderCancelled")
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
	CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderReply, error)
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

func (c *paymentServiceClient) CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderReply, error) {
	out := new(CancelOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/CancelOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PaymentService service

type PaymentServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderReply, error)
	BuyOrder(context.Context, *BuyOrderRequest) (*BuyOrderReply, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderReply, error)
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

func _PaymentService_CancelOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CancelOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/CancelOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CancelOrder(ctx, req.(*CancelOrderRequest))
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
		{
			MethodName: "CancelOrder",
			Handler:    _PaymentService_CancelOrder_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 825 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x55, 0x4d, 0x6f, 0xdb, 0x46,
	0x10, 0x35, 0x69, 0xeb, 0x83, 0x13, 0x4b, 0xa6, 0xb7, 0x68, 0xca, 0xb8, 0x3d, 0xb4, 0x2c, 0x82,
	0x06, 0x2e, 0x12, 0x14, 0xe9, 0xbd, 0x00, 0x4d, 0xd2, 0x0d, 0x61, 0x95, 0x14, 0x56, 0x52, 0x8a,
	0x9c, 0x08, 0x86, 0xda, 0xd4, 0x04, 0x24, 0x52, 0x5d, 0x8a, 0x29, 0x04, 0x14, 0xe9, 0x47, 0x7a,
	0xef, 0x5f, 0x2c, 0x7a, 0xe8, 0xa5, 0x7f, 0x22, 0xb3, 0x4b, 0x52, 0xa4, 0x92, 0x08, 0xf1, 0xc1,
	0xb7, 0x9d, 0x37, 0x6f, 0x77, 0xde, 0xbe, 0x1d, 0x0e, 0x61, 0xb0, 0x8a, 0x36, 0x4b, 0x96, 0xae,
	0x1f, 0xad, 0x78, 0xb6, 0xce, 0x48, 0xaf, 0x0a, 0xcd, 0xff, 0x54, 0x20, 0x36, 0x67, 0xd1, 0x9a,
	0x05, 0x7c, 0xce, 0x38, 0x65, 0x3f, 0x17, 0x2c, 0x5f, 0x93, 0xbb, 0xd0, 0x8d, 0x96, 0x59, 0x91,
	0xae, 0x0d, 0xe5, 0x73, 0xe5, 0xc1, 0x11, 0xad, 0x22, 0xf2, 0x10, 0xfa, 0x71, 0xc1, 0x39, 0x4b,
	0xe3, 0x8d, 0xa1, 0x62, 0x66, 0xf8, 0xf8, 0xf4, 0x51, 0x7d, 0xb2, 0x5d, 0x25, 0xe8, 0x96, 0x42,
	0x3e, 0x81, 0xde, 0x82, 0x45, 0xf3, 0x30, 0x99, 0x1b, 0x87, 0xe5, 0x39, 0x22, 0xf4, 0xe6, 0x22,
	0x51, 0xe4, 0x8c, 0x8b, 0xc4, 0x51, 0x99, 0x10, 0x21, 0x26, 0xbe, 0x82, 0x93, 0x38, 0x4b, 0x5f,
	0x32, 0x9e, 0x47, 0xeb, 0x24, 0x4b, 0x05, 0xa1, 0x23, 0x09, 0xc3, 0x36, 0x8c, 0xc4, 0x6f, 0x40,
	0x9b, 0x27, 0x9c, 0xc5, 0x22, 0x34, 0xba, 0x52, 0x0a, 0xd9, 0x4a, 0x71, 0xea, 0x0c, 0x6d, 0x48,
	0xe4, 0x01, 0xe8, 0xf9, 0x75, 0xb6, 0x0a, 0xe3, 0x88, 0xcf, 0xc3, 0xb4, 0x58, 0x3e, 0x67, 0xdc,
	0xe8, 0xe1, 0x46, 0x8d, 0x0e, 0x05, 0x6e, 0x23, 0xec, 0x4b, 0x94, 0xdc, 0x07, 0xac, 0xb6, 0x5c,
	0x26, 0x79, 0x2e, 0x24, 0xbc, 0x60, 0xcc, 0xe8, 0x4b, 0x0d, 0x83, 0x06, 0xbd, 0x64, 0x8c, 0x7c,
	0x0d, 0xa7, 0x2d, 0x5a, 0x9e, 0x15, 0x3c, 0x66, 0x86, 0x26, 0x99, 0x7a, 0x93, 0x98, 0x48, 0xdc,
	0x4c, 0x41, 0xdf, 0xf1, 0x79, 0xb5, 0xd8, 0x90, 0x21, 0xa8, 0x78, 0xbf, 0xd2, 0x61, 0x5c, 0x61,
	0xdd, 0x0e, 0xe3, 0x3c, 0xe3, 0x95, 0xb5, 0x27, 0xdb, 0xfb, 0xb8, 0x02, 0xcd, 0x69, 0x99, 0x25,
	0x5f, 0xc2, 0x40, 0x2e, 0xc2, 0x25, 0xcb, 0xf3, 0xe8, 0x27, 0x26, 0xbd, 0xd5, 0xe8, 0xb1, 0x04,
	0x7f, 0x28, 0x31, 0xf3, 0xb5, 0x02, 0x27, 0x17, 0xc5, 0x66, 0xe7, 0x55, 0x3f, 0x86, 0x2e, 0x9e,
	0x18, 0x6e, 0x6b, 0x76, 0x30, 0x42, 0x2b, 0x85, 0x8c, 0x95, 0xac, 0xa9, 0xa1, 0x8c, 0xd5, 0xfe,
	0x57, 0xdb, 0xf1, 0xfc, 0xe8, 0x06, 0x9e, 0x9b, 0xaf, 0x60, 0xd0, 0x88, 0x10, 0x57, 0xfe, 0x02,
	0x8e, 0x39, 0x2b, 0xf3, 0x61, 0xc1, 0x17, 0x52, 0x88, 0x46, 0xef, 0xd4, 0xd8, 0x8c, 0x2f, 0x6e,
	0xd5, 0x85, 0xbf, 0x15, 0x6c, 0xef, 0x28, 0x8d, 0xd9, 0xe2, 0x26, 0x46, 0xdc, 0xde, 0xc5, 0xdb,
	0x0d, 0xde, 0x69, 0x37, 0xb8, 0xf9, 0x2b, 0xf6, 0x41, 0x5b, 0x90, 0x30, 0xe5, 0x33, 0xd0, 0x62,
	0x89, 0x2d, 0x58, 0xa9, 0xa8, 0x4f, 0x1b, 0xe0, 0x56, 0xfd, 0xc8, 0xe1, 0x23, 0xfb, 0x3a, 0x5a,
	0x57, 0xa1, 0xcf, 0x7e, 0x91, 0x2a, 0xf6, 0xf9, 0xd1, 0x4c, 0x01, 0x75, 0xef, 0x14, 0x38, 0xfc,
	0xe0, 0x14, 0x30, 0xff, 0x51, 0xe0, 0xac, 0x55, 0x75, 0x5c, 0x32, 0x2f, 0x93, 0x34, 0xc9, 0xaf,
	0xf1, 0x7e, 0x7b, 0x8a, 0x1b, 0xd0, 0xcb, 0x8b, 0x38, 0xc6, 0x3d, 0xb2, 0x46, 0x9f, 0xd6, 0xa1,
	0xc8, 0xbc, 0x88, 0x92, 0x45, 0xc1, 0x99, 0x7c, 0x0b, 0xcc, 0x54, 0x61, 0x4b, 0x70, 0x67, 0xaf,
	0xe0, 0xee, 0x87, 0xc7, 0xd6, 0xce, 0x73, 0xf7, 0x6e, 0xd2, 0xe7, 0x57, 0x70, 0xaf, 0x75, 0x43,
	0x69, 0xaa, 0xbd, 0x7d, 0xc0, 0xfd, 0xdd, 0x56, 0xb7, 0x88, 0xda, 0x6e, 0x91, 0xf3, 0xfb, 0xd0,
	0xaf, 0x45, 0x91, 0x1e, 0x1c, 0xd2, 0xd9, 0x85, 0x7e, 0x20, 0x16, 0xb3, 0x89, 0xa3, 0x2b, 0x62,
	0x61, 0x07, 0x63, 0x5d, 0x3d, 0xff, 0x5f, 0x81, 0x6e, 0xd9, 0x02, 0xa4, 0x0b, 0x6a, 0x70, 0x85,
	0x24, 0x1d, 0x8e, 0x3d, 0xff, 0xa9, 0x35, 0xf2, 0x9c, 0xd0, 0xb1, 0xa6, 0x16, 0xb2, 0x07, 0xa0,
	0x39, 0x17, 0xe1, 0xa5, 0xe5, 0x8d, 0x5c, 0x47, 0x57, 0xc9, 0x29, 0x0c, 0xac, 0x11, 0x75, 0x2d,
	0xe7, 0x59, 0x38, 0xb6, 0x9e, 0x21, 0x74, 0x28, 0x20, 0x5c, 0x86, 0xb6, 0xe5, 0xdb, 0xee, 0x48,
	0xb0, 0x8e, 0xd0, 0x46, 0x62, 0xf9, 0xc1, 0xf4, 0x89, 0x4b, 0xc3, 0x60, 0xec, 0xfa, 0x61, 0x40,
	0x1d, 0x97, 0xea, 0x1d, 0x41, 0x9d, 0xf9, 0x57, 0x7e, 0xf0, 0xa3, 0x1f, 0xba, 0x94, 0x06, 0x54,
	0x7f, 0x45, 0x4e, 0xe0, 0x8e, 0xe7, 0x7b, 0xd3, 0xba, 0xc2, 0x6f, 0x08, 0x80, 0x38, 0xae, 0x8a,
	0x7f, 0x57, 0x70, 0xba, 0x68, 0xf6, 0x13, 0x6b, 0x1a, 0x3a, 0xb8, 0x4d, 0xff, 0x43, 0x11, 0x04,
	0x3b, 0xf0, 0xfc, 0x49, 0x09, 0xfc, 0xa9, 0xe0, 0xa9, 0xc7, 0x58, 0x7c, 0x1a, 0xca, 0x6d, 0xae,
	0xab, 0xbf, 0x96, 0x10, 0x75, 0x2f, 0x67, 0xbe, 0x53, 0xd5, 0xf9, 0x4b, 0x39, 0x7f, 0x88, 0x17,
	0xd9, 0x7e, 0x5d, 0x58, 0xd5, 0x1e, 0x79, 0x6e, 0xb9, 0x69, 0x82, 0x17, 0x6f, 0x00, 0xea, 0xda,
	0x4f, 0x75, 0xe5, 0xf1, 0xbf, 0x0a, 0x0c, 0xab, 0x46, 0x9b, 0x30, 0xfe, 0x32, 0x89, 0x19, 0xf9,
	0x1e, 0x39, 0xcd, 0x04, 0x26, 0x9f, 0x36, 0x1d, 0xf0, 0xce, 0xff, 0xef, 0xec, 0xde, 0xfb, 0x93,
	0xf8, 0xb1, 0x9a, 0x07, 0xe4, 0x3b, 0xe8, 0xd7, 0x43, 0x8d, 0x18, 0x5b, 0xe2, 0x5b, 0xc3, 0xf6,
	0xec, 0xee, 0x7b, 0x32, 0xe5, 0x7e, 0x21, 0xa4, 0x19, 0x01, 0x6d, 0x21, 0xef, 0x4c, 0xaa, 0xb6,
	0x90, 0xb7, 0xa6, 0x86, 0x79, 0xf0, 0xbc, 0x2b, 0x7f, 0xe6, 0xdf, 0xbe, 0x09, 0x00, 0x00, 0xff,
	0xff, 0x45, 0xb9, 0xb6, 0x58, 0xdd, 0x07, 0x00, 0x00,
}
