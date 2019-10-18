package robot

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/token"
	"github.com/goharbor/harbor/src/pkg/token/claim"
	"github.com/pkg/errors"
	"time"
)

var (
	// RobotCtr is a global variable for the default robot account controller implementation
	RobotCtr = NewController(NewDefaultRobotAccountManager())
)

// Controller to handle the requests related with robot account
type Controller interface {
	// GetRobotAccount ...
	GetRobotAccount(id int64) (*model.Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(robotReq *model.RobotCreate) (*model.Robot, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(id int64) error

	// UpdateRobotAccount ...
	UpdateRobotAccount(r *model.Robot) error

	// ListRobotAccount ...
	ListRobotAccount(query *q.Query) ([]*model.Robot, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager Manager
}

// NewController ...
func NewController(robotMgr Manager) Controller {
	return &DefaultAPIController{
		manager: robotMgr,
	}
}

// GetRobotAccount ...
func (d *DefaultAPIController) GetRobotAccount(id int64) (*model.Robot, error) {
	return d.manager.GetRobotAccount(id)
}

// CreateRobotAccount ...
func (d *DefaultAPIController) CreateRobotAccount(robotReq *model.RobotCreate) (*model.Robot, error) {

	var deferDel error
	// Token duration in minutes
	tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
	expiresAt := time.Now().UTC().Add(tokenDuration).Unix()
	createdName := common.RobotPrefix + robotReq.Name

	// first to add a robot account, and get its id.
	robot := &model.Robot{
		Name:        createdName,
		Description: robotReq.Description,
		ProjectID:   robotReq.ProjectID,
		ExpiresAt:   expiresAt,
		Visible:     robotReq.Visible,
	}
	id, err := d.manager.CreateRobotAccount(robot)
	if err != nil {
		return nil, err
	}

	// generate the token, and return it with response data.
	// token is not stored in the database.
	opt := token.DefaultTokenOptions()
	rClaims := &claim.Robot{
		TokenID:     id,
		ProjectID:   robotReq.ProjectID,
		Access:      robotReq.Access,
		PolicyCheck: robotReq.PolicyCheck,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: expiresAt,
			Issuer:    opt.Issuer,
		},
	}
	tk, err := token.New(opt, rClaims)
	if err != nil {
		deferDel = err
		return nil, fmt.Errorf("failed to valid parameters to generate token for robot account, %v", err)
	}
	rawTk, err := tk.Raw()
	if err != nil {
		deferDel = err
		return nil, fmt.Errorf("failed to sign token for robot account, %v", err)
	}

	defer func(deferDel error) {
		if deferDel != nil {
			if err := d.manager.DeleteRobotAccount(id); err != nil {
				log.Error(errors.Wrap(err, fmt.Sprintf("failed to delete the robot account: %d", id)))
			}
		}
	}(deferDel)

	robot.Token = rawTk
	robot.ID = id
	return robot, nil
}

// DeleteRobotAccount ...
func (d *DefaultAPIController) DeleteRobotAccount(id int64) error {
	return d.manager.DeleteRobotAccount(id)
}

// UpdateRobotAccount ...
func (d *DefaultAPIController) UpdateRobotAccount(r *model.Robot) error {
	return d.manager.UpdateRobotAccount(r)
}

// ListRobotAccount ...
func (d *DefaultAPIController) ListRobotAccount(query *q.Query) ([]*model.Robot, error) {
	return d.manager.ListRobotAccount(query)
}
