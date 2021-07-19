package ovirtclient

import (
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

type Logger interface {
	ovirtclientlog.Logger
}
