package services

import (
	"github.com/aaron70/decoy/internal/model"
	"github.com/aaron70/goaty/repositories"
)

type Template struct {
	repo repositories.Repository[string, model.Template]
}

func NewTemplate(repo repositories.Repository[string, model.Template]) *Template {
	return &Template{
		repo: repo,
	}
}

func (svc Template) Save(name, contents string) (model.Template, error) {
	return svc.repo.Save(name, model.Template{
		Name: name,
		Tmpl: contents,
	})
}

func (svc Template) Update(name, contents string) (model.Template, error) {
	return svc.repo.Update(name, model.Template{
		Name: name,
		Tmpl: contents,
	})
}

func (svc Template) Get(name string) (model.Template, error) {
	return svc.repo.Get(name)
}

func (svc Template) GetAll() ([]model.Template, error) {
	return svc.repo.GetAll()
}

func (svc Template) Delete(id string) (model.Template, error) {
	return svc.repo.Delete(id)
}
