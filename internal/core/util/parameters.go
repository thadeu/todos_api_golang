package util

import "github.com/gin-gonic/gin"

func ParamsToMap[T any](c *gin.Context) (T, error) {
	var params T

	if err := c.ShouldBindJSON(&params); err != nil {
		return params, err
	}

	return params, nil
}
