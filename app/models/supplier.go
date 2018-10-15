package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"./app/models/mongodb"
)

type Supplier struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Name      string        `json:"name" bson:"name"`
	Type      int           `json:"type" bson:"type"`
	Phone     string        `json:"phone" bson:"phone"`
	Email     string        `json:"email" bson:"email"`
	Website   string        `json:"website" bson:"website"`
	Des       string        `json:"des" bson:"des"`
	Pstate    int           `json:"pstate" bson:"pstate"`
	Sstate    int           `json:"sstate" bson:"sstate"`
	Priority  int           `json:"priority" bson:"priority"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}


func newSupplierCollection() *mongodb.Collection {
	return mongodb.NewCollectionSession("suppliers")
}

// AddSupplier insert a new Supplier into database and returns
// last inserted supplier on success.
func AddSupplier(m Supplier) (supplier Supplier, err error) {
	c := newSupplierCollection()
	defer c.Close()
	m.ID = bson.NewObjectId()
	m.CreatedAt = time.Now()
	return m, c.Session.Insert(m)
}

// UpdateSupplier update a Supplier into database and returns
// last nil on success.
func (m Supplier) UpdateSupplier() error {
	c := newSupplierCollection()
	defer c.Close()

	err := c.Session.Update(bson.M{
		"_id": m.ID,
	}, bson.M{
		"$set": bson.M{
			"name": m.Name, "type": m.Type, "phone": m.Phone, "email": m.Email, "website": m.Website, "des": m.Des, "pstate": m.Pstate, "sstate": m.Sstate, "priority": m.Priority, "updatedAt": time.Now()},
	})
	return err
}

// DeleteSupplier Delete Supplier from database and returns
// last nil on success.
func (m Supplier) DeleteSupplier() error {
	c := newSupplierCollection()
	defer c.Close()

	err := c.Session.Remove(bson.M{"_id": m.ID})
	return err
}

// GetSuppliers Get all Supplier from database and returns
// list of Supplier on success
func GetSuppliers() ([]Supplier, error) {
	var (
		suppliers []Supplier
		err       error
	)

	c := newSupplierCollection()
	defer c.Close()

	err = c.Session.Find(nil).Sort("-createdAt").All(&suppliers)
	return suppliers, err
}

// GetSupplier Get a Supplier from database and returns
// a Supplier on success
func GetSupplier(id bson.ObjectId) (Supplier, error) {
	var (
		supplier Supplier
		err      error
	)

	c := newSupplierCollection()
	defer c.Close()

	err = c.Session.Find(bson.M{"_id": id}).One(&supplier)
	return supplier, err
}
