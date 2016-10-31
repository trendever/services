// Code generated by protoc-gen-go.
// source: payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment.proto

It has these top-level messages:
	OrderData
	CreateOrderRequest
	PaymentNotification
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
	// refund fails
	// realy bad in case of Create order: commission was writed off, but pay wasn't created
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

type OrderData struct {
	Amount   uint64   `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	Gateway  string   `protobuf:"bytes,3,opt,name=gateway" json:"gateway,omitempty"`
	UserId   uint64   `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// payment of our service
	ServiceName string `protobuf:"bytes,5,opt,name=service_name,json=serviceName" json:"service_name,omitempty"`
	ServiceData string `protobuf:"bytes,6,opt,name=service_data,json=serviceData" json:"service_data,omitempty"`
	// p2p payment
	LeadId         uint64    `protobuf:"varint,7,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	ConversationId uint64    `protobuf:"varint,8,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	Direction      Direction `protobuf:"varint,9,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ShopCardNumber string    `protobuf:"bytes,10,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,11,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,12,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
}

func (m *OrderData) Reset()                    { *m = OrderData{} }
func (m *OrderData) String() string            { return proto.CompactTextString(m) }
func (*OrderData) ProtoMessage()               {}
func (*OrderData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateOrderRequest struct {
	Data *OrderData `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CreateOrderRequest) GetData() *OrderData {
	if m != nil {
		return m.Data
	}
	return nil
}

type PaymentNotification struct {
	Id   uint64     `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Data *OrderData `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
}

func (m *PaymentNotification) Reset()                    { *m = PaymentNotification{} }
func (m *PaymentNotification) String() string            { return proto.CompactTextString(m) }
func (*PaymentNotification) ProtoMessage()               {}
func (*PaymentNotification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PaymentNotification) GetData() *OrderData {
	if m != nil {
		return m.Data
	}
	return nil
}

type CreateOrderReply struct {
	Id           uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

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
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type BuyOrderReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

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
func (*CancelOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type CancelOrderReply struct {
	Cancelled    bool   `protobuf:"varint,1,opt,name=cancelled" json:"cancelled,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CancelOrderReply) Reset()                    { *m = CancelOrderReply{} }
func (m *CancelOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderReply) ProtoMessage()               {}
func (*CancelOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

// chat messages
type ChatMessageNewOrder struct {
	PayId    uint64   `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

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
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type ChatMessageOrderCancelled struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *ChatMessageOrderCancelled) Reset()                    { *m = ChatMessageOrderCancelled{} }
func (m *ChatMessageOrderCancelled) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageOrderCancelled) ProtoMessage()               {}
func (*ChatMessageOrderCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func init() {
	proto.RegisterType((*OrderData)(nil), "payment.OrderData")
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*PaymentNotification)(nil), "payment.PaymentNotification")
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
	// 912 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x56, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x36, 0xf5, 0xaf, 0xb1, 0x24, 0xd3, 0x1b, 0x34, 0x65, 0xdc, 0x1e, 0x5a, 0x15, 0x6e, 0x03,
	0x17, 0x09, 0x8a, 0xf4, 0x5a, 0x14, 0xa0, 0x49, 0xba, 0x11, 0xac, 0x50, 0xc6, 0x4a, 0x4a, 0x91,
	0x13, 0xc1, 0x50, 0xeb, 0x98, 0x80, 0x44, 0xaa, 0x4b, 0x32, 0x81, 0x80, 0x22, 0xfd, 0x49, 0xef,
	0x7d, 0xa2, 0xbe, 0x4b, 0xd1, 0x6b, 0x5f, 0x22, 0xb3, 0xcb, 0x5f, 0xc5, 0x11, 0xec, 0x83, 0x6f,
	0x3b, 0xdf, 0xcc, 0xce, 0x7c, 0x33, 0xfb, 0x71, 0x24, 0xe8, 0xaf, 0xdd, 0xcd, 0x8a, 0x05, 0xf1,
	0xe3, 0x35, 0x0f, 0xe3, 0x90, 0xb4, 0x33, 0x73, 0xf8, 0x4f, 0x1d, 0xba, 0x13, 0xbe, 0x60, 0xdc,
	0x74, 0x63, 0x97, 0xdc, 0x87, 0x96, 0xbb, 0x0a, 0x93, 0x20, 0xd6, 0x94, 0x2f, 0x94, 0x87, 0x0d,
	0x9a, 0x59, 0xe4, 0x11, 0x74, 0xbc, 0x84, 0x73, 0x16, 0x78, 0x1b, 0xad, 0x86, 0x9e, 0xc1, 0x93,
	0xc3, 0xc7, 0x79, 0x42, 0x23, 0x73, 0xd0, 0x22, 0x84, 0x68, 0xd0, 0x7e, 0xe5, 0xc6, 0xec, 0x8d,
	0xbb, 0xd1, 0xea, 0x18, 0xdd, 0xa5, 0xb9, 0x49, 0x3e, 0x85, 0x76, 0x12, 0x31, 0xee, 0xf8, 0x0b,
	0xad, 0x91, 0x56, 0x10, 0xe6, 0x68, 0x41, 0xbe, 0x84, 0x1e, 0x1e, 0x5e, 0xfb, 0x1e, 0x73, 0x02,
	0x77, 0xc5, 0xb4, 0xa6, 0xbc, 0xb7, 0x9f, 0x61, 0x36, 0x42, 0xd5, 0x90, 0x05, 0x92, 0xd5, 0x5a,
	0x5b, 0x21, 0x92, 0x3f, 0xa6, 0x5f, 0x32, 0x77, 0x21, 0xd2, 0xb7, 0xd3, 0xf4, 0xc2, 0xc4, 0xf4,
	0xdf, 0xc0, 0x81, 0x17, 0x06, 0xaf, 0x19, 0x8f, 0xdc, 0xd8, 0x0f, 0x03, 0x11, 0xd0, 0x91, 0x01,
	0x83, 0x2a, 0x8c, 0x81, 0xdf, 0x41, 0x77, 0xe1, 0x73, 0xe6, 0x09, 0x53, 0xeb, 0xca, 0x56, 0x49,
	0xd1, 0xaa, 0x99, 0x7b, 0x68, 0x19, 0x44, 0x1e, 0x82, 0x1a, 0x5d, 0x85, 0x6b, 0xc7, 0x73, 0xf9,
	0xc2, 0x09, 0x92, 0xd5, 0x4b, 0xc6, 0x35, 0x90, 0xd4, 0x06, 0x02, 0x37, 0x10, 0xb6, 0x25, 0x4a,
	0x8e, 0x01, 0xab, 0xad, 0x56, 0x7e, 0x14, 0x09, 0x0a, 0x97, 0x8c, 0x69, 0xfb, 0x92, 0x43, 0xbf,
	0x44, 0xcf, 0x18, 0x23, 0xdf, 0xc2, 0x61, 0x25, 0x2c, 0x0a, 0x13, 0xee, 0x31, 0xad, 0x27, 0x23,
	0xd5, 0xd2, 0x31, 0x95, 0xf8, 0xf0, 0x07, 0x20, 0x06, 0x67, 0x38, 0x5d, 0xf9, 0x88, 0x94, 0xfd,
	0x92, 0xb0, 0x28, 0x26, 0x5f, 0x43, 0x43, 0x8e, 0x48, 0xbc, 0xe2, 0x7e, 0xa5, 0x81, 0xe2, 0xa5,
	0xa9, 0xf4, 0x0f, 0x9f, 0xc1, 0xbd, 0x8b, 0xd4, 0x65, 0x87, 0xb1, 0x7f, 0xe9, 0x7b, 0x72, 0x0c,
	0x64, 0x00, 0x35, 0x1c, 0x50, 0x2a, 0x01, 0x3c, 0x15, 0xe9, 0x6a, 0x37, 0xa4, 0x0b, 0x40, 0xdd,
	0x22, 0xb3, 0x5e, 0x6e, 0xae, 0xe5, 0x3a, 0x86, 0x26, 0xe3, 0x3c, 0xe4, 0x99, 0x8e, 0x0e, 0x8a,
	0x64, 0x96, 0x40, 0x23, 0x9a, 0x7a, 0xc9, 0x57, 0xd0, 0x97, 0x07, 0x67, 0xc5, 0xa2, 0xc8, 0x7d,
	0xc5, 0x32, 0x21, 0xf5, 0x24, 0xf8, 0x2c, 0xc5, 0x86, 0xef, 0x14, 0x38, 0x38, 0x4d, 0x36, 0x5b,
	0xad, 0x7f, 0x02, 0x2d, 0xcc, 0xe8, 0x14, 0x35, 0x9b, 0x68, 0xe1, 0xbb, 0x0a, 0x1a, 0x6b, 0x59,
	0xb3, 0x8b, 0x34, 0xd6, 0x55, 0xa5, 0xd4, 0xb7, 0x94, 0xb2, 0x25, 0x80, 0xc6, 0x2d, 0x04, 0x30,
	0x7c, 0x0b, 0xfd, 0x92, 0x84, 0x68, 0x19, 0x85, 0xca, 0x59, 0xea, 0x77, 0x12, 0xbe, 0x94, 0x44,
	0x50, 0xa8, 0x39, 0x36, 0xe7, 0xcb, 0x3b, 0x9d, 0xc2, 0xdf, 0x0a, 0x6a, 0xc0, 0x0d, 0x3c, 0xb6,
	0xbc, 0xcd, 0x20, 0xee, 0xae, 0xf1, 0xea, 0xc7, 0xdc, 0xac, 0x7e, 0xcc, 0xc3, 0x5f, 0x51, 0x07,
	0x55, 0x42, 0x62, 0x28, 0x9f, 0x43, 0xd7, 0x93, 0xd8, 0x92, 0xa5, 0x8c, 0x3a, 0xb4, 0x04, 0xee,
	0x74, 0x1e, 0x11, 0xdc, 0x33, 0xae, 0xdc, 0x38, 0x33, 0x6d, 0xf6, 0x46, 0xb2, 0xd8, 0x35, 0x8f,
	0x72, 0xe5, 0xd5, 0x76, 0xae, 0xbc, 0xfa, 0x8d, 0x2b, 0x6f, 0xf8, 0xaf, 0x02, 0x47, 0x95, 0xaa,
	0xd9, 0x57, 0x75, 0xe6, 0x07, 0x7e, 0x74, 0x85, 0xfd, 0xed, 0x28, 0x8e, 0x8b, 0x32, 0x4a, 0x3c,
	0x0f, 0xef, 0xc8, 0x1a, 0x1d, 0x9a, 0x9b, 0xc2, 0x73, 0xe9, 0xfa, 0xcb, 0x84, 0x33, 0xf9, 0x16,
	0xe8, 0xc9, 0xcc, 0x0a, 0xe1, 0xe6, 0x4e, 0xc2, 0xad, 0x9b, 0x77, 0xf4, 0xd6, 0x73, 0xb7, 0x6f,
	0xa3, 0xf3, 0x73, 0x78, 0x50, 0xe9, 0x50, 0x0e, 0xd5, 0x28, 0x1e, 0x70, 0xb7, 0xda, 0x72, 0x89,
	0xd4, 0xaa, 0x12, 0x39, 0x39, 0x86, 0x4e, 0x4e, 0x8a, 0xb4, 0xa1, 0x4e, 0xe7, 0xa7, 0xea, 0x9e,
	0x38, 0xcc, 0xa7, 0xa6, 0xaa, 0x88, 0x83, 0x31, 0xb9, 0x50, 0x6b, 0x27, 0xff, 0x2b, 0xd0, 0x4a,
	0x25, 0x40, 0x5a, 0x50, 0x9b, 0x9c, 0x63, 0x90, 0x0a, 0xbd, 0x91, 0xfd, 0x5c, 0x1f, 0x8f, 0x4c,
	0xc7, 0xd4, 0x67, 0x3a, 0x46, 0xf7, 0xa1, 0x6b, 0x9e, 0x3a, 0x67, 0xfa, 0x68, 0x6c, 0x99, 0x6a,
	0x8d, 0x1c, 0x42, 0x5f, 0x1f, 0x53, 0x4b, 0x37, 0x5f, 0x38, 0x17, 0xfa, 0x0b, 0x84, 0xea, 0x02,
	0xc2, 0xa3, 0x63, 0xe8, 0xb6, 0x61, 0x8d, 0x45, 0x54, 0x03, 0xc7, 0x48, 0x74, 0x7b, 0x32, 0x7b,
	0x6a, 0x51, 0x67, 0x72, 0x61, 0xd9, 0xce, 0x84, 0x9a, 0x16, 0x55, 0x9b, 0x22, 0x74, 0x6e, 0x9f,
	0xdb, 0x93, 0x9f, 0x6d, 0xc7, 0xa2, 0x74, 0x42, 0xd5, 0xb7, 0xe4, 0x00, 0xf6, 0x47, 0xf6, 0x68,
	0x96, 0x57, 0xf8, 0x0d, 0x01, 0x10, 0xe9, 0x32, 0xfb, 0x77, 0x05, 0xb7, 0x4b, 0xd7, 0x78, 0xaa,
	0xcf, 0x1c, 0x13, 0xaf, 0xa9, 0x7f, 0x28, 0x22, 0xc0, 0x98, 0x8c, 0xec, 0x69, 0x0a, 0xfc, 0xa9,
	0x60, 0xd6, 0x1e, 0x16, 0x9f, 0x39, 0xf2, 0x9a, 0x65, 0xa9, 0xef, 0x24, 0x44, 0xad, 0xb3, 0xb9,
	0x6d, 0x66, 0x75, 0xfe, 0x52, 0x4e, 0x1e, 0x61, 0x23, 0xc5, 0xd7, 0x85, 0x55, 0x8d, 0xf1, 0xc8,
	0x4a, 0x2f, 0x4d, 0xb1, 0xf1, 0x12, 0xa0, 0x96, 0xf1, 0x5c, 0x55, 0x9e, 0xfc, 0xa7, 0xc0, 0x20,
	0x13, 0xda, 0x34, 0xfd, 0x11, 0x24, 0x3f, 0x61, 0x4c, 0xb9, 0x81, 0xc9, 0x67, 0xa5, 0x02, 0xae,
	0xfd, 0x48, 0x1c, 0x3d, 0xf8, 0xb8, 0x13, 0x3f, 0xd6, 0xe1, 0x1e, 0xf9, 0x11, 0x3a, 0xf9, 0x52,
	0x23, 0x5a, 0x11, 0xf8, 0xc1, 0xb2, 0x3d, 0xba, 0xff, 0x11, 0x4f, 0x7a, 0x5f, 0x10, 0x29, 0x57,
	0x40, 0x95, 0xc8, 0xb5, 0x4d, 0x55, 0x25, 0xf2, 0xc1, 0xd6, 0x18, 0xee, 0xbd, 0x6c, 0xc9, 0x3f,
	0x2c, 0xdf, 0xbf, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x5b, 0xdd, 0xf5, 0xf1, 0xc1, 0x08, 0x00, 0x00,
}
