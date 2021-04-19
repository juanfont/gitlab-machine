package drivers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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
}

type VcdDriver struct {
	cfg    VcdDriverConfig
	client *govcd.VCDClient
}

func NewVcdDriver(cfg VcdDriverConfig) (*VcdDriver, error) {
	u, err := url.ParseRequestURI(cfg.VcdURL)
	if err != nil {
		return nil, err
	}
	c, err := newClient(*u, cfg.VcdUser, cfg.VcdPassword, cfg.VcdOrg, cfg.VcdInsecure)
	if err != nil {
		return nil, err
	}

	d := VcdDriver{
		cfg:    cfg,
		client: c,
	}

	return &d, nil
}

func (d *VcdDriver) Create(instanceName string) error {
	org, err := d.client.GetOrgByName(d.cfg.VcdOrg)
	if err != nil {
		return err
	}
	vdc, err := org.GetVDCByName(d.cfg.VcdVdc, false)
	if err != nil {
		return err
	}

	log.Printf("Finding network...")
	net, err := vdc.GetOrgVdcNetworkByName(d.cfg.VcdOrgVDCNetwork, true)
	if err != nil {
		return err
	}

	log.Printf("Finding catalog...")
	catalog, err := org.GetCatalogByName(d.cfg.Catalog, true)
	if err != nil {
		return err
	}

	log.Printf("Finding template...")
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
			return fmt.Errorf("No storage profile available")
		}
		storageProfile = *(vdc.Vdc.VdcStorageProfiles.VdcStorageProfile[0])
		if err != nil {
			return err
		}
	}

	log.Printf("Creating a new vApp: %s...", instanceName)
	networks := []*types.OrgVDCNetwork{}
	networks = append(networks, net.OrgVDCNetwork)
	task, err := vdc.ComposeVApp(
		networks,
		vapptemplate,
		storageProfile,
		instanceName,
		d.cfg.Description,
		true)

	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	vapp, err := vdc.GetVAppByName(instanceName, true)
	if err != nil {
		return err
	}
	//d.VAppHREF = vapp.VApp.HREF

	if len(vapp.VApp.Children.VM) != 1 {
		return fmt.Errorf("VM count != 1")
	}

	vm := govcd.NewVM(&d.client.Client)
	vm.VM.HREF = vapp.VApp.Children.VM[0].HREF
	err = vm.Refresh()
	if err != nil {
		return err
	}
	log.Printf("Found VM: %s...", vm.VM.Name)
	//d.VMHREF = vm.VM.HREF

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
			return fmt.Errorf("Error waiting for vApp deploy")
		}
	case <-time.After(15 * time.Minute):
		return fmt.Errorf("Reached timeout while deploying VM")
	}

	if vm.VM.VmSpecSection == nil {
		return fmt.Errorf("VM Spec Section empty")
	}
	vm.Refresh()

	vm.VM.VmSpecSection.MemoryResourceMb.Configured = int64(d.cfg.MemorySizeMb)
	vm.VM.VmSpecSection.NumCpus = &d.cfg.NumCpus
	vm.VM.VmSpecSection.NumCoresPerSocket = &d.cfg.CoresPerSocket

	log.Printf("Updating virtual hardware specs...")
	vm, err = vm.UpdateVmSpecSection(vm.VM.VmSpecSection, d.cfg.Description)
	if err != nil {
		return err
	}

	log.Printf("Configuring network...")
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

	vm.UpdateNetworkConnectionSection(netSection)

	log.Printf("Setting up guest customization...")
	// sshCustomScript, err := d.getGuestCustomizationScript()
	if err != nil {
		return err
	}

	enabled := true
	vm.VM.GuestCustomizationSection.Enabled = &enabled
	// vm.VM.GuestCustomizationSection.CustomizationScript = sshCustomScript
	_, err = vm.SetGuestCustomizationSection(vm.VM.GuestCustomizationSection)
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	log.Printf("Booting up %s...", instanceName)
	task, err = vapp.PowerOn()
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	// d.VAppHREF = vapp.VApp.HREF
	// d.VMHREF = vm.VM.HREF

	return nil
}

func (d *VcdDriver) RunCommand(instanceName string) error {
	return nil
}
func (d *VcdDriver) Destroy(instanceName string) error {
	return nil
}

func newClient(apiURL url.URL, user, password, org string, insecure bool) (*govcd.VCDClient, error) {
	vcdclient := &govcd.VCDClient{
		Client: govcd.Client{
			APIVersion: "32.0", // supported by 9.5, 9.7, 10.0, 10.1
			VCDHREF:    apiURL,
			Http: http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
					},
					Proxy:               http.ProxyFromEnvironment,
					TLSHandshakeTimeout: 120 * time.Second, // Default timeout for TSL hand shake
				},
				Timeout: 1200 * time.Second, // Default value for http request+response timeout
			},
			MaxRetryTimeout: 60, // Default timeout in seconds for retries calls in functions
		},
	}
	err := vcdclient.Authenticate(user, password, org)
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate to Org \"%s\": %s", org, err)
	}
	return vcdclient, nil
}
