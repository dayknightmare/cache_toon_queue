package utils

import (
	"errors"
	"net/http"

	"github.com/Vupy/cache-toon-queue/src/structs/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ShouldBind(context *gin.Context, data interface{}) bool {
	should := true

	if err := context.ShouldBindJSON(data); err != nil {
		var ve validator.ValidationErrors
		should = false

		if errors.As(err, &ve) {
			out := make([]models.ErrorMessage, len(ve))

			for i, fe := range ve {
				out[i] = models.ErrorMessage{
					Field:   fe.Field(),
					Message: "This field is required",
				}
			}

			context.AbortWithStatusJSON(http.StatusBadRequest, models.ResponseError{
				Message: out,
			})
		}
	}

	return should
}
