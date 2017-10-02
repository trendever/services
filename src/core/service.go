package main

import (
	"common/db"
	"common/log"
	"core/conf"
	"core/models"
	"core/project"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"golang.org/x/net/context"
	"os"
	"proto/core"
	"utils/rpc"
)

func main() {

	app := cli.NewApp()
	app.Name = "Trendever website core and management interface"
	app.Usage = "Shop"
	app.Version = "0.0.1"

	svc := project.Service{}
	testFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "method",
			Value: "user_leads",
			Usage: "Method for cli test command",
		},
		cli.IntFlag{
			Name:  "user_id",
			Value: 0,
			Usage: "User id for cli test command",
		},
		cli.IntFlag{
			Name:  "lead_id",
			Value: 0,
			Usage: "Lead id for cli test command",
		},
		cli.StringFlag{
			Name:  "instagram_pk",
			Value: "",
			Usage: "Instagram pk",
		},
		cli.IntFlag{
			Name:  "product_id",
			Value: 0,
			Usage: "Product id",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "Run the http server",
			Action: svc.Run,
		},
		{
			Name:   "migrate",
			Usage:  "Migrate Database",
			Action: svc.AutoMigrate,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "drop, d",
				},
			},
		},
		{
			Name:  "test_local",
			Usage: "Cli client local tests",
			Flags: testFlags,
			Action: func(cli *cli.Context) {
				conf.Init()
				db.Init(&conf.GetSettings().DB)
				switch cli.String("method") {
				case "json_product":
					var product models.Product
					err := db.New().Unscoped().Preload("InstagramImages").Where("id = ?", cli.Int("product_id")).Find(&product).Error
					if err != nil {
						fmt.Println(err)
						return
					}
					json, err := json.Marshal(product.Encode())
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println(string(json))
				}
			},
		},
		{
			Name:  "test",
			Usage: "Cli client for tests grpc",
			Action: func(cli *cli.Context) {
				conn := rpc.Connect(cli.Args()[0])
				c := core.NewLeadServiceClient(conn)
				p := core.NewProductServiceClient(conn)
				var resp interface{}
				var err error
				switch cli.String("method") {
				case "user_leads":
					resp, err = c.GetUserLeads(context.Background(), &core.UserLeadsRequest{
						UserId: uint64(cli.Int("user_id")),
						Role: []core.LeadUserRole{
							core.LeadUserRole_CUSTOMER,
							core.LeadUserRole_SUPPLIER,
							core.LeadUserRole_SELLER,
						},
					})
				case "lead_status":
					resp, err = c.SetLeadStatus(context.Background(), &core.SetLeadStatusRequest{
						LeadId: uint64(cli.Int("lead_id")),
						UserId: uint64(cli.Int("user_id")),
						Event:  core.LeadStatusEvent_PROGRESS,
					})
				case "create_lead":
					resp, err = c.CreateLead(context.Background(), &core.Lead{
						//Id: int64(cli.Int("lead_id")),
						CustomerId:  int64(cli.Int("user_id")),
						Source:      "website",
						InstagramPk: cli.String("instagram_pk"),
						ProductId:   int64(cli.Int("product_id")),
					})
				case "get_lead_by_id":
					resp, err = c.GetLead(context.Background(), &core.GetLeadRequest{
						UserId: uint64(cli.Int("user_id")),
						SearchBy: &core.GetLeadRequest_Id{
							Id: uint64(cli.Int("lead_id")),
						},
					})
				case "get_lead_by_chat_id":
					resp, err = c.GetLead(context.Background(), &core.GetLeadRequest{
						UserId: uint64(cli.Int("user_id")),
						SearchBy: &core.GetLeadRequest_ConversationId{
							ConversationId: uint64(cli.Int("lead_id")),
						},
					})
				case "call_supplier":
					resp, err = c.CallSupplier(context.Background(), &core.CallSupplierRequest{
						LeadId: uint64(cli.Int("lead_id")),
					})
				case "products":
					resp, err = p.SearchProducts(context.Background(), &core.SearchProductRequest{
						Limit:      10,
						IsSaleOnly: true,
						UserId:     uint64(cli.Int("user_id")),
					})
				default:
					log.Fatal(errors.New("unknown method for test"))
				}
				js, _ := json.MarshalIndent(resp, "", " ")
				fmt.Println(string(js), err)
			},
			Flags: testFlags,
		},
		{
			Name:  "conf",
			Usage: "Show current config",
			Action: func(cli *cli.Context) {
				j, _ := json.MarshalIndent(conf.GetSettings(), "", " ")
				fmt.Println(string(j))
			},
		},
	}
	app.Run(os.Args)
}
