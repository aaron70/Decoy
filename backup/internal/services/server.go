package services

import (
	"os"
	"path"

	"github.com/aaron70/decoy/internal/model"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/repositories"
)

type Server struct {
	repo     repositories.Repository[string, model.ServerSpec]
	BasePath string
	SpecPath string
}

func NewServer(basePath string, repo repositories.Repository[string, model.ServerSpec]) (*Server, error) {
	specPath := path.Join(basePath, "specs")
	if err := os.MkdirAll(specPath, 0755); err != nil {
		return nil, nil
	}
	return &Server{
		repo:     repo,
		BasePath: basePath,
		SpecPath: specPath,
	}, nil
}

func (svc Server) Save(name, contents string) (model.ServerSpec, error) {
	path := svc.fileSpecPath(name)
	saved, err := svc.repo.Save(name, model.ServerSpec{
		Name: name,
		Spec: path,
	})
	if err != nil {
		return model.ServerSpec{}, err
	}
	err = os.WriteFile(path, []byte(contents), os.ModePerm)
	if err != nil {
		return model.ServerSpec{}, err
	}
	return saved, nil
}

func (svc Server) Update(name, contents string) (model.ServerSpec, error) {
	path := svc.fileSpecPath(name)
	updated, err := svc.repo.Update(name, model.ServerSpec{
		Name: name,
		Spec: path,
	})
	if err != nil {
		return model.ServerSpec{}, err
	}
	err = os.WriteFile(path, []byte(contents), os.ModePerm)
	if err != nil {
		return model.ServerSpec{}, err
	}
	return updated, nil
}

func (svc Server) Get(name string) (model.ServerSpec, error) {
	spec, err := svc.repo.Get(name)
	if err != nil {
		return model.ServerSpec{}, err
	}

	path := svc.fileSpecPath(name)
	contents, err := os.ReadFile(path)
	if err != nil {
		return model.ServerSpec{}, err
	}
	spec.Spec = string(contents)

	return spec, nil
}

func (svc Server) GetAll() ([]model.ServerSpec, error) {
	return svc.repo.GetAll()
}

func (svc Server) Delete(name string) (model.ServerSpec, error) {
	path := svc.fileSpecPath(name)
	err := os.Remove(path)
	if err != nil {
		return model.ServerSpec{}, errors.NewError(nil, err, "Couldn't remove the %s file", path)
	}
	return svc.repo.Delete(name)
}

func (svc Server) fileSpecPath(name string) string { return path.Join(svc.SpecPath, name) }
