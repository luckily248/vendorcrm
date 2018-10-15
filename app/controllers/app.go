package controllers

import (
	"github.com/revel/revel"
	"vendorcrm/app/models"
	"gopkg.in/mgo.v2/bson"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}
func (c App) Memberpage() revel.Result {
	return c.Render()
}

func (c App) Pipeline() revel.Result {
	suppliers,err:=models.GetSuppliers()
	if err!=nil{
		return c.Render();
	}
	var	tbc []models.Supplier
	var	sc []models.Supplier
	var	sr []models.Supplier
	var	ar []models.Supplier
	var	as []models.Supplier
	var tbcl int
	var scl int
	var srl int
	var arl int
	var asl int
	for _,supplier :=range suppliers{
		switch supplier.Pstate{
			case 0:
				tbc=append(tbc, supplier)
			case 1:
				sc=append(sc, supplier)
			case 2:
				sr=append(sr, supplier)
			case 3:
				ar=append(ar, supplier)
			case 4:
				as=append(as, supplier)
			default:
			
		}
	}
	tbcl = len(tbc)
	scl = len(sc)
	srl	= len(sr)
	arl = len(ar)
	asl = len(as)
	return c.Render(tbc,tbcl,sc,scl,sr,srl,ar,arl,as,asl)
}

func (c App) Products() revel.Result {
	return c.Render()
}

func (c App) Suppliers() revel.Result {
	suppliers,err:=models.GetSuppliers()
	if err!=nil{
		return c.Render();
	}
	return c.Render(suppliers)
}

func (c App) Account() revel.Result {
	return c.Render()
}

func (c App) Supplierpage(id string) revel.Result {
	var (
		supplier   models.Supplier
		err        error
		supplierID bson.ObjectId
	)
	
	if id == "" {
		println("id empty")
		return c.Redirect("/App/suppliers")
	}
	supplierID, err = convertToObjectIdHex(id)
	if err != nil {
		println("err:" + err.Error())
		return c.Redirect("/App/suppliers")
	}
	supplier, err = models.GetSupplier(supplierID)
	if err != nil {
		println("err:" + err.Error())
		return c.Redirect("/App/suppliers")
	}
	return c.Render(supplier)
}

