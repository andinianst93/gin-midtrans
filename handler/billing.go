package handler

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransRequest struct {
	UserId   int    `json:"user_id" binding:"required"`
	Amount   int64  `json:"amount" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
	ItemName string `json:"item_name" binding:"required"`
}

type WebResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type MidtransResponse struct {
	Token       string `json:"token"`
	RedirectUrl string `json:"redirect_url"`
}

type MidtransController interface {
	Create(c *gin.Context)
}

type MidtransControllerImpl struct {
	Validate *validator.Validate
}

func NewMidtransControllerImpl(validate *validator.Validate) *MidtransControllerImpl {
	return &MidtransControllerImpl{
		Validate: validate,
	}
}

func (controller *MidtransControllerImpl) Create(c *gin.Context) {
	validate := validator.New()

	var request MidtransRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		panic(err)
	}

	err := validate.Struct(request)
	if err != nil {
		panic(err)
	}

	// request midtrans
	var snapClient = snap.Client{}
	snapClient.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Sandbox)

	// user id
	user_id := strconv.Itoa(request.UserId)

	// customer
	custAddress := &midtrans.CustomerAddress{
		FName:       "John",
		LName:       "Doe",
		Phone:       "081234567890",
		Address:     "Baker Street 97th",
		City:        "Jakarta",
		Postcode:    "16000",
		CountryCode: "IDN",
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  "MID-User-" + user_id + "-" + request.ItemID,
			GrossAmt: request.Amount,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName:    "John",
			LName:    "Doe",
			Email:    "john@doe.com",
			Phone:    "081234567890",
			BillAddr: custAddress,
			ShipAddr: custAddress,
		},
		EnabledPayments: snap.AllSnapPaymentType,
		Items: &[]midtrans.ItemDetails{
			{
				ID:    "Property-" + request.ItemID,
				Qty:   1,
				Price: request.Amount,
				Name:  request.ItemName,
			},
		},
	}

	response, errSnap := snapClient.CreateTransaction(req)
	if errSnap != nil {
		panic(errSnap.GetRawError())
	}

	midtransReponse := MidtransResponse{
		Token:       response.Token,
		RedirectUrl: response.RedirectURL,
	}

	webResponse := WebResponse{
		Code:   http.StatusOK,
		Status: "OK",
		Data:   midtransReponse,
	}

	c.JSON(http.StatusOK, webResponse)
}

func ErrorHandle() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		if validationErrors(c, err) {
			return
		}

		internalServerError(c, err)
	})
}

func validationErrors(c *gin.Context, err interface{}) bool {
	if exception, ok := err.(validator.ValidationErrors); ok {
		var ve validator.ValidationErrors
		out := make([]ErrorResponse, len(ve))
		if errors.As(exception, &ve) {
			for _, fe := range ve {
				out = append(out, ErrorResponse{
					Field:   fe.Field(),
					Message: MessageForTag(fe.Tag()),
				})
			}
		}
		webResponse := WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   out,
		}
		c.JSON(http.StatusBadRequest, webResponse)
		c.Abort()

		return true
	} else {
		return false
	}
}

func internalServerError(c *gin.Context, err interface{}) {
	webResponse := WebResponse{
		Code:   http.StatusInternalServerError,
		Status: "INTERNAL SERVER ERROR",
		Data:   err,
	}

	c.JSON(http.StatusInternalServerError, webResponse)
	c.Abort()
}

func MessageForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	}
	return ""
}
