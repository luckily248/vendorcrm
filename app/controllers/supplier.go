package controllers

import (
	_ "encoding/json"
	"errors"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"vendorcrm/app/models"
)

type SupplierController struct {
	*revel.Controller
}

func (c SupplierController) Index() revel.Result {
	var (
		suppliers []models.Supplier
		err       error
	)
	suppliers, err = models.GetSuppliers()
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}
	c.Response.Status = 200
	return c.RenderJSON(suppliers)
}

func (c SupplierController) Show(id string) revel.Result {
	var (
		supplier   models.Supplier
		err        error
		supplierID bson.ObjectId
	)

	if id == "" {
		errResp := buildErrResponse(errors.New("Invalid supplier id format"), "400")
		c.Response.Status = 400
		return c.RenderJSON(errResp)
	}

	supplierID, err = convertToObjectIdHex(id)
	if err != nil {
		errResp := buildErrResponse(errors.New("Invalid supplier id format"), "400")
		c.Response.Status = 400
		return c.RenderJSON(errResp)
	}

	supplier, err = models.GetSupplier(supplierID)
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}

	c.Response.Status = 200
	return c.RenderJSON(supplier)
}

func (c SupplierController) Create() revel.Result {
	var (
		supplier models.Supplier
		err      error
	)

	err = c.Params.BindJSON(&supplier)
	if err != nil {
		errResp := buildErrResponse(err, "403")
		c.Response.Status = 403
		return c.RenderJSON(errResp)
	}

	supplier, err = models.AddSupplier(supplier)
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}
	c.Response.Status = 201
	return c.RenderJSON(supplier)
}

func (c SupplierController) Update() revel.Result {
	var (
		supplier models.Supplier
		err      error
	)
	err = c.Params.BindJSON(&supplier)
	if err != nil {
		errResp := buildErrResponse(err, "400")
		c.Response.Status = 400
		return c.RenderJSON(errResp)
	}

	err = supplier.UpdateSupplier()
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}
	return c.RenderJSON(supplier)
}

func (c SupplierController) Delete(id string) revel.Result {
	var (
		err        error
		supplier   models.Supplier
		supplierID bson.ObjectId
	)
	if id == "" {
		errResp := buildErrResponse(errors.New("Invalid supplier id format"), "400")
		c.Response.Status = 400
		return c.RenderJSON(errResp)
	}

	supplierID, err = convertToObjectIdHex(id)
	if err != nil {
		errResp := buildErrResponse(errors.New("Invalid supplier id format"), "400")
		c.Response.Status = 400
		return c.RenderJSON(errResp)
	}

	supplier, err = models.GetSupplier(supplierID)
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}
	err = supplier.DeleteSupplier()
	if err != nil {
		errResp := buildErrResponse(err, "500")
		c.Response.Status = 500
		return c.RenderJSON(errResp)
	}
	c.Response.Status = 204
	return c.RenderJSON(nil)
}
