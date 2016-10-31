package main

import (
	"fmt"
	"golang.org/x/net/context"
	proto "proto/trendcoin"
	"time"
	"utils/db"
	"utils/log"
)

type TrendcoinServer struct {
	requestChan chan *TransactionsRequest
}

var server *TrendcoinServer

type TransactionsRequest struct {
	Transactions TransactionsSlice
	AnswerChan   chan *proto.MakeTransactionsReply
}

func GetTrendcoinServer() *TrendcoinServer {
	if server == nil {
		server = &TrendcoinServer{
			requestChan: make(chan *TransactionsRequest),
		}
		server.subscribe()
		go server.loop()
	}
	return server
}

func (s *TrendcoinServer) Stop() {
	close(s.requestChan)
}

func (s *TrendcoinServer) Balance(_ context.Context, in *proto.BalanceRequest) (*proto.BalanceReply, error) {
	account := Account{UserID: in.UserId}
	res := db.New().Find(&account)
	if res.Error != nil && !res.RecordNotFound() {
		err := fmt.Errorf("failed to load balance for user %v: %v", account.UserID, res.Error)
		log.Error(err)
		return &proto.BalanceReply{
			Error: err.Error(),
		}, nil
	}
	return &proto.BalanceReply{
		Balance: account.Balance,
	}, nil
}

// synchronization matters here for real, we will perform all transactions in single gorutine
// while we don't expect thousands of transactions per second it should be fine
func (s *TrendcoinServer) MakeTransactions(_ context.Context, in *proto.MakeTransactionsRequest) (*proto.MakeTransactionsReply, error) {
	req := &TransactionsRequest{}
	for _, protoTrans := range in.Transactions {
		trans := DecodeTransactionData(protoTrans)
		if err := trans.Validate(); err != nil {
			log.Debug("invalid transaction request {%+v}: %v", protoTrans, err)
			return &proto.MakeTransactionsReply{
				Error: err.Error(),
			}, nil
		}
		req.Transactions = append(req.Transactions, trans)
	}
	req.AnswerChan = make(chan *proto.MakeTransactionsReply)
	s.requestChan <- req
	ans := <-req.AnswerChan
	return ans, nil
}

func (s *TrendcoinServer) loop() {
	for req := range s.requestChan {
		ans := &proto.MakeTransactionsReply{}
		err := req.Transactions.Perform()
		if err != nil {
			log.Errorf("failed to perform trasactions: %v", err)
			ans.Error = err.Error()
		}
		for _, trans := range req.Transactions {
			ans.TransactionIds = append(ans.TransactionIds, trans.ID)
		}
		req.AnswerChan <- ans
	}
}

func (s *TrendcoinServer) TransactionLog(_ context.Context, in *proto.TransactionLogRequest) (*proto.TransactionLogReply, error) {
	var transactions TransactionsSlice
	scope := db.New().Where("source = ? OR destination = ?", in.UserId, in.UserId)
	if in.Asc {
		scope = scope.Order("id ASC")
	} else {
		scope = scope.Order("id DESC")
	}
	if in.Before != 0 {
		scope = scope.Where("created_at < ?", time.Unix(in.Before, 0))
	}
	if in.After != 0 {
		scope = scope.Where("created_at >= ?", time.Unix(in.After, 0))
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	scope = scope.Limit(in.Limit + 1)
	if in.Offset != 0 {
		scope = scope.Offset(in.Offset)
	}
	err := scope.Find(&transactions).Error
	if err != nil {
		log.Errorf("failed to load transactions log for user %v: %v", in.UserId, err)
		return &proto.TransactionLogReply{
			Error: err.Error(),
		}, nil
	}
	hasMore := false
	if uint64(len(transactions)) > in.Limit {
		hasMore = true
		transactions = transactions[:in.Limit]
	}
	return &proto.TransactionLogReply{
		Transactions: transactions.Encode(),
		HasMore:      hasMore,
	}, nil
}
