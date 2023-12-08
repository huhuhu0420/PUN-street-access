/*
 * PUN street Universal Access - OpenAPI 3.0
 *
 * pua
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type OrderInfoShort struct {
	UserId int64 `json:"user_id"`

	CartId int64 `json:"cart_id"`

	StoreId int64 `json:"store_id"`

	OrderDate string `json:"order_date"`

	UserName string `json:"user_name"`
}
