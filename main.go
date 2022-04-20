package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type CheckoutProduct struct {
	ID          int  `json:"id"`
	Quantity    int  `json:"quantity"`
	UnitAmount  int  `json:"unit_amount"`
	TotalAmount int  `json:"total_amount"`
	Discount    int  `json:"discount"`
	IsGift      bool `json:"is_gift"`
}

type Checkout struct {
	TotalAmount             int               `json:"total_amount"`
	TotalAmountWithDiscount int               `json:"total_amount_with_discount"`
	TotalDiscount           int               `json:"total_discount"`
	Products                []CheckoutProduct `json:"products"`
}

type ProductsRequest struct {
	Products []ProductItem `json:"products"`
}

type ProductItem struct {
	ID       int `json:"id"`
	Quantity int `json:"quantity"`
}

type ProductInventory struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Amount      int    `json:"amount"`
	IsGift      bool   `json:"is_gift"`
}

func main() {
	handleRequests()
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/checkout", generateShoppingCartResume).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func generateShoppingCartResume(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var productsRequest ProductsRequest
	json.Unmarshal(reqBody, &productsRequest)

	var checkoutProducts []CheckoutProduct
	for _, productItem := range productsRequest.Products {
		productWithInfo, err := productInfo(productItem)
		if err != nil {
			log.Fatal("Error when getting productInfo: ", err)
		}

		checkoutProducts = append(checkoutProducts, productWithInfo)

	}
	checkout := checkoutPayload(checkoutProducts)

	fmt.Fprintf(w, "%+v", checkout)

	//procura produtos no json de products - pode ser tratado como um client e retorna um slice de products
	//	- precisa bater no grpc para pegar o desconto (pode ficar pra depois, assumir 0 por ora)
	//validar os product ids (pode ficar pra depois)
	// PurchaseResumeService(products):
	//	- verifica se é black friday e adiciona um gift(pode ficar pra depois)
	//	- iterar os products e somar os valores
}

func productInfo(productItem ProductItem) (CheckoutProduct, error) {
	content, err := ioutil.ReadFile("./products.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var productsInventory []ProductInventory
	err = json.Unmarshal(content, &productsInventory)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	for _, productInventory := range productsInventory {
		amount := productInventory.Amount
		quantity := productItem.Quantity

		if productItem.ID == productInventory.ID {
			checkoutProduct := CheckoutProduct{
				productInventory.ID,
				quantity,
				amount,
				quantity * amount,
				0,
				productInventory.IsGift,
			}

			return checkoutProduct, nil
		}
	}

	return CheckoutProduct{}, errors.New("Product id not available in inventory")
}

func checkoutPayload(checkoutProducts []CheckoutProduct) Checkout {
	var totalAmount int
	var totalAmountWithDiscount int
	var totalDiscount int

	for _, product := range checkoutProducts {
		totalAmount += product.TotalAmount
		totalAmountWithDiscount += product.TotalAmount - product.Discount
		totalDiscount += product.Discount
	}

	checkout := Checkout{
		totalAmount,
		totalAmountWithDiscount,
		totalDiscount,
		checkoutProducts,
	}

	return checkout
}
