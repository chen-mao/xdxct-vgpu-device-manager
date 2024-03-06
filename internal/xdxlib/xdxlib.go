package xdxlib

import (
	"github.com/chen-mao/go-xdxlib/pkg/xdxmdev"
	"github.com/chen-mao/go-xdxlib/pkg/xdxpci"
)

type Interface struct {
	Xdxpci  xdxpci.Interface
	Xdxmdev xdxmdev.Interface
}

func New() Interface {
	return Interface{
		Xdxpci:  xdxpci.New(),
		Xdxmdev: xdxmdev.New(),
	}
}
