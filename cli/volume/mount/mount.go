package mount

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	clibazil "bazil.org/bazil/cli"
	"bazil.org/bazil/cliutil/flagx"
	"bazil.org/bazil/cliutil/subcommands"
	"bazil.org/bazil/control/wire"
	"github.com/golang/protobuf/proto"
)

type mountCommand struct {
	subcommands.Description
	Arguments struct {
		VolumeName string
		Mountpoint flagx.AbsPath
	}
}

func (cmd *mountCommand) Run() error {
	req := wire.VolumeMountRequest{
		VolumeName: cmd.Arguments.VolumeName,
		Mountpoint: cmd.Arguments.Mountpoint.String(),
	}
	buf, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	resp, err := clibazil.Bazil.Control.Post(
		"http+unix://bazil/control/volumeMount",
		"binary/x.bazil.control.volumeMountRequest",
		bytes.NewReader(buf),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		if len(buf) == 0 {
			buf = []byte(resp.Status)
		}
		return errors.New(string(buf))
	}
	return nil
}

var mount = mountCommand{
	Description: "mount a volume",
}

func init() {
	subcommands.Register(&mount)
}
