package vcd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func (d *VcdDriver) getVApp() (*govcd.VApp, error) {
	if d.VAppHREF != "" { // this is way quicker
		vapp := govcd.NewVApp(&d.client.Client)
		vapp.VApp.HREF = d.VAppHREF
		err := vapp.Refresh()
		if err != nil {
			return nil, err
		}
		return vapp, nil
	}

	org, err := d.client.GetOrgByName(d.cfg.VcdOrg)
	if err != nil {
		return nil, err
	}
	vdc, err := org.GetVDCByName(d.cfg.VcdVdc, false)
	if err != nil {
		return nil, err
	}
	vapp, err := vdc.GetVAppByName(d.machineName, true)
	if err != nil {
		return nil, err
	}
	d.VAppHREF = vapp.VApp.HREF
	return vapp, nil
}

func (d *VcdDriver) getVM() (*govcd.VM, error) {
	if d.VMHREF != "" {
		vm := govcd.NewVM(&d.client.Client)
		vm.VM.HREF = d.VMHREF
		err := vm.Refresh()
		if err != nil {
			return nil, err
		}
		return vm, nil
	}

	vapp, err := d.getVApp()
	if err != nil {
		return nil, err
	}

	if len(vapp.VApp.Children.VM) != 1 {
		return nil, fmt.Errorf("VM count != 1")
	}
	vm := govcd.NewVM(&d.client.Client)
	vm.VM.HREF = vapp.VApp.Children.VM[0].HREF
	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	d.VMHREF = vm.VM.HREF
	return vm, nil
}

func newClient(apiURL url.URL, user, password, org string, insecure bool) (*govcd.VCDClient, error) {
	vcdclient := &govcd.VCDClient{
		Client: govcd.Client{
			APIVersion: "36.3",
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
