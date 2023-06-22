package rbd

import (
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"

	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

type API interface {
	Create(vb *rbdtypes.VolumeBinding) error
	Remove(vb *rbdtypes.VolumeBinding) error
	Resize(vb *rbdtypes.VolumeBinding) error
	Exists(vb *rbdtypes.VolumeBinding) (bool, error)
	Close()
}

func New() (API, error) {
	conn, err := rados.NewConn()
	if err != nil {
		return nil, err
	}

	// 打开默认的配置文件（/etc/ceph/ceph.conf）
	if err := conn.ReadDefaultConfigFile(); err != nil {
		return nil, err
	}
	if err := conn.Connect(); err != nil {
		return nil, err
	}
	return &api{conn: conn}, nil
}

type api struct {
	conn *rados.Conn
}

func (a *api) Close() {
	a.conn.Shutdown()
}

func (a *api) Create(vb *rbdtypes.VolumeBinding) error {
	ctx, err := a.conn.OpenIOContext(vb.Pool)
	if err != nil {
		return err
	}
	defer ctx.Destroy()

	// 这里使用默认配置创建，也可以根据自己需求，指定image的features
	return rbd.CreateImage(ctx, vb.Image, uint64(vb.SizeInBytes), rbd.NewRbdImageOptions())
}

func (a *api) Remove(vb *rbdtypes.VolumeBinding) error {
	ctx, err := a.conn.OpenIOContext(vb.Pool)
	if err != nil {
		return err
	}
	defer ctx.Destroy()

	return rbd.RemoveImage(ctx, vb.Image)
}

func (a *api) Resize(vb *rbdtypes.VolumeBinding) error {
	ctx, err := a.conn.OpenIOContext(vb.Pool)
	if err != nil {
		return err
	}
	defer ctx.Destroy()

	img, err := rbd.OpenImage(ctx, vb.Image, rbd.NoSnapshot)
	if err != nil {
		return err
	}
	return img.Resize(uint64(vb.SizeInBytes))
}

func (a *api) Exists(vb *rbdtypes.VolumeBinding) (bool, error) {
	ctx, err := a.conn.OpenIOContext(vb.Pool)
	if err != nil {
		return false, err
	}
	defer ctx.Destroy()
	_, err = rbd.OpenImageReadOnly(ctx, vb.Image, rbd.NoSnapshot)
	if err != nil {
		if err == rbd.ErrNotFound {
			return false, nil
		}
		return false, err
	} else {
		return true, nil
	}
}

func (a *api) Clone(vb, dest *rbdtypes.VolumeBinding) error {
	ctx, err := a.conn.OpenIOContext(vb.Pool)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	img, err := rbd.OpenImageReadOnly(ctx, vb.Image, rbd.NoSnapshot)
	if err != nil {
		return err
	}
	destctx := ctx
	if vb.Pool != dest.Pool {
		destctx, err = a.conn.OpenIOContext(vb.Pool)
		if err != nil {
			return err
		}
		defer destctx.Destroy()
	}
	return rbd.CloneFromImage(img, vb.Image, destctx, dest.Image, rbd.NewRbdImageOptions())
}
