package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"strconv"
	"strings"
)

type User struct {
	Name   string
	Uid    *int
	Group  string
	Gid    *int
	System *bool
	Home   string
	Shell  string
}

type UserClient interface {
	Get(ctx context.Context, uid int) (*User, error)
	Create(ctx context.Context, user User) (int, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, uid int) error
}

func NewUserClient(s system.System) UserClient {
	return &userClient{
		s: s,
	}
}

var (
	ErrUser = errors.New("user resource")

	ErrUserNotFound = errors.Join(ErrUser, errors.New("user not found"))

	ErrUserNameExists = errors.Join(ErrUser, errors.New("user name exists"))

	ErrUserUidExists = errors.Join(ErrUser, errors.New("user uid exists"))

	ErrUserGroupNotFound = errors.Join(ErrUser, errors.New("primary group not found"))

	ErrUserUnexpected = errors.Join(ErrUser, errors.New("unexpected error"))
)

const (
	codeUserUnexpected = 1

	codeUserNotFound = 2

	codeUserUidExists = 4

	codeUserGroupNotFound = 6

	codeUserNameExists = 9
)

type userClient struct {
	s system.System
}

func (c *userClient) Get(ctx context.Context, uid int) (*User, error) {
	cmd := NewCommand(fmt.Sprintf(`getent passwd %[1]d`, uid))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, errors.Join(ErrUserUnexpected, err)
	}

	if res.ExitCode == codeUserNotFound {
		return nil, ErrUserNotFound
	}
	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return nil, ErrUserUnexpected
	}

	parsedUser, err := parsePasswdEntry(res.Stdout)
	if err != nil {
		return nil, ErrUserUnexpected
	}

	userSystem := parsedUser.Uid < 1000

	groupCmd := NewCommand(fmt.Sprintf(`getent group %[1]d`, parsedUser.Gid))
	resGroup, err := ExecuteCommand(ctx, c.s, groupCmd)
	if err != nil {
		return nil, errors.Join(ErrUserUnexpected, err)
	}

	if resGroup.ExitCode == codeGroupNotFound {
		return nil, ErrUserGroupNotFound
	}
	if resGroup.ExitCode != 0 || len(resGroup.Stdout) == 0 {
		return nil, ErrUserUnexpected
	}

	parsedGroup, err := parseGroupEntry(resGroup.Stdout)
	if err != nil {
		return nil, ErrGroupUnexpected
	}

	user := &User{
		Name:   parsedUser.Name,
		Uid:    to.IntPtr(parsedUser.Uid),
		Group:  parsedGroup.Name,
		Gid:    to.IntPtr(parsedUser.Gid),
		System: to.BoolPtr(userSystem),
		Home:   parsedUser.Home,
		Shell:  parsedUser.Shell,
	}

	return user, nil
}

func (c *userClient) Create(ctx context.Context, u User) (int, error) {
	var args []string

	if u.Uid != nil {
		args = append(args, fmt.Sprintf("--uid %d", to.Int(u.Uid)))
	}

	if u.System != nil && *u.System {
		args = append(args, "--system")
	} else {
		// Prevent create home directory for regular users
		args = append(args, "--no-create-home")
	}

	if u.Gid != nil {
		// Primary group by gid
		args = append(args, fmt.Sprintf("--gid %d", to.Int(u.Gid)))
	} else if u.Group != "" {
		// Primary group by group name
		args = append(args, fmt.Sprintf("--gid %s", u.Group))
	} else {
		// Default primary group
		args = append(args, "--no-user-group")
	}

	if u.Home != "" {
		args = append(args, fmt.Sprintf("--home %s", u.Home))
	}

	if u.Shell != "" {
		args = append(args, fmt.Sprintf("--shell %s", u.Shell))
	}

	args = append(args, u.Name)

	cmd := NewCommand(fmt.Sprintf(`useradd %[1]s && getent passwd %[2]s`, strings.Join(args, " "), u.Name))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return -1, errors.Join(ErrUserUnexpected, err)
	}

	switch res.ExitCode {
	case codeUserUidExists:
		// uid already exists
		return -1, ErrUserUidExists
	case codeUserGroupNotFound:
		// group does not exist
		return -1, ErrUserGroupNotFound
	case codeUserNameExists:
		// username already exists
		return -1, ErrUserNameExists
	}

	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return -1, ErrUserUnexpected
	}

	createdUser, err := parsePasswdEntry(res.Stdout)
	if err != nil {
		return -1, ErrUserUnexpected
	}

	return createdUser.Uid, nil
}

func (c *userClient) Update(ctx context.Context, u User) error {
	if u.Uid == nil {
		return errors.Join(ErrUserUnexpected, errors.New("update requires uid"))
	}

	var args []string

	if u.Name != "" {
		// Update name
		args = append(args, fmt.Sprintf(`--login '%s'`, u.Name))
	}

	if u.Home != "" {
		// Update home
		args = append(args, fmt.Sprintf(`--home '%s'`, u.Home))
	}

	if u.Shell != "" {
		// Update shell
		args = append(args, fmt.Sprintf(`--shell '%s'`, u.Shell))
	}

	if u.Gid != nil || u.Group != "" {
		if u.Gid != nil {
			// Update primary group by gid
			args = append(args, fmt.Sprintf("--gid %d", to.Int(u.Gid)))
		} else if u.Group != "" {
			// Update primary group by group name
			args = append(args, fmt.Sprintf("--gid %s", u.Group))
		}
	}

	if len(args) == 0 {
		return nil
	}

	usermodCmd := fmt.Sprintf(`usermod %s "${user}"`, strings.Join(args, " "))
	cmd := NewCommand(fmt.Sprintf(`_do() { uid=$1; user=$(getent passwd $uid | cut -d: -f1); [ ! -z "${user}" ] || return %[2]d; %[3]s; return $?; }; _do '%[1]d';`, to.Int(u.Uid), codeUserNotFound, usermodCmd))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return errors.Join(ErrUserUnexpected, err)
	}

	switch res.ExitCode {
	case 0:
		// Success
		break
	case codeUserNotFound:
		// User not found
		return ErrUserNotFound
	case codeUserNameExists:
		// Username not unique
		return ErrUserNameExists
	default:
		return ErrUserUnexpected
	}

	return nil
}

func (c *userClient) Delete(ctx context.Context, uid int) error {
	// Note: userdel will also remove the primary group of the user
	cmd := NewCommand(fmt.Sprintf(`_do() { uid=$1; user=$(getent passwd $uid | cut -d: -f1); [ ! -z "${user}" ] || return %[2]d; userdel "${user}"; return $?; }; _do '%[1]d';`, uid, codeUserNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return errors.Join(ErrUserUnexpected, err)
	}

	// https://github.com/shadow-maint/shadow/blob/dc9fc048de56aa7b6eaf80b1c068a8b5d59b1bf0/src/userdel.c#L77
	switch res.ExitCode {
	case 0:
		// Success
		break
	case codeUserNotFound:
		// Not interpreted as error because this is the desired state
		break
	default:
		return errors.Join(ErrUser, fmt.Errorf("failed to delete user with uid %d", uid))
	}

	return nil
}

type passwdEntry struct {
	Name  string
	Uid   int
	Gid   int
	Home  string
	Shell string
}

func parsePasswdEntry(data []byte) (*passwdEntry, error) {
	parts := strings.Split(strings.TrimSpace(string(data)), ":")
	if len(parts) != 7 || parts[0] == "" || parts[2] == "" || parts[3] == "" {
		return nil, ErrUserUnexpected
	}

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, ErrUserUnexpected
	}

	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, ErrUserUnexpected
	}

	return &passwdEntry{
		Name:  parts[0],
		Uid:   uid,
		Gid:   gid,
		Home:  parts[5],
		Shell: parts[6],
	}, nil
}
