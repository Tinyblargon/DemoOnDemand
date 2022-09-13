package ssh

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yahoo/vssh"
)

func New(user, password, addr string, port uint16) (vs *vssh.VSSH, err error) {
	vs = vssh.New().Start()
	sshConfig := vssh.GetConfigUserPass(user, password)
	for _, addr := range []string{addr + ":" + strconv.Itoa(int(port))} {
		err = vs.AddClient(addr, sshConfig, vssh.SetMaxSessions(1))
	}
	vs.Wait()
	return
}

type Response struct {
	OutTxt   string
	ErrText  string
	Exitcode int
	Err      error
}

func command(vs *vssh.VSSH, cmd string) (response []Response) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeout, _ := time.ParseDuration("6s")
	respChan := vs.Run(ctx, cmd, timeout)

	for resp := range respChan {
		var err error
		if err = resp.Err(); err != nil {
			continue
		}
		outTxt, errTxt, _ := resp.GetText(vs)

		currentResponse := Response{
			OutTxt:   outTxt,
			ErrText:  errTxt,
			Exitcode: resp.ExitStatus(),
			Err:      err,
		}
		response = append(response, currentResponse)
	}
	return
}

type NetworkInterfaces struct {
	Name string
	Mac  string
}

func ListNetworkInterfaces(vs *vssh.VSSH) (*[]NetworkInterfaces, error) {
	response := command(vs, "ls /sys/class/net")
	err := returnResponseError(response)
	if err != nil {
		return nil, err
	}
	interfaces := strings.Split(removeNewline(tabToSpaces(response[0].OutTxt)), "  ")
	networkInterfaces := make([]NetworkInterfaces, len(interfaces))
	for i, e := range interfaces {
		networkInterfaces[i].Name = e
	}
	return &networkInterfaces, nil
}

func GetMacAddresses(vs *vssh.VSSH, networkInterfaces *[]NetworkInterfaces) (err error) {
	for i := range *networkInterfaces {
		response := command(vs, "cat /sys/class/net/"+(*networkInterfaces)[i].Name+"/address")
		err = returnResponseError(response)
		if err != nil {
			return
		}
		(*networkInterfaces)[i].Mac = removeNewline(response[0].OutTxt)
	}
	return
}

func returnResponseError(response []Response) error {
	for _, e := range response {
		if e.Exitcode > 0 {
			if e.Err != nil {
				return e.Err
			}
			if e.ErrText != "" {
				return errors.New(e.ErrText)
			}
			if e.OutTxt != "" {
				return errors.New(e.OutTxt)
			}
		}
	}
	return nil
}

func tabToSpaces(text string) string {
	return strings.ReplaceAll(text, "\t", "  ")
}

func removeNewline(text string) string {
	return strings.ReplaceAll(text, "\n", "")
}
