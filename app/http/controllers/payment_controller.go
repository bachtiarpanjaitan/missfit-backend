package controllers

import (
	"time"

	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"

	"github.com/goravel/framework/contracts/http"
)

type PaymentController struct {
	// Dependent services
}

func NewPaymentController() *PaymentController {
	return &PaymentController{
		// Inject services
	}
}

func (r *PaymentController) InitiateFree(ctx http.Context) http.Response {
	data, err := utils.ValidateRequest(ctx, map[string]string{
		"packageId": "required|min_len:1",
	})
	// return utils.DdResponseJson(ctx, data)
	if err != nil {
		return err.(http.Response)
	}
	packageId := data["packageId"]
	user, err := utils.AuthUser(ctx)
	if err != nil {
		return err.(http.Response)
	}

	quizPackage := models.QuizPackage{}
	err = facades.Orm().Query().Where("id", packageId.(string)).First(&quizPackage)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	if quizPackage.Id == "" {
		return utils.BadRequest(ctx, "Package not found", nil)
	}

	if !quizPackage.IsFree {
		return utils.BadRequest(ctx, "Package is not free", nil)
	}

	userPurchasedPackages := models.UserPurchasedPackage{}
	err = facades.Orm().Query().Where("user_id", user.Id).Where("quiz_package_id", packageId.(string)).First(&userPurchasedPackages)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	if userPurchasedPackages.Id != "" {
		return utils.BadRequest(ctx, "You have already purchased this package", nil)
	}

	payment := models.UserPurchasedPackage{
		UserId:        user.Id,
		QuizPackageId: packageId.(string),
		TransactionId: "",
		PurchasedDate: time.Now(),
		IsActive:      true,
		ExpiredDate:   time.Now().AddDate(0, 1, 0),
	}

	err = facades.Orm().Query().Create(&payment)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal server error", err)
	}

	return utils.Ok(ctx, "Payment initiated successfully", payment)
}
