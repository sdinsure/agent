package runtime

import (
	"context"
	"errors"
	"regexp"

	sdinsureerrors "github.com/sdinsure/agent/pkg/errors"
	logger "github.com/sdinsure/agent/pkg/logger"
	"github.com/sdinsure/agent/pkg/reflection"
)

type ProjectResolver interface {
	// WithProjectInfo set project information into context
	WithProjectInfo(ctx context.Context, reqPath string) context.Context

	// ProjectInfo retrieves project information from context
	ProjectInfo(ctx context.Context) (ProjectInfor, bool)
}

type ProjectInfor interface {
	GetProjectID() (string, error)
	GetProject(v any) error
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

func (i *projectResolver) WithProjectInfo(ctx context.Context, reqPath string) context.Context {
	i.log.Infox(ctx, "projectresolver: reqpath=%+v\n", reqPath)
	projectId, found := findProjectIdFromPath(reqPath)
	if !found || len(projectId) == 0 {
		i.log.Warnx(ctx, "projectresolver: project id not found from request string\n")
		return context.WithValue(ctx, projectInfoKey{}, invalidProjectInfor{})
	}
	i.log.Infox(ctx, "projectresolver: parsed project Id:%+v\n", projectId)
	projectInfor, err := i.projectGetter.GetProject(ctx, projectId)
	if err != nil {
		i.log.Errorx(ctx, "projectresolver: failed to lookup project from id:%s, err:%+v", projectId, err)
		return context.WithValue(ctx, projectInfoKey{}, invalidProjectInfor{})
	}
	i.log.Infox(ctx, "projectresolver, resolved project inform:%+v\n", projectInfor)
	return context.WithValue(ctx, projectInfoKey{}, projectInfor)
}

func findProjectIdFromPath(path string) (string, bool) {
	var re = regexp.MustCompile(`\/projects\/([a-zA-Z0-9]+)\/?`)
	matchedStrings := re.FindAllStringSubmatch(path, -1)
	if len(matchedStrings) != 1 || len(matchedStrings[0]) != 2 {
		return "", false
	}
	return matchedStrings[0][1], true
}

func findProjectIdFromStruct(req interface{}) (string, bool) {
	return reflection.GetStringValue(req, "ProjectId")
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
}

func NewInvalidProjectInfor() invalidProjectInfor {
	return invalidProjectInfor{}
}

func (i invalidProjectInfor) GetProjectID() (string, error) {
	return "", sdinsureerrors.NewBadParamsError(errors.New("invalid projectid"))
}

func (i invalidProjectInfor) GetProject(v any) error {
	return sdinsureerrors.NewBadParamsError(errors.New("invalid projectid"))
}
