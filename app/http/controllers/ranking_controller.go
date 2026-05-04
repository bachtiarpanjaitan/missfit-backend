package controllers

import (
	"fmt"
	"missfit/app/services"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
)

type RankingController struct {
	packageService services.PackageServiceInterface
}

func NewRankingController(packageService services.PackageServiceInterface) *RankingController {
	return &RankingController{
		packageService: packageService,
	}
}

func (r *RankingController) GlobalRankings(ctx http.Context) http.Response {
	limit := ctx.Request().QueryInt("limit", 10)
	ranking, err := r.packageService.GetGlobalRankings(limit)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}
	return utils.Ok(ctx, "loaded", ranking)
}

func (r *RankingController) MyRank(ctx http.Context) http.Response {
	user := utils.User(ctx)
	ranking, err := r.packageService.GetMyRank(user.Id)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}
	return utils.Ok(ctx, "loaded", ranking)
}

func (r *RankingController) PackageRank(ctx http.Context) http.Response {
	packageId := ctx.Request().Route("package_id")
	fmt.Println(utils.ToJson(packageId))
	if packageId == "" {
		return utils.BadRequest(ctx, "Parameter ID Paket tidak boleh kosong", nil)
	}
	ranking, err := r.packageService.GetPackageRank(packageId)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}
	return utils.Ok(ctx, "loaded", ranking)
}
