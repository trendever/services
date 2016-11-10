// Code generated by protoc-gen-go.
// source: payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment.proto

It has these top-level messages:
	UsualData
	ChatMessageNewOrder
	ChatMessagePaymentFinished
	ChatMessageOrderCancelled
	OrderData
	CreateOrderRequest
	GetOrderRequest
	GetOrderReply
	UpdateServiceDataRequest
	UpdateServiceDataReply
	PaymentNotification
	CreateOrderReply
	UserInfo
	BuyOrderRequest
	BuyOrderReply
	CancelOrderRequest
	CancelOrderReply
	AddCardRequest
	AddCardReply
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

type Event int32

const (
	Event_Created    Event = 0
	Event_Cancelled  Event = 1
	Event_PayFailed  Event = 2
	Event_PaySuccess Event = 3
)

var Event_name = map[int32]string{
	0: "Created",
	1: "Cancelled",
	2: "PayFailed",
	3: "PaySuccess",
}
var Event_value = map[string]int32{
	"Created":    0,
	"Cancelled":  1,
	"PayFailed":  2,
	"PaySuccess": 3,
}

func (x Event) String() string {
	return proto.EnumName(Event_name, int32(x))
}
func (Event) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

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
func (Currency) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Errors int32

const (
	Errors_OK Errors = 0
	// internal errors
	Errors_INVALID_DATA       Errors = 1
	Errors_DB_FAILED          Errors = 2
	Errors_ALREADY_PAYED      Errors = 3
	Errors_PAY_CANCELLED      Errors = 4
	Errors_ANOTHER_OPEN_ORDER Errors = 5
	Errors_ALREADY_CANCELLED  Errors = 6
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
	Errors_NATS_FAILED  Errors = 133
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	3:   "ALREADY_PAYED",
	4:   "PAY_CANCELLED",
	5:   "ANOTHER_OPEN_ORDER",
	6:   "ALREADY_CANCELLED",
	126: "UNKNOWN_ERROR",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
	129: "CHAT_DOWN",
	130: "COINS_DOWN",
	131: "CANT_PAY_FEE",
	132: "REFUND_ERROR",
	133: "NATS_FAILED",
}
var Errors_value = map[string]int32{
	"OK":                 0,
	"INVALID_DATA":       1,
	"DB_FAILED":          2,
	"ALREADY_PAYED":      3,
	"PAY_CANCELLED":      4,
	"ANOTHER_OPEN_ORDER": 5,
	"ALREADY_CANCELLED":  6,
	"UNKNOWN_ERROR":      126,
	"INIT_FAILED":        127,
	"PAY_FAILED":         128,
	"CHAT_DOWN":          129,
	"COINS_DOWN":         130,
	"CANT_PAY_FEE":       131,
	"REFUND_ERROR":       132,
	"NATS_FAILED":        133,
}

func (x Errors) String() string {
	return proto.EnumName(Errors_name, int32(x))
}
func (Errors) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

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
func (Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// service_data structs
type UsualData struct {
	Direction      Direction `protobuf:"varint,1,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ConversationId uint64    `protobuf:"varint,2,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	MessageId      uint64    `protobuf:"varint,3,opt,name=message_id,json=messageId" json:"message_id,omitempty"`
}

func (m *UsualData) Reset()                    { *m = UsualData{} }
func (m *UsualData) String() string            { return proto.CompactTextString(m) }
func (*UsualData) ProtoMessage()               {}
func (*UsualData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// chat messages
type ChatMessageNewOrder struct {
	PayId    uint64   `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type ChatMessageOrderCancelled struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *ChatMessageOrderCancelled) Reset()                    { *m = ChatMessageOrderCancelled{} }
func (m *ChatMessageOrderCancelled) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageOrderCancelled) ProtoMessage()               {}
func (*ChatMessageOrderCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type OrderData struct {
	Amount   uint64   `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	Gateway  string   `protobuf:"bytes,3,opt,name=gateway" json:"gateway,omitempty"`
	UserId   uint64   `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// payment of our service
	ServiceName string `protobuf:"bytes,5,opt,name=service_name,json=serviceName" json:"service_name,omitempty"`
	ServiceData string `protobuf:"bytes,6,opt,name=service_data,json=serviceData" json:"service_data,omitempty"`
	// p2p payment
	LeadId         uint64 `protobuf:"varint,7,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	ShopCardNumber string `protobuf:"bytes,10,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,11,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,12,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
	Cancelled        bool   `protobuf:"varint,13,opt,name=cancelled" json:"cancelled,omitempty"`
	Comment          string `protobuf:"bytes,14,opt,name=comment" json:"comment,omitempty"`
}

func (m *OrderData) Reset()                    { *m = OrderData{} }
func (m *OrderData) String() string            { return proto.CompactTextString(m) }
func (*OrderData) ProtoMessage()               {}
func (*OrderData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CreateOrderRequest struct {
	Data *OrderData `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *CreateOrderRequest) GetData() *OrderData {
	if m != nil {
		return m.Data
	}
	return nil
}

type GetOrderRequest struct {
	Id uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetOrderRequest) Reset()                    { *m = GetOrderRequest{} }
func (m *GetOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*GetOrderRequest) ProtoMessage()               {}
func (*GetOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type GetOrderReply struct {
	Order *OrderData `protobuf:"bytes,1,opt,name=order" json:"order,omitempty"`
}

func (m *GetOrderReply) Reset()                    { *m = GetOrderReply{} }
func (m *GetOrderReply) String() string            { return proto.CompactTextString(m) }
func (*GetOrderReply) ProtoMessage()               {}
func (*GetOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *GetOrderReply) GetOrder() *OrderData {
	if m != nil {
		return m.Order
	}
	return nil
}

type UpdateServiceDataRequest struct {
	Id      uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	NewData string `protobuf:"bytes,2,opt,name=new_data,json=newData" json:"new_data,omitempty"`
}

func (m *UpdateServiceDataRequest) Reset()                    { *m = UpdateServiceDataRequest{} }
func (m *UpdateServiceDataRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateServiceDataRequest) ProtoMessage()               {}
func (*UpdateServiceDataRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type UpdateServiceDataReply struct {
}

func (m *UpdateServiceDataReply) Reset()                    { *m = UpdateServiceDataReply{} }
func (m *UpdateServiceDataReply) String() string            { return proto.CompactTextString(m) }
func (*UpdateServiceDataReply) ProtoMessage()               {}
func (*UpdateServiceDataReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type PaymentNotification struct {
	Id            uint64     `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Data          *OrderData `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
	Event         Event      `protobuf:"varint,3,opt,name=event,enum=payment.Event" json:"event,omitempty"`
	InvokerUserId uint64     `protobuf:"varint,4,opt,name=invoker_user_id,json=invokerUserId" json:"invoker_user_id,omitempty"`
}

func (m *PaymentNotification) Reset()                    { *m = PaymentNotification{} }
func (m *PaymentNotification) String() string            { return proto.CompactTextString(m) }
func (*PaymentNotification) ProtoMessage()               {}
func (*PaymentNotification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

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
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

type UserInfo struct {
	Ip     string `protobuf:"bytes,1,opt,name=ip" json:"ip,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	Phone  string `protobuf:"bytes,3,opt,name=phone" json:"phone,omitempty"`
}

func (m *UserInfo) Reset()                    { *m = UserInfo{} }
func (m *UserInfo) String() string            { return proto.CompactTextString(m) }
func (*UserInfo) ProtoMessage()               {}
func (*UserInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type BuyOrderRequest struct {
	PayId uint64    `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	User  *UserInfo `protobuf:"bytes,2,opt,name=user" json:"user,omitempty"`
	// use only when pay_id == 0
	Gateway string `protobuf:"bytes,3,opt,name=gateway" json:"gateway,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *BuyOrderRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type BuyOrderReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

type CancelOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	// userID just to log it
	UserId uint64 `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *CancelOrderRequest) Reset()                    { *m = CancelOrderRequest{} }
func (m *CancelOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderRequest) ProtoMessage()               {}
func (*CancelOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

type CancelOrderReply struct {
	Cancelled    bool   `protobuf:"varint,1,opt,name=cancelled" json:"cancelled,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CancelOrderReply) Reset()                    { *m = CancelOrderReply{} }
func (m *CancelOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderReply) ProtoMessage()               {}
func (*CancelOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

type AddCardRequest struct {
	User    *UserInfo `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	Gateway string    `protobuf:"bytes,2,opt,name=gateway" json:"gateway,omitempty"`
}

func (m *AddCardRequest) Reset()                    { *m = AddCardRequest{} }
func (m *AddCardRequest) String() string            { return proto.CompactTextString(m) }
func (*AddCardRequest) ProtoMessage()               {}
func (*AddCardRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *AddCardRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type AddCardReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,33,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *AddCardReply) Reset()                    { *m = AddCardReply{} }
func (m *AddCardReply) String() string            { return proto.CompactTextString(m) }
func (*AddCardReply) ProtoMessage()               {}
func (*AddCardReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

func init() {
	proto.RegisterType((*UsualData)(nil), "payment.UsualData")
	proto.RegisterType((*ChatMessageNewOrder)(nil), "payment.ChatMessageNewOrder")
	proto.RegisterType((*ChatMessagePaymentFinished)(nil), "payment.ChatMessagePaymentFinished")
	proto.RegisterType((*ChatMessageOrderCancelled)(nil), "payment.ChatMessageOrderCancelled")
	proto.RegisterType((*OrderData)(nil), "payment.OrderData")
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*GetOrderRequest)(nil), "payment.GetOrderRequest")
	proto.RegisterType((*GetOrderReply)(nil), "payment.GetOrderReply")
	proto.RegisterType((*UpdateServiceDataRequest)(nil), "payment.UpdateServiceDataRequest")
	proto.RegisterType((*UpdateServiceDataReply)(nil), "payment.UpdateServiceDataReply")
	proto.RegisterType((*PaymentNotification)(nil), "payment.PaymentNotification")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*UserInfo)(nil), "payment.UserInfo")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*CancelOrderRequest)(nil), "payment.CancelOrderRequest")
	proto.RegisterType((*CancelOrderReply)(nil), "payment.CancelOrderReply")
	proto.RegisterType((*AddCardRequest)(nil), "payment.AddCardRequest")
	proto.RegisterType((*AddCardReply)(nil), "payment.AddCardReply")
	proto.RegisterEnum("payment.Event", Event_name, Event_value)
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
	GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*GetOrderReply, error)
	UpdateServiceData(ctx context.Context, in *UpdateServiceDataRequest, opts ...grpc.CallOption) (*UpdateServiceDataReply, error)
	AddCard(ctx context.Context, in *AddCardRequest, opts ...grpc.CallOption) (*AddCardReply, error)
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

func (c *paymentServiceClient) GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*GetOrderReply, error) {
	out := new(GetOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/GetOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) UpdateServiceData(ctx context.Context, in *UpdateServiceDataRequest, opts ...grpc.CallOption) (*UpdateServiceDataReply, error) {
	out := new(UpdateServiceDataReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/UpdateServiceData", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) AddCard(ctx context.Context, in *AddCardRequest, opts ...grpc.CallOption) (*AddCardReply, error) {
	out := new(AddCardReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/AddCard", in, out, c.cc, opts...)
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
	GetOrder(context.Context, *GetOrderRequest) (*GetOrderReply, error)
	UpdateServiceData(context.Context, *UpdateServiceDataRequest) (*UpdateServiceDataReply, error)
	AddCard(context.Context, *AddCardRequest) (*AddCardReply, error)
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

func _PaymentService_GetOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).GetOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/GetOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetOrder(ctx, req.(*GetOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_UpdateServiceData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateServiceDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).UpdateServiceData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/UpdateServiceData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).UpdateServiceData(ctx, req.(*UpdateServiceDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_AddCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddCardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).AddCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/AddCard",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).AddCard(ctx, req.(*AddCardRequest))
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
		{
			MethodName: "GetOrder",
			Handler:    _PaymentService_GetOrder_Handler,
		},
		{
			MethodName: "UpdateServiceData",
			Handler:    _PaymentService_UpdateServiceData_Handler,
		},
		{
			MethodName: "AddCard",
			Handler:    _PaymentService_AddCard_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1192 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x57, 0xdd, 0x72, 0xdb, 0xd4,
	0x13, 0xaf, 0xfc, 0xed, 0xf5, 0x97, 0x72, 0xfa, 0x6f, 0xab, 0xe6, 0x0f, 0x03, 0x11, 0x04, 0x3a,
	0x61, 0xda, 0x61, 0xca, 0x15, 0x03, 0xc3, 0x8c, 0x22, 0x29, 0xad, 0xa7, 0x41, 0x0e, 0xb2, 0x5d,
	0x26, 0x57, 0x1a, 0x55, 0x3a, 0x69, 0x34, 0xd8, 0x92, 0x91, 0xec, 0x64, 0x32, 0x03, 0xe5, 0xb3,
	0x6f, 0xc0, 0x15, 0xbc, 0x09, 0x4f, 0xc3, 0xa3, 0xb0, 0xe7, 0x1c, 0xc9, 0x92, 0xe3, 0x38, 0xcd,
	0x45, 0xb9, 0xd3, 0xfe, 0xf6, 0x7b, 0xf7, 0xec, 0xae, 0x0d, 0x9d, 0x99, 0x7b, 0x31, 0xa5, 0xe1,
	0xfc, 0xd1, 0x2c, 0x8e, 0xe6, 0x11, 0xa9, 0xa7, 0xa4, 0xfa, 0x5a, 0x82, 0xe6, 0x38, 0x59, 0xb8,
	0x13, 0xc3, 0x9d, 0xbb, 0xe4, 0x53, 0x68, 0xfa, 0x41, 0x4c, 0xbd, 0x79, 0x10, 0x85, 0x8a, 0xf4,
	0xbe, 0xf4, 0xa0, 0xfb, 0x98, 0x3c, 0xca, 0x34, 0x8d, 0x8c, 0x63, 0xe7, 0x42, 0xe4, 0x63, 0xe8,
	0x79, 0x51, 0x78, 0x46, 0xe3, 0xc4, 0x65, 0xb4, 0x13, 0xf8, 0x4a, 0x09, 0xf5, 0x2a, 0x76, 0xb7,
	0x08, 0xf7, 0x7d, 0xf2, 0x2e, 0xc0, 0x94, 0x26, 0x89, 0xfb, 0x92, 0x32, 0x99, 0x32, 0x97, 0x69,
	0xa6, 0x48, 0xdf, 0x57, 0x13, 0xb8, 0xad, 0x9f, 0xba, 0xf3, 0xaf, 0x05, 0x60, 0xd1, 0xf3, 0x41,
	0xec, 0xd3, 0x98, 0xdc, 0x81, 0x1a, 0xba, 0x67, 0x1a, 0x12, 0xd7, 0xa8, 0x22, 0x85, 0xc6, 0xee,
	0x42, 0xcd, 0x9d, 0x46, 0x8b, 0x70, 0x9e, 0x3a, 0x4b, 0x29, 0xf2, 0x10, 0x1a, 0xde, 0x22, 0x8e,
	0x69, 0xe8, 0x5d, 0x70, 0x17, 0xdd, 0xc7, 0x5b, 0xcb, 0xf0, 0xf5, 0x94, 0x61, 0x2f, 0x45, 0xd4,
	0x7f, 0x24, 0xd8, 0x2e, 0x78, 0x3d, 0x12, 0x92, 0x07, 0x41, 0x18, 0x24, 0xa7, 0xd4, 0xdf, 0xe4,
	0x5c, 0x81, 0x7a, 0xb2, 0xf0, 0x3c, 0xd4, 0xe1, 0x3e, 0x1a, 0x76, 0x46, 0x32, 0xce, 0x89, 0x1b,
	0x4c, 0x16, 0x31, 0x55, 0x2a, 0x82, 0x93, 0x92, 0x85, 0x80, 0xab, 0x1b, 0x03, 0xae, 0xbd, 0x31,
	0xe0, 0xd5, 0xfe, 0xd4, 0x6f, 0xd0, 0x1f, 0xf5, 0x19, 0xdc, 0x2f, 0x64, 0xc8, 0x8b, 0xaa, 0xbb,
	0xa1, 0x47, 0x27, 0x93, 0xcd, 0x09, 0xde, 0x83, 0xfa, 0x22, 0xa1, 0x71, 0xde, 0xcb, 0x1a, 0x23,
	0xb1, 0x49, 0x7f, 0x96, 0xa1, 0xc9, 0x4d, 0xf0, 0xc7, 0x92, 0xe7, 0x24, 0x6d, 0xcc, 0xa9, 0xf4,
	0xe6, 0x9c, 0xb0, 0x68, 0x2f, 0xdd, 0x39, 0x3d, 0x77, 0x45, 0xcb, 0x9a, 0x76, 0x46, 0x16, 0xe3,
	0xa8, 0x14, 0xe3, 0x20, 0x3b, 0xd0, 0xc6, 0x8f, 0xb3, 0xc0, 0xa3, 0x4e, 0xe8, 0x4e, 0x29, 0xaf,
	0x69, 0xd3, 0x6e, 0xa5, 0x98, 0x85, 0x50, 0x51, 0xc4, 0xc7, 0x60, 0x79, 0x71, 0x73, 0x11, 0x1e,
	0x3f, 0x9a, 0x9f, 0x50, 0xd7, 0x67, 0xe6, 0xeb, 0xc2, 0x3c, 0x23, 0xd1, 0xfc, 0x03, 0x90, 0x93,
	0xd3, 0x68, 0xe6, 0x78, 0x6e, 0xec, 0x3b, 0xe1, 0x62, 0xfa, 0x82, 0xc6, 0x0a, 0x70, 0xfd, 0x2e,
	0xc3, 0x75, 0x84, 0x2d, 0x8e, 0x92, 0x5d, 0xc0, 0x67, 0x3e, 0x9d, 0x06, 0x49, 0xc2, 0xde, 0xfe,
	0x09, 0xa5, 0x4a, 0x8b, 0x5b, 0xea, 0xe4, 0xe8, 0x01, 0xa5, 0xe4, 0x13, 0xd8, 0x2a, 0x88, 0x25,
	0xd1, 0x22, 0xf6, 0xa8, 0xd2, 0xe6, 0x92, 0x72, 0xce, 0x18, 0x72, 0x9c, 0xbc, 0x03, 0x4d, 0x2f,
	0xeb, 0x90, 0xd2, 0xe1, 0xcf, 0x28, 0x07, 0x58, 0xb5, 0x98, 0x06, 0xd6, 0x52, 0xe9, 0x8a, 0x6a,
	0xa5, 0xa4, 0xfa, 0x25, 0x10, 0x3d, 0xa6, 0x58, 0x3a, 0xde, 0x21, 0x9b, 0x7e, 0xbf, 0xa0, 0xc9,
	0x9c, 0x7c, 0x04, 0x15, 0x9e, 0x3f, 0x6b, 0x51, 0xab, 0xf0, 0x58, 0x96, 0x6d, 0xb4, 0x39, 0x5f,
	0xdd, 0x81, 0xde, 0x13, 0x3a, 0x5f, 0x51, 0xed, 0x42, 0x69, 0xf9, 0x32, 0xf0, 0x4b, 0xfd, 0x1c,
	0x3a, 0xb9, 0xc8, 0x6c, 0x72, 0x81, 0x75, 0xaa, 0x46, 0x8c, 0xba, 0xc6, 0xb8, 0x10, 0x50, 0x4d,
	0x50, 0xc6, 0x33, 0xf4, 0x43, 0x87, 0x79, 0xfd, 0x37, 0xb8, 0x21, 0xf7, 0xa1, 0x11, 0xd2, 0x73,
	0xd1, 0xb5, 0x92, 0x48, 0x11, 0x69, 0xa6, 0xa1, 0x2a, 0x70, 0xf7, 0x0a, 0x33, 0x18, 0x8a, 0xfa,
	0x97, 0x04, 0xb7, 0xd3, 0xf1, 0xb5, 0xa2, 0x79, 0x70, 0x12, 0x78, 0x7c, 0xef, 0xac, 0x19, 0xcf,
	0xca, 0x51, 0xba, 0xbe, 0x1c, 0xe4, 0x43, 0xa8, 0xd2, 0x33, 0x56, 0x64, 0xb1, 0x45, 0xba, 0x4b,
	0x41, 0x93, 0xa1, 0xb6, 0x60, 0xa2, 0xb5, 0x5e, 0x10, 0x9e, 0x45, 0xdf, 0xe1, 0x1b, 0x5d, 0x7d,
	0xa8, 0x9d, 0x14, 0x1e, 0x8b, 0xb9, 0x09, 0x41, 0x5e, 0x69, 0x0d, 0x2b, 0xde, 0xe5, 0xc8, 0x76,
	0xd1, 0x63, 0x1c, 0x47, 0x71, 0x3a, 0x32, 0xbd, 0xdc, 0x23, 0x43, 0x13, 0x5b, 0x70, 0xc9, 0x07,
	0xd0, 0xe1, 0x1f, 0x4e, 0xba, 0x3a, 0xd3, 0x99, 0x69, 0x73, 0x30, 0x9d, 0x72, 0xb5, 0x0f, 0x0d,
	0xee, 0x39, 0x3c, 0x89, 0xb8, 0x9f, 0x19, 0xf7, 0xd3, 0x44, 0x3f, 0xb3, 0x8d, 0xc3, 0x4d, 0xfe,
	0x07, 0xd5, 0xd9, 0x69, 0x14, 0x66, 0x16, 0x05, 0xa1, 0x06, 0xd0, 0xdb, 0x5f, 0x5c, 0xac, 0xbc,
	0x8b, 0x0d, 0x5b, 0x63, 0x17, 0x2a, 0xcc, 0x52, 0x5a, 0xda, 0x7c, 0xe4, 0xb3, 0x48, 0x6c, 0xce,
	0xde, 0x3c, 0xee, 0xea, 0x2b, 0xe8, 0xe4, 0xae, 0x58, 0x89, 0x70, 0x86, 0x63, 0x2a, 0x56, 0x99,
	0xb3, 0x88, 0x27, 0x69, 0x12, 0xad, 0x0c, 0x1b, 0xc7, 0x93, 0xb7, 0x5a, 0x35, 0x03, 0x07, 0x88,
	0xcf, 0xd9, 0x4d, 0xb2, 0x2d, 0x94, 0xb1, 0xba, 0xb2, 0x23, 0x7f, 0xc0, 0x5e, 0x17, 0xad, 0xb0,
	0x44, 0x56, 0x46, 0x5a, 0xba, 0x3c, 0xd2, 0x6f, 0x33, 0x87, 0x6f, 0xa0, 0xab, 0xf9, 0x3e, 0xdb,
	0x50, 0x59, 0xfc, 0x59, 0x5b, 0xa4, 0x1b, 0xb7, 0xa5, 0xb4, 0xda, 0x96, 0x1f, 0xa1, 0xbd, 0x34,
	0xf9, 0x1f, 0x77, 0x65, 0x67, 0x3d, 0xa3, 0xbd, 0x7d, 0xa8, 0xf2, 0x99, 0x23, 0x2d, 0xa8, 0x8b,
	0x21, 0xf2, 0xe5, 0x5b, 0xa4, 0x03, 0xcd, 0xe5, 0x19, 0x93, 0x25, 0x46, 0xe2, 0xf4, 0x1f, 0xe0,
	0xb1, 0x45, 0xb2, 0x84, 0x6f, 0x1e, 0x90, 0x1c, 0x8a, 0xab, 0x2c, 0x97, 0xf7, 0x76, 0xa1, 0x91,
	0x1d, 0x1e, 0x52, 0x87, 0xb2, 0x3d, 0xde, 0x47, 0x13, 0xf8, 0x31, 0x1e, 0x1a, 0xa8, 0x8c, 0x1f,
	0xfa, 0xe0, 0x48, 0x2e, 0xed, 0xfd, 0x51, 0x82, 0x9a, 0x88, 0x90, 0xd4, 0xa0, 0x34, 0x78, 0x86,
	0x42, 0x32, 0xb4, 0xfb, 0xd6, 0x73, 0xed, 0xb0, 0x6f, 0x38, 0x86, 0x36, 0xd2, 0x84, 0x2b, 0x63,
	0xdf, 0x39, 0xd0, 0xfa, 0x87, 0xa6, 0x81, 0xae, 0xb6, 0xa0, 0xa3, 0x1d, 0xda, 0xa6, 0x66, 0x1c,
	0x3b, 0x47, 0xda, 0x31, 0x42, 0x65, 0x06, 0xe1, 0xa7, 0xa3, 0x6b, 0x96, 0x6e, 0x1e, 0x32, 0xa9,
	0x0a, 0x9e, 0x4a, 0xa2, 0x59, 0x83, 0xd1, 0x53, 0xd3, 0x76, 0x06, 0x47, 0xa6, 0xe5, 0x0c, 0x6c,
	0xc3, 0xb4, 0xe5, 0x2a, 0x3e, 0xae, 0xad, 0x4c, 0x3b, 0x17, 0xaf, 0x31, 0x0b, 0x63, 0xeb, 0x99,
	0x35, 0xf8, 0xd6, 0x72, 0x4c, 0xdb, 0x1e, 0xd8, 0xf2, 0x2b, 0xd2, 0x83, 0x56, 0xdf, 0xea, 0x8f,
	0x32, 0xc7, 0x3f, 0x21, 0x00, 0xcc, 0x4b, 0x4a, 0xff, 0x2c, 0x61, 0xd2, 0x4d, 0xfd, 0xa9, 0x36,
	0x72, 0x0c, 0x54, 0x93, 0x7f, 0x91, 0x98, 0x80, 0x3e, 0xe8, 0x5b, 0x43, 0x01, 0xfc, 0x2a, 0xa1,
	0xd5, 0x36, 0x3a, 0x19, 0x39, 0x5c, 0xcd, 0x34, 0xe5, 0xdf, 0x38, 0x64, 0x9b, 0x07, 0x63, 0xcb,
	0x48, 0xfd, 0xfc, 0x2e, 0x61, 0xc6, 0x2d, 0x4b, 0x1b, 0x0d, 0x33, 0xc3, 0xaf, 0xa5, 0xbd, 0x87,
	0x98, 0xf1, 0xf2, 0xf7, 0x1e, 0xc6, 0xa1, 0x1f, 0xf6, 0x4d, 0x61, 0x66, 0x88, 0x15, 0xca, 0x01,
	0xdb, 0xd4, 0x9f, 0xcb, 0xd2, 0xe3, 0xbf, 0xcb, 0xd0, 0x4d, 0x57, 0x71, 0xba, 0xa6, 0xc9, 0x13,
	0x94, 0xc9, 0xf7, 0x1f, 0xf9, 0x7f, 0xfe, 0x73, 0x60, 0xed, 0x60, 0x6d, 0xdf, 0xbf, 0x9a, 0xc9,
	0x96, 0xfc, 0x2d, 0xf2, 0x15, 0x34, 0xb2, 0x15, 0x41, 0x94, 0xa5, 0xe0, 0xa5, 0x05, 0xb5, 0x7d,
	0xf7, 0x0a, 0x8e, 0xd0, 0x67, 0x81, 0xe4, 0xc3, 0x59, 0x0c, 0x64, 0x6d, 0xf0, 0x8b, 0x81, 0x5c,
	0x9a, 0x67, 0x11, 0x48, 0x76, 0x0b, 0x0b, 0x81, 0x5c, 0xba, 0xa0, 0x85, 0x40, 0x56, 0x0e, 0x27,
	0xea, 0x1f, 0xc3, 0xd6, 0xda, 0x25, 0x23, 0x3b, 0xf9, 0x70, 0x6e, 0x38, 0x96, 0xdb, 0xef, 0x5d,
	0x27, 0x22, 0x4c, 0x7f, 0x01, 0xf5, 0x74, 0x5e, 0xc9, 0xbd, 0xa5, 0xf4, 0xea, 0x52, 0xd8, 0xbe,
	0xb3, 0xce, 0xe0, 0xca, 0x2f, 0x6a, 0xfc, 0xef, 0xc1, 0x67, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff,
	0x40, 0x43, 0xa9, 0x0d, 0x2f, 0x0c, 0x00, 0x00,
}
