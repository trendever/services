package fixtures

//go:generate mockgen -package=fixtures -destination=mock_repo.go payments/models Repo
//go:generate mockgen -package=fixtures -destination=mock_gw.go payments/models Gateway
