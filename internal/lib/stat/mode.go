package stat

import "io/fs"

// FileMode is a Linux file mode
type FileMode uint16

const (
	// ModeType is the mask for the type bits
	ModeType = 0o170000

	ModeSocket      FileMode = 0o140000 // socket
	ModeSymlink     FileMode = 0o120000 // symbolic link
	ModeRegularFile FileMode = 0o100000 // regular file
	ModeBlockDevice FileMode = 0o060000 // block device
	ModeDirectory   FileMode = 0o040000 // directory
	ModeCharDevice  FileMode = 0o020000 // character device
	ModeNamedPipe   FileMode = 0o010000 // FIFO
	ModeSetuid      FileMode = 0o004000 // set UID bit
	ModeSetgid      FileMode = 0o002000 // set-group-ID bit
	ModeStick       FileMode = 0o001000 // sticky bit

	// ModePerm is the mask for the permission bits
	ModePerm FileMode = 0o777 // Unix permission bits
)

func (m FileMode) IsRegular() bool {
	return m&ModeRegularFile == ModeRegularFile
}

func (m FileMode) IsDir() bool {
	return m&ModeDirectory == ModeDirectory
}

func (m FileMode) IsSymlink() bool {
	return m&ModeSymlink == ModeSymlink
}

func (m FileMode) Perm() FileMode {
	return m & ModePerm
}

// ToFsFileMode returns the equivalent fs.FileMode of a stat.FileMode.
func (m FileMode) ToFsFileMode() fs.FileMode {
	// See fillFileStatFromSys in stat_linux.go
	return fs.FileMode(m & ModePerm)
}
