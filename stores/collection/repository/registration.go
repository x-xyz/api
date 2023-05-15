package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type registrationImpl struct {
	q query.Mongo
}

func NewRegistration(q query.Mongo) collection.RegistrationRepo {
	return &registrationImpl{q}
}

func (im *registrationImpl) FindAll(c ctx.Ctx) ([]*collection.Registration, error) {
	res := []*collection.Registration{}

	qry := bson.M{"state": collection.RegistrationStatePending}

	if err := im.q.Search(c, domain.TableCollectionRegistrations, 0, 0, "_id", qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	}

	return res, nil
}

func (im *registrationImpl) FindOne(c ctx.Ctx, id collection.CollectionId) (*collection.Registration, error) {
	res := &collection.Registration{}

	if err := im.q.FindOne(c, domain.TableCollectionRegistrations, id, res); err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}

	return res, nil
}

func (im *registrationImpl) Create(c ctx.Ctx, value collection.Registration) error {
	if err := im.q.Insert(c, domain.TableCollectionRegistrations, value); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}

	return nil
}

func (im *registrationImpl) Patch(c ctx.Ctx, id collection.CollectionId, value collection.UpdateRegistration) error {
	if err := im.q.Patch(c, domain.TableCollectionRegistrations, id, value); err != nil {
		c.WithField("err", err).Error("q.Patch failed")
		return err
	}
	return nil
}
