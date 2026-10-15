package services

import (
	"path"

	"github.com/aaron70/decoy/internal/model"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/repositories"
)

type Decoy struct {
	TemplateSvc *Template
	RunnerSvc   *Runner
	ServerSvc   *Server
	Decoy       *decoy.Decoy
}

func NewDecoyFS(basePath string) (*Decoy, error) {
	templateRepo, err := repositories.NewFSRepository[string, model.Template](path.Join(basePath, "templates"))
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the templates repository")
	}
	runnerRepo, err := repositories.NewFSRepository[string, model.Runner](path.Join(basePath, "runners"))
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the runners repository")
	}
	serversPath := path.Join(basePath, "servers")
	serverRepo, err := repositories.NewFSRepository[string, model.ServerSpec](serversPath)
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the runners repository")
	}
	decoy, err := decoy.NewDecoyWithSeed(0)
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the decoy instance")
	}
	serverSvc, err := NewServer(serversPath, serverRepo)
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the servers service")
	}
	return &Decoy{
		TemplateSvc: NewTemplate(templateRepo),
		RunnerSvc:   NewRunner(runnerRepo, decoy),
		ServerSvc:   serverSvc,
		Decoy:       decoy,
	}, nil
}
