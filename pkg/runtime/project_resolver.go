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
	ProjectInfo(ctx context.Context) (ProjectInfor, error)
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
	projectInfor, err := i.projectGetter(ctx, projectId)
	if err != nil {
		i.log.Error("projectresolver: failed to lookup project from id:%s", projectId)
		return context.WithValue(ctx, projectInfoKey{}, invalidProjectInfor{projectId})
	}
	i.log.Infox(ctx, "projectresolver, project model:%+v\n", projectmodel)
	return context.WithValue(ctx, projectInfoKey{}, projectInfor)
}

type ProjectInfor interface {
	GetProject() (any, error)
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

func (i invalidProjectInfor) GetProject() (any, error) {
	return nil, sdinsureerrors.NewBadParamsError(fmt.Errorf("invalid id:%s", i.retrievedProjectId))
}
