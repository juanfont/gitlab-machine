package vcd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/juanfont/gitlab-machine/pkg/drivers"
	"github.com/juanfont/gitlab-machine/pkg/ssh"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const SSHPort = 22

type VcdDriverConfig struct {
	VcdURL           string
	VcdOrg           string
	VcdVdc           string // virtual datacenter
	VcdInsecure      bool
	VcdUser          string
	VcdPassword      string
	VcdOrgVDCNetwork string
	Catalog          string
	Template         string

	NumCpus        int
	CoresPerSocket int
	MemorySizeMb   int
	VAppHREF       string
	VMHREF         string
	Description    string
	StorageProfile string

	DefaultPassword string
}

type VcdDriver struct {
	cfg           VcdDriverConfig
	client        *govcd.VCDClient
	machineName   string
	VAppHREF      string
	VMHREF        string
	adminPassword string
}

func NewVcdDriver(cfg VcdDriverConfig, machineName string) (*VcdDriver, error) {
	u, err := url.ParseRequestURI(cfg.VcdURL)
	if err != nil {
		return nil, err
	}
	c, err := newClient(*u, cfg.VcdUser, cfg.VcdPassword, cfg.VcdOrg, cfg.VcdInsecure)
	if err != nil {
		return nil, err
	}

	d := VcdDriver{
		cfg:           cfg,
		client:        c,
		machineName:   machineName,
		adminPassword: cfg.DefaultPassword,
	}

	return &d, nil
}

func (d *VcdDriver) GetMachineName() string {
	return d.machineName
}

func (d *VcdDriver) Create() error {
	log.Info().Msgf("Creating a new machine %s", d.machineName)
	org, err := d.client.GetOrgByName(d.cfg.VcdOrg)
	if err != nil {
		return err
	}
	vdc, err := org.GetVDCByName(d.cfg.VcdVdc, false)
	if err != nil {
		return err
	}

	net, err := vdc.GetOrgVdcNetworkByName(d.cfg.VcdOrgVDCNetwork, true)
	if err != nil {
		return err
	}

	catalog, err := org.GetCatalogByName(d.cfg.Catalog, true)
	if err != nil {
		return err
	}

	template, err := catalog.GetCatalogItemByName(d.cfg.Template, true)
	if err != nil {
		return err
	}
	vapptemplate, err := template.GetVAppTemplate()
	if err != nil {
		return err
	}

	var storageProfile types.Reference
	if d.cfg.StorageProfile != "" {
		storageProfile, err = vdc.FindStorageProfileReference(d.cfg.StorageProfile)
		if err != nil {
			return err
		}
	} else {
		if len(vdc.Vdc.VdcStorageProfiles.VdcStorageProfile) < 1 {
			return fmt.Errorf("no storage profile available")
		}
		storageProfile = *(vdc.Vdc.VdcStorageProfiles.VdcStorageProfile[0])
		if err != nil {
			return err
		}
	}

	networks := []*types.OrgVDCNetwork{}
	networks = append(networks, net.OrgVDCNetwork)
	task, err := vdc.ComposeVApp(
		networks,
		vapptemplate,
		storageProfile,
		d.machineName,
		d.cfg.Description,
		true)

	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	vapp, err := vdc.GetVAppByName(d.machineName, true)
	if err != nil {
		return err
	}
	d.VAppHREF = vapp.VApp.HREF

	if len(vapp.VApp.Children.VM) != 1 {
		return fmt.Errorf("VM count != 1")
	}

	vm := govcd.NewVM(&d.client.Client)
	vm.VM.HREF = vapp.VApp.Children.VM[0].HREF
	err = vm.Refresh()
	if err != nil {
		return err
	}

	d.VMHREF = vm.VM.HREF

	cWait := make(chan string, 1)
	go func() {
		for {
			status, _ := vm.GetStatus()
			if status == "POWERED_OFF" {
				break
			}
			time.Sleep(5 * time.Second)
		}

		for {
			vapp.Refresh()
			if err != nil {
				cWait <- "err"
				return
			}
			if vapp.VApp.Tasks == nil {
				time.Sleep(10) // let's give this old chap some time
				break

			}
			time.Sleep(5 * time.Second)
		}

		cWait <- "ok"
	}()

	select {
	case res := <-cWait:
		if res == "err" {
			return fmt.Errorf("error waiting for vApp deploy")
		}
	case <-time.After(15 * time.Minute):
		return fmt.Errorf("reached timeout while deploying VM")
	}

	if vm.VM.VmSpecSection == nil {
		return fmt.Errorf("VM Spec Section empty")
	}
	vm.Refresh()

	vm.VM.VmSpecSection.MemoryResourceMb.Configured = int64(d.cfg.MemorySizeMb)
	vm.VM.VmSpecSection.NumCpus = &d.cfg.NumCpus
	vm.VM.VmSpecSection.NumCoresPerSocket = &d.cfg.CoresPerSocket

	vm, err = vm.UpdateVmSpecSection(vm.VM.VmSpecSection, d.cfg.Description)
	if err != nil {
		return err
	}

	log.Debug().Msg("Configuring network")
	var netConn *types.NetworkConnection
	var netSection *types.NetworkConnectionSection
	if vm.VM.NetworkConnectionSection == nil {
		netSection = &types.NetworkConnectionSection{}
	} else {
		netSection = vm.VM.NetworkConnectionSection
	}

	if len(netSection.NetworkConnection) < 1 {
		netConn = &types.NetworkConnection{}
		netSection.NetworkConnection = append(netSection.NetworkConnection, netConn)
	}

	netConn = netSection.NetworkConnection[0]

	netConn.IPAddressAllocationMode = types.IPAllocationModePool
	netConn.NetworkConnectionIndex = 0
	netConn.IsConnected = true
	netConn.NeedsCustomization = true
	netConn.Network = d.cfg.VcdOrgVDCNetwork

	err = vm.UpdateNetworkConnectionSection(netSection)
	if err != nil {
		return err
	}

	enabled := true
	disabled := false
	vm.VM.GuestCustomizationSection.Enabled = &enabled
	vm.VM.GuestCustomizationSection.AdminPassword = d.adminPassword
	vm.VM.GuestCustomizationSection.AdminPasswordEnabled = &enabled
	vm.VM.GuestCustomizationSection.AdminPasswordAuto = &disabled
	vm.VM.GuestCustomizationSection.ResetPasswordRequired = &disabled
	// vm.VM.GuestCustomizationSection.CustomizationScript = sshCustomScript
	_, err = vm.SetGuestCustomizationSection(vm.VM.GuestCustomizationSection)
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	log.Info().Msgf("Booting up %s", d.machineName)
	task, err = vapp.PowerOn()
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	d.VAppHREF = vapp.VApp.HREF
	d.VMHREF = vm.VM.HREF

	ip, err := d.GetIP()

	log.Info().Msg("Waiting for the machine to be up")
	err = d.waitForRDPStable(ip)
	if err != nil {
		return err
	}

	log.Info().Msgf("Waiting for SSH to be available")
	for i := 0; i < 10; i++ {
		// fmt.Printf("Attempt %d", i
		err = drivers.WaitForSSH(d)
	}
	if err != nil {
		return err
	}

	log.Debug().Msg("SSH is available")
	return nil
}

func (d *VcdDriver) GetSSHClientFromDriver() (ssh.Client, error) {
	auth := ssh.Auth{
		Passwords: []string{d.adminPassword},
	}

	vm, err := d.getVM()
	if err != nil {
		return nil, err
	}
	var user string
	if strings.Contains(vm.VM.VmSpecSection.OsType, "windows") {
		user = "Administrator"
	} else {
		user = "root"
	}

	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}

	client, err := ssh.NewClient(user, ip, SSHPort, &auth)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (d *VcdDriver) Destroy() error {
	vapp, err := d.getVApp()
	if err != nil {
		return err
	}

	task, err := vapp.PowerOff()
	if err == nil {
		log.Info().Msg("Powering off...")
		if err = task.WaitTaskCompletion(); err != nil {
			log.Warn().Msg("Error powering off")
		}
	}

	task, err = vapp.Undeploy()
	if err == nil {
		if err = task.WaitTaskCompletion(); err != nil {
			log.Warn().Msg("Error undeploying")
		}
	}

	task, err = vapp.Delete()
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	return nil
}

func (d *VcdDriver) GetOS() (drivers.OStype, error) {
	vm, err := d.getVM()
	if err != nil {
		return "", err
	}
	if strings.Contains(vm.VM.VmSpecSection.OsType, "windows") {
		return drivers.Windows, nil
	} else {
		return drivers.Linux, nil
	}
}

func (d *VcdDriver) GetIP() (string, error) {
	vm, err := d.getVM()
	if err != nil {
		return "", err
	}

	// We assume that the vApp has only one VM with only one NIC
	if vm.VM.NetworkConnectionSection != nil {
		networks := vm.VM.NetworkConnectionSection.NetworkConnection
		for _, n := range networks {
			if n.ExternalIPAddress != "" {
				return n.ExternalIPAddress, nil
			}
			if n.IPAddress != "" { // perhaps this is too opinionated ?
				return n.IPAddress, nil
			}
		}
	}

	return "", fmt.Errorf("could not get public IP")
}
