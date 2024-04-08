package runtime

import (
	"context"
	"fmt"

	sdinsureerrors "github.com/sdinsure/agent/pkg/errors"
	logger "github.com/sdinsure/agent/pkg/logger"
	"github.com/sdinsure/agent/pkg/reflection"
)

type ProjectResolver interface {
	// WithProjectInfo set project information into context
	WithProjectInfo(ctx context.Context, req interface{}) context.Context

	// ProjectInfo retrieves project information from context
	ProjectInfo(ctx context.Context) (ProjectInfor, bool)
}

func NewProjectResolver(log logger.Logger, projectGetter ProjectGetter) *projectResolver {
	return &projectResolver{
		log:           log,
		projectGetter: projectGetter,
	}
}

type ProjectGetter interface {
	GetProject(ctx context.Context, projectId string) (ProjectInfor, error)
}

type projectResolver struct {
	log           logger.Logger
	projectGetter ProjectGetter
}

func (i *projectResolver) WithProjectInfo(ctx context.Context, req interface{}) context.Context {
	i.log.Infox(ctx, "projectresolver, with project info is called\n")
	projectId, found := reflection.GetStringValue(req, "ProjectId")
	if !found || len(projectId) == 0 {
		i.log.Warnx(ctx, "projectresolver, project id not found\n")
		return context.WithValue(ctx, projectInfoKey{}, invalidProjectInfor{projectId})
	}
	i.log.Infox(ctx, "projectresolver, project Id:%+v\n", projectId)
	projectInfor, err := i.projectGetter.GetProject(ctx, projectId)
	if err != nil {
		i.log.Error("projectresolver: failed to lookup project from id:%s", projectId)
		return context.WithValue(ctx, projectInfoKey{}, invalidProjectInfor{projectId})
	}
	i.log.Infox(ctx, "projectresolver, project inform:%+v\n", projectInfor)
	return context.WithValue(ctx, projectInfoKey{}, projectInfor)
}

type ProjectInfor interface {
	GetProjectID() (string, error)
	GetProject(v any) error
}

func (i *projectResolver) ProjectInfo(ctx context.Context) (ProjectInfor, bool) {
	info := ctx.Value(projectInfoKey{})
	infov, castable := info.(ProjectInfor)
	return infov, castable
}

type (
	projectInfoKey struct{}
)

var (
	_ ProjectInfor = invalidProjectInfor{}
)

type invalidProjectInfor struct {
	retrievedProjectId string
}

func NewInvalidProjectInfor(pid string) invalidProjectInfor {
	return invalidProjectInfor{
		retrievedProjectId: pid,
	}
}

func (i invalidProjectInfor) GetProjectID() (string, error) {
	return "", sdinsureerrors.NewBadParamsError(fmt.Errorf("invalid id:%s", i.retrievedProjectId))
}

func (i invalidProjectInfor) GetProject(v any) error {
	return sdinsureerrors.NewBadParamsError(fmt.Errorf("invalid id:%s", i.retrievedProjectId))
}
