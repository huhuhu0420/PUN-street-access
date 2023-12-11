package usecase

import (
	"context"

	"github.com/PUArallelepiped/PUN-street-Universal-Access/domain"
	"github.com/PUArallelepiped/PUN-street-Universal-Access/swagger"
	"github.com/sirupsen/logrus"
)

type cartUsecase struct {
	cartRepo domain.CartRepo
}

func NewCartUsecase(cartRepo domain.CartRepo) domain.CartUsecase {
	return &cartUsecase{
		cartRepo: cartRepo,
	}
}

func (cu *cartUsecase) GetAllHistoryById(ctx context.Context, id int64) (*[]swagger.HistoryInfo, error) {
	historyArray, err := cu.cartRepo.GetAllHistoryById(ctx, id)
	if err != nil {
		return nil, err
	}

	return historyArray, nil
}

func (cu *cartUsecase) GetRunOrderByID(ctx context.Context, id int64) (*[]swagger.RunOrderInfo, error) {
	runOrderArray, err := cu.cartRepo.GetRunOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return runOrderArray, nil
}

func (cu *cartUsecase) DeleteProduct(ctx context.Context, customerId int64, productId int64) error {
	// delete product
	storeId, err := cu.cartRepo.DeleteProduct(ctx, customerId, productId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// check now the delete product's store exits in cart
	exits, err := cu.cartRepo.IsExitsCartByStoreCartId(ctx, customerId, storeId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// if not exits delete the store order status = 0
	if !exits {
		err = cu.cartRepo.DeleteOrder(ctx, customerId, storeId)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}

func (cu *cartUsecase) AddProductToCart(ctx context.Context, customerId int64, cartInfo *swagger.CartInfo) error {
	// add or update product in cart
	err := cu.cartRepo.AddProductToCart(ctx, customerId, cartInfo)
	if err != nil {
		logrus.Error(err)
		return err
	}
	// check order exits by store_id
	exits, err := cu.cartRepo.IsExitsOrderByStoreCartId(ctx, customerId, cartInfo.StoreId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// if not exits add order status = 0
	if !exits {
		err := cu.cartRepo.AddOrderByCartInfo(ctx, customerId, cartInfo.StoreId)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}

func (cu *cartUsecase) GetHistoryCart(ctx context.Context, customerId int64, cartId int64, storeId int64) (*swagger.StoreOrderInfo, error) {
	storeOrder, err := cu.cartRepo.GetHistoryCart(ctx, customerId, cartId, storeId)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	var total_price int64
	// cal event discount
	for _, product := range storeOrder.ProductOrder {
		quantity := product.ProductQuantity - int64(product.ProductQuantity/int64(product.EventDiscountMaxQuantity+1))
		total_price += quantity * product.ProductPrice
	}

	// cal shipping discount
	storeOrder.ShippingDiscountBool = false
	if storeOrder.ShippingDiscount != nil {
		if total_price > storeOrder.ShippingDiscount.DiscountMaxPrice && storeOrder.ShippingDiscount.DiscountMaxPrice > 0 {
			storeOrder.ShippingDiscountBool = true
		} else {
			total_price += storeOrder.StoreShippingFee
			storeOrder.ShippingDiscountBool = false
		}
	}

	// cal shipping discount
	storeOrder.SeasoningDiscountBool = false
	if storeOrder.SeasoningDiscount != nil {
		if storeOrder.SeasoningDiscount.DiscountPercentage != 0 {
			total_price = int64(float32(total_price) * float32(storeOrder.SeasoningDiscount.DiscountPercentage) / 100)
			storeOrder.SeasoningDiscountBool = true
		} else {
			storeOrder.SeasoningDiscountBool = false
		}
	}

	storeOrder.TotalPrice = total_price

	return storeOrder, nil
}

func (cu *cartUsecase) GetCurrentCartsByUserID(ctx context.Context, id int64) (*swagger.CartOrderInfo, error) {
	ids, err := cu.cartRepo.GetCurrentCartID(ctx, id)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	var cartOrder swagger.CartOrderInfo
	for _, id := range ids {
		storeOrder, err := cu.GetHistoryCart(ctx, id.UserID, id.CartID, id.StoreID)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		cartOrder.StoreOrderInfoArray = append(cartOrder.StoreOrderInfoArray, *storeOrder)
	}

	var realTotalPrice int64
	for _, storeOrder := range cartOrder.StoreOrderInfoArray {
		realTotalPrice += storeOrder.TotalPrice
	}
	cartOrder.RealTotalPrice = realTotalPrice

	return &cartOrder, nil
}

func (cu *cartUsecase) Checkout(ctx context.Context, customerId int64, storeId int64) error {
	cartOrder, err := cu.GetCurrentCartsByUserID(ctx, customerId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = cu.cartRepo.UpdateOrderInfo(ctx, customerId, storeId, cartOrder.RealTotalPrice)
	if err != nil {
		logrus.Error(err)
		return err
	}

	//cartID not work
	//fix me
	err = cu.cartRepo.AddUserCurrentCart(ctx, customerId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
