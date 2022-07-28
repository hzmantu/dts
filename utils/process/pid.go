package process

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type Pid struct {
	pidFile string
}

func NewPid(pidFile string) *Pid {
	return &Pid{pidFile}
}

func (p *Pid) SaveFile() {
	pid := os.Getpid()
	_ = ioutil.WriteFile(p.pidFile, []byte(strconv.Itoa(pid)), 0666)
}

func (p *Pid) RemoveFile() {
	_ = os.Remove(p.pidFile)
}

func (p *Pid) Kill() {
	pid := p.GetPid()
	if pid != "" {
		_ = exec.Command("kill", pid).Run()
	}
	p.RemoveFile()
}

func (p *Pid) GetPid() string {
	data, _ := ioutil.ReadFile(p.pidFile)
	return string(data)
}
