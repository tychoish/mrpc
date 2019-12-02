package shell

import (
	"context"
	"io"

	"github.com/evergreen-ci/mrpc"
	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/pkg/errors"
)

type shellService struct {
	mrpc.Service
}

// NewShellService returns a service for mongo shell clients listening on the
// given host and port.
func NewShellService(host string, port int) (mrpc.Service, error) {
	s := &shellService{Service: mrpc.NewBasicService(host, port)}
	if err := s.registerHandlers(); err != nil {
		return nil, errors.Wrap(err, "could not register handlers")
	}
	return s, nil
}

// MakeShellService takes an existing mrpc.Service and adds support for mongo
// shell clients.
func MakeShellService(service mrpc.Service) (mrpc.Service, error) {
	s := &shellService{Service: service}
	if err := s.registerHandlers(); err != nil {
		return nil, errors.Wrap(err, "could not register handlers")
	}
	return s, nil
}

// Constants representing required shell commands.
const (
	isMasterCommand   = "isMaster"
	whatsMyURICommand = "whatsmyuri"
	// The shell sends commands with different casing so we need two different
	// handlers for the different "buildinfo" commands
	buildInfoCommand               = "buildInfo"
	BuildinfoCommand               = "buildinfo"
	getCmdLineOptsCommand          = "getCmdLineOpts"
	getLogCommand                  = "getLog"
	getFreeMonitoringStatusCommand = "getFreeMonitoringStatus"
	replSetGetStatusCommand        = "replSetGetStatus"
	listCollectionsCommand         = "listCollections"
)

func (s *shellService) registerHandlers() error {
	// TODO: support OP_MSG handlers
	for name, handler := range map[string]mrpc.HandlerFunc{
		// Required initialization commands
		isMasterCommand:                s.isMaster,
		whatsMyURICommand:              s.whatsMyURI,
		BuildinfoCommand:               s.buildInfo,
		buildInfoCommand:               s.buildInfo,
		getLogCommand:                  s.getLog,
		replSetGetStatusCommand:        s.replSetGetStatus,
		getFreeMonitoringStatusCommand: s.getFreeMonitoringStatus,
		listCollectionsCommand:         s.listCollections,
		getCmdLineOptsCommand:          s.getCmdLineOpts,
	} {
		for _, opType := range []mongowire.OpType{mongowire.OP_COMMAND, mongowire.OP_MSG} {
			if err := s.RegisterOperation(&mongowire.OpScope{
				Type:    opType,
				Command: name,
			}, handler); err != nil {
				return errors.Wrapf(err, "could not register %s handler for %s", opType.String(), name)
			}
		}
	}

	return nil
}

const opMsgWireVersion = 6

func (s *shellService) isMaster(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := ResponseToMessage(makeIsMasterResponse(0, opMsgWireVersion))
	if err != nil {
		WriteErrorResponse(ctx, w, msg.Header().OpCode, errors.Wrap(err, "could not make response"), isMasterCommand)
		return
	}
	WriteResponse(ctx, w, msg.Header().OpCode, resp, isMasterCommand)
}

func (s *shellService) whatsMyURI(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := ResponseToMessage(makeWhatsMyURIResponse(s.Address()))
	if err != nil {
		WriteErrorResponse(ctx, w, msg.Header().OpCode, errors.Wrap(err, "could not make response"), whatsMyURICommand)
		return
	}
	WriteResponse(ctx, w, msg.Scope(), resp, whatsMyURICommand)
}

func (s *shellService) buildInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	WriteResponse(ctx, w, msg.Header().OpCode, resp, buildInfoCommand)
}

func (s *shellService) getCmdLineOpts(ctx context.Context, w io.Writer, msg mongowire.Message) {
	WriteNotOKResponse(ctx, w, msg.Header().OpCode, getCmdLineOptsCommand)
}

func (s *shellService) getFreeMonitoringStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	WriteNotOKResponse(ctx, w, msg.Header().OpCode, getFreeMonitoringStatusCommand)
}

func (s *shellService) getLog(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := ResponseToMessage(makeGetLogResponse([]string{}))
	if err != nil {
		return
	}
	WriteResponse(ctx, w, msg.Header().OpCode, resp, getLogCommand)
}

func (s *shellService) listCollections(ctx context.Context, w io.Writer, msg mongowire.Message) {
	WriteNotOKResponse(ctx, w, msg.Header().OpCode, listCollectionsCommand)
}

func (s *shellService) replSetGetStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	WriteNotOKResponse(ctx, w, msg.Header().OpCode, replSetGetStatusCommand)
}
