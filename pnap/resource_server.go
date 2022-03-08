package pnap

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/PNAP/go-sdk-helper-bmc/command/bmcapi/server"
	"github.com/PNAP/go-sdk-helper-bmc/receiver"

	bmcapiclient "github.com/phoenixnap/go-sdk-bmc/bmcapi"
)

const (
	pnapRetryTimeout       = 100 * time.Minute
	pnapDeleteRetryTimeout = 15 * time.Minute
	pnapRetryDelay         = 5 * time.Second
	pnapRetryMinTimeout    = 3 * time.Second
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(pnapRetryTimeout),
			Update: schema.DefaultTimeout(pnapRetryTimeout),
			Delete: schema.DefaultTimeout(pnapDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_ip_addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ip_addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"os": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ssh_keys": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"location": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cores_per_cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_frequency_in_ghz": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"action": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"install_default_ssh_keys": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ssh_key_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"reservation_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pricing_model": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"rdp_allowed_ips": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"management_ui_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_password": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				//Sensitive: true,
			},
			"management_access_allowed_ips": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"provisioned_on": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_configuration": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_network_configuration": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gateway_address": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"configuration_type": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
										Default:  nil,
									},
									"private_networks": &schema.Schema{
										Type:     schema.TypeList,
										Computed: true,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_private_network": &schema.Schema{
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": &schema.Schema{
																Type:     schema.TypeString,
																Required: true,
															},
															"ips": &schema.Schema{
																Type:     schema.TypeSet,
																Optional: true,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"dhcp": &schema.Schema{
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
																Default:  nil,
															},
															"status_description": &schema.Schema{
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"ip_blocks_configuration": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"configuration_type": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
									},
									"ip_blocks": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_ip_block": &schema.Schema{
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": &schema.Schema{
																Type:     schema.TypeString,
																Required: true,
															},
															"vlan_id": &schema.Schema{
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(receiver.BMCSDK)

	request := &bmcapiclient.ServerCreate{}
	request.Hostname = d.Get("hostname").(string)
	var desc = d.Get("description").(string)
	if len(desc) > 0 {
		request.Description = &desc
	}
	request.Os = d.Get("os").(string)
	request.Type = d.Get("type").(string)
	request.Location = d.Get("location").(string)
	var networkType = d.Get("network_type").(string)

	if len(networkType) > 0 {
		request.NetworkType = &networkType
	}

	var resId = d.Get("reservation_id").(string)
	if len(resId) > 0 {
		request.ReservationId = &resId
	}

	var prModel = d.Get("pricing_model").(string)
	if len(prModel) > 0 {
		request.PricingModel = &prModel
	}

	var installDefault = d.Get("install_default_ssh_keys").(bool)
	request.InstallDefaultSshKeys = &installDefault
	temp := d.Get("ssh_keys").(*schema.Set).List()
	keys := make([]string, len(temp))
	for i, v := range temp {
		keys[i] = fmt.Sprint(v)
	}
	//todo
	request.SshKeys = &keys

	temp1 := d.Get("ssh_key_ids").(*schema.Set).List()
	keyIds := make([]string, len(temp1))
	for i, v := range temp1 {
		keyIds[i] = fmt.Sprint(v)
	}
	//todo
	request.SshKeyIds = &keyIds

	temp2 := d.Get("rdp_allowed_ips").(*schema.Set).List()
	allowedIps := make([]string, len(temp2))
	for i, v := range temp2 {
		allowedIps[i] = fmt.Sprint(v)
	}
	temp3 := d.Get("management_access_allowed_ips").(*schema.Set).List()
	managementAccessAllowedIps := make([]string, len(temp3))
	for i, v := range temp3 {
		managementAccessAllowedIps[i] = fmt.Sprint(v)
	}
	if len(temp2) > 0 || len(temp3) > 0 {
		dtoOsConfiguration := bmcapiclient.OsConfiguration{}

		if len(temp2) > 0 {
			dtoWindows := bmcapiclient.OsConfigurationWindows{}
			dtoWindows.RdpAllowedIps = &allowedIps
			dtoOsConfiguration.Windows = &dtoWindows
		}
		if len(temp3) > 0 {
			dtoOsConfiguration.ManagementAccessAllowedIps = &managementAccessAllowedIps
		}
		request.OsConfiguration = &dtoOsConfiguration
	}

	// private network block
	if d.Get("network_configuration") != nil && len(d.Get("network_configuration").([]interface{})) > 0 {

		networkConfiguration := d.Get("network_configuration").([]interface{})[0]
		networkConfigurationItem := networkConfiguration.(map[string]interface{})

		networkConfigurationObject := bmcapiclient.NetworkConfiguration{}

		if networkConfigurationItem["private_network_configuration"] != nil && len(networkConfigurationItem["private_network_configuration"].([]interface{})) > 0 {
			privateNetworkConfiguration := networkConfigurationItem["private_network_configuration"].([]interface{})[0]
			privateNetworkConfigurationItem := privateNetworkConfiguration.(map[string]interface{})

			gatewayAddress := privateNetworkConfigurationItem["gateway_address"].(string)
			configurationType := privateNetworkConfigurationItem["configuration_type"].(string)
			privateNetworks := privateNetworkConfigurationItem["private_networks"].([]interface{})

			if len(gatewayAddress) > 0 || len(configurationType) > 0 || len(privateNetworks) > 0 {
				privateNetworkConfigurationObject := bmcapiclient.PrivateNetworkConfiguration{}
				if len(gatewayAddress) > 0 {
					privateNetworkConfigurationObject.GatewayAddress = &gatewayAddress
				}

				if len(configurationType) > 0 {
					privateNetworkConfigurationObject.ConfigurationType = &configurationType
				}

				networkConfigurationObject.PrivateNetworkConfiguration = &privateNetworkConfigurationObject
				if len(privateNetworks) > 0 {

					serPrivateNets := make([]bmcapiclient.ServerPrivateNetwork, len(privateNetworks))

					for k, j := range privateNetworks {
						serverPrivateNetworkObject := bmcapiclient.ServerPrivateNetwork{}

						privateNetworkItem := j.(map[string]interface{})

						serverPrivateNetwork := privateNetworkItem["server_private_network"].([]interface{})[0]
						serverPrivateNetworkItem := serverPrivateNetwork.(map[string]interface{})

						id := serverPrivateNetworkItem["id"].(string)
						tempIps := serverPrivateNetworkItem["ips"].(*schema.Set).List()

						NetIps := make([]string, len(tempIps))
						for i, v := range tempIps {
							NetIps[i] = fmt.Sprint(v)
						}
						dhcp := serverPrivateNetworkItem["dhcp"].(bool)

						if (len(id)) > 0 {
							serverPrivateNetworkObject.Id = id
						}
						if (len(NetIps)) > 0 {
							serverPrivateNetworkObject.Ips = &NetIps
						}

						serverPrivateNetworkObject.Dhcp = &dhcp
						serPrivateNets[k] = serverPrivateNetworkObject

					}
					privateNetworkConfigurationObject.PrivateNetworks = &serPrivateNets
				}
			}
		}
		if networkConfigurationItem["ip_blocks_configuration"] != nil && len(networkConfigurationItem["ip_blocks_configuration"].([]interface{})) > 0 {
			ipBlocksConfiguration := networkConfigurationItem["ip_blocks_configuration"].([]interface{})[0]
			ipBlocksConfigurationItem := ipBlocksConfiguration.(map[string]interface{})

			confType := ipBlocksConfigurationItem["configuration_type"].(string)
			ipBlocks := ipBlocksConfigurationItem["ip_blocks"].([]interface{})

			if len(confType) > 0 || len(ipBlocks) > 0 {
				ipBlocksConfigurationObject := bmcapiclient.IpBlocksConfiguration{}
				if len(confType) > 0 {
					ipBlocksConfigurationObject.ConfigurationType = &confType
				}

				networkConfigurationObject.IpBlocksConfiguration = &ipBlocksConfigurationObject
				if len(ipBlocks) > 0 {

					serIpBlocks := make([]bmcapiclient.ServerIpBlock, len(ipBlocks))

					for k, j := range ipBlocks {
						serverIpBlockObject := bmcapiclient.ServerIpBlock{}

						ipBlockItem := j.(map[string]interface{})

						serverIpBlock := ipBlockItem["server_ip_block"].([]interface{})[0]
						serverIpBlockItem := serverIpBlock.(map[string]interface{})

						id := serverIpBlockItem["id"].(string)
						vlanId := int32(serverIpBlockItem["vlan_id"].(int))

						if (len(id)) > 0 {
							serverIpBlockObject.Id = id
						}
						serverIpBlockObject.VlanId = &vlanId
						serIpBlocks[k] = serverIpBlockObject
					}
					ipBlocksConfigurationObject.IpBlocks = &serIpBlocks
				}
			}
		}
		request.NetworkConfiguration = &networkConfigurationObject
		b, _ := json.MarshalIndent(request, "", "  ")
		log.Printf("request object is" + string(b))
	}

	// end of private network block
	requestCommand := server.NewCreateServerCommand(client, *request)

	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	} else {

		d.SetId(resp.Id)
		d.Set("password", resp.Password)
		if resp.OsConfiguration != nil {
			d.Set("root_password", resp.OsConfiguration.RootPassword)
			d.Set("management_ui_url", resp.OsConfiguration.ManagementUiUrl)
		}

		waitResultError := resourceWaitForCreate(resp.Id, &client)
		if waitResultError != nil {
			return waitResultError
		}
	}
	/* code := resp.StatusCode
	if code == 200 {
		response := &dto.LongServer{}
		response.FromBytes(resp)
		d.SetId(response.ID)
		d.Set("password", response.Password)
		if(&response.OsConfiguration != nil){
			d.Set("root_password", response.OsConfiguration.RootPassword)
			d.Set("management_ui_url", response.OsConfiguration.ManagementUiUrl)
		}

		waitResultError := resourceWaitForCreate(response.ID, &client)
		if waitResultError != nil {
			return waitResultError
		}
	} else {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API create server Returned Code %v Message: %s Validation Errors: %s", code, response.Message, response.ValidationErrors)
	} */

	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	serverID := d.Id()
	requestCommand := server.NewGetServerCommand(client, serverID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	/* code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	response := &dto.LongServer{}
	response.FromBytes(resp) */
	//d.SetId(resp.Id)
	d.Set("status", resp.Status)
	d.Set("hostname", resp.Hostname)
	d.Set("description", resp.Description)
	d.Set("os", resp.Os)
	d.Set("type", resp.Type)
	d.Set("location", resp.Location)
	d.Set("cpu", resp.Cpu)
	d.Set("cpu_count", resp.CpuCount)
	d.Set("cores_per_cpu", resp.CoresPerCpu)
	d.Set("cpu_frequency_in_ghz", resp.CpuFrequency)
	d.Set("ram", resp.Ram)
	d.Set("storage", resp.Storage)
	d.Set("network_type", resp.NetworkType)
	d.Set("action", "")
	var privateIPs []interface{}
	for _, v := range resp.PrivateIpAddresses {
		privateIPs = append(privateIPs, v)
	}
	d.Set("private_ip_addresses", privateIPs)
	var publicIPs []interface{}
	for _, k := range resp.PublicIpAddresses {
		publicIPs = append(publicIPs, k)
	}
	d.Set("public_ip_addresses", publicIPs)
	d.Set("reservation_id", resp.ReservationId)
	d.Set("pricing_model", resp.PricingModel)

	d.Set("cluster_id", resp.ClusterId)
	if resp.OsConfiguration != nil && resp.OsConfiguration.ManagementAccessAllowedIps != nil {
		var mgmntAccessAllowedIps []interface{}
		for _, k := range *resp.OsConfiguration.ManagementAccessAllowedIps {
			mgmntAccessAllowedIps = append(mgmntAccessAllowedIps, k)
		}
		d.Set("management_access_allowed_ips", mgmntAccessAllowedIps)
	}

	if resp.OsConfiguration != nil && resp.OsConfiguration.Windows != nil && resp.OsConfiguration.Windows.RdpAllowedIps != nil {
		var rdpAllowedIps []interface{}
		for _, k := range *resp.OsConfiguration.Windows.RdpAllowedIps {
			rdpAllowedIps = append(rdpAllowedIps, k)
		}
		d.Set("rdp_allowed_ips", rdpAllowedIps)
	}

	d.Set("provisioned_on", resp.ProvisionedOn.String())

	var ncInput = d.Get("network_configuration").([]interface{})
	networkConfiguration := flattenNetworkConfiguration(&resp.NetworkConfiguration, ncInput)

	if err := d.Set("network_configuration", networkConfiguration); err != nil {
		return err
	}

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("action") {
		client := m.(receiver.BMCSDK)
		//var requestCommand helpercommand.Executor
		newStatus := d.Get("action").(string)

		switch newStatus {
		case "powered-on":
			//do power-on request
			serverID := d.Id()
			requestCommand := server.NewPowerOnServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerON(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "powered-off":
			//power off request

			serverID := d.Id()

			requestCommand := server.NewPowerOffServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerOff(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "reboot":
			//reboot

			serverID := d.Id()

			requestCommand := server.NewRebootServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForCreate(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "reset":
			//reset
			request := &bmcapiclient.ServerReset{}
			temp := d.Get("ssh_keys").(*schema.Set).List()
			keys := make([]string, len(temp))
			for i, v := range temp {
				keys[i] = fmt.Sprint(v)
			}
			request.SshKeys = &keys
			request.InstallDefaultSshKeys = d.Get("install_default_ssh_keys").(*bool)

			temp1 := d.Get("ssh_key_ids").(*schema.Set).List()
			keyIds := make([]string, len(temp1))
			for i, v := range temp1 {
				keyIds[i] = fmt.Sprint(v)
			}
			request.SshKeyIds = &keyIds

			dtoOsConfiguration := bmcapiclient.OsConfigurationMap{}
			isWindows := strings.Contains(d.Get("os").(string), "windows")
			isEsxi := strings.Contains(d.Get("os").(string), "esxi")

			if isWindows {
				//log.Printf("Waiting for server windows to be reseted...")
				dtoWindows := bmcapiclient.OsConfigurationWindows{}
				temp2 := d.Get("rdp_allowed_ips").(*schema.Set).List()
				allowedIps := make([]string, len(temp2))
				for i, v := range temp2 {
					allowedIps[i] = fmt.Sprint(v)
				}

				dtoWindows.RdpAllowedIps = &allowedIps
				dtoOsConfiguration.Windows = &dtoWindows
				dtoOsConfiguration.Esxi = nil
				request.OsConfiguration = &dtoOsConfiguration
			}

			if isEsxi {
				//log.Printf("Waiting for server esxi to be reseted...")
				dtoEsxi := bmcapiclient.OsConfigurationMapEsxi{}
				temp3 := d.Get("management_access_allowed_ips").(*schema.Set).List()
				managementAccessAllowedIps := make([]string, len(temp3))
				for i, v := range temp3 {
					managementAccessAllowedIps[i] = fmt.Sprint(v)
				}
				dtoEsxi.ManagementAccessAllowedIps = &managementAccessAllowedIps
				dtoOsConfiguration.Esxi = &dtoEsxi
				dtoOsConfiguration.Windows = nil
				request.OsConfiguration = &dtoOsConfiguration

			}

			//b, err := json.MarshalIndent(request, "", "  ")
			//log.Printf("request object is" + string(b))
			//request.Id = d.Id()
			requestCommand := server.NewResetServerCommand(client, d.Id(), *request)
			resp, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			d.Set("password", resp.Password)

			if resp.OsConfiguration != nil && resp.OsConfiguration.Esxi != nil {
				d.Set("root_password", resp.OsConfiguration.Esxi.RootPassword)
				d.Set("management_ui_url", resp.OsConfiguration.Esxi.ManagementUiUrl)
			}

			waitResultError := resourceWaitForCreate(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}

		case "shutdown":

			serverID := d.Id()

			requestCommand := server.NewShutDownServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerOff(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}

		case "default":
			return fmt.Errorf("Unsuported action")
		}

	} else if d.HasChange("pricing_model") {
		client := m.(receiver.BMCSDK)
		//var requestCommand command.Executor
		//reserve action
		request := &bmcapiclient.ServerReserve{}
		//request.Id = d.Id()
		request.PricingModel = d.Get("pricing_model").(string)

		requestCommand := server.NewReserveServerCommand(client, d.Id(), *request)
		_, err := requestCommand.Execute()
		if err != nil {
			return err
		}
		/* 	code := resp.StatusCode
		if code != 200 {
			response := &dto.ErrorMessage{}
			response.FromBytes(resp)
			return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
		} */
	} else {
		return fmt.Errorf("Unsuported action")
	}
	return resourceServerRead(d, m)

}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)

	serverID := d.Id()
	relinquishIpBlock := bmcapiclient.RelinquishIpBlock{}
	deleteIpBlocks := false
	relinquishIpBlock.DeleteIpBlocks = &deleteIpBlocks
	b, _ := json.MarshalIndent(relinquishIpBlock, "", "  ")
	log.Printf("relinquishIpBlock object is" + string(b))
	requestCommand := server.NewDeprovisionServerCommand(client, serverID, relinquishIpBlock)

	_, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	/* code := resp.StatusCode
	if code != 200 && code != 404 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	} */
	return nil
}

func resourceWaitForCreate(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to be created...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating", "resetting", "rebooting"},
		Target:     []string{"powered-on", "powered-off"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to switch to target state: %v", id, err)
	}

	return nil
}

func resourceWaitForPowerON(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to power on...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"powered-off"},
		Target:     []string{"powered-on"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to power on: %v", id, err)
	}

	return nil
}

func resourceWaitForPowerOff(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to power off...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"powered-on"},
		Target:     []string{"powered-off"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to power off: %v", id, err)
	}

	return nil
}

func refreshForCreate(client *receiver.BMCSDK, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		requestCommand := server.NewGetServerCommand(*client, id)
		/* requestCommand.SetRequester(client)
		serverID := id
		requestCommand.SetServerID(serverID) */
		resp, err := requestCommand.Execute()
		if err != nil {
			return 0, "", err
		} else {
			return 0, resp.Status, nil
		}
		/* 	code := resp.StatusCode
		if code != 200 {
			response := &dto.ErrorMessage{}
			response.FromBytes(resp)
			return 0, "", fmt.Errorf("API refressh for create Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
		}
		response := &dto.LongServer{}
		response.FromBytes(resp)
		return 0, response.Status, nil*/
	}
}

/* func run(command command.Executor) error {
	resp, err := command.Execute()
	if err != nil {
		return err
	}
	code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	return nil
}

func runResetCommand(command command.Executor) (error, dto.ServerActionResponse) {
	resp, err := command.Execute()
	if err != nil {
		return err, dto.ServerActionResponse{}
	}
	code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors), dto.ServerActionResponse{}
	}
	if code == 200 {
		response := &dto.ServerActionResponse{}
		response.FromBytes(resp)
		return nil, *response
	}
	return nil, dto.ServerActionResponse{}
} */

func flattenNetworkConfiguration(netConf *bmcapiclient.NetworkConfiguration, ncInput []interface{}) []interface{} {
	if netConf != nil { //len(ncInput)
		if len(ncInput) == 0 {
			ncInput = make([]interface{}, 1)
			n := make(map[string]interface{})
			ncInput[0] = n
		}
		nci := ncInput[0]
		nciMap := nci.(map[string]interface{})

		if netConf != nil {
			if netConf.PrivateNetworkConfiguration != nil {
				prNetConf := *netConf.PrivateNetworkConfiguration
				pnc := make([]interface{}, 1)
				pncItem := make(map[string]interface{})

				if prNetConf.GatewayAddress != nil {
					pncItem["gateway_adress"] = *prNetConf.GatewayAddress
				}
				if prNetConf.ConfigurationType != nil {
					pncItem["configuration_type"] = *prNetConf.ConfigurationType
				}
				if prNetConf.PrivateNetworks != nil {
					prNet := *prNetConf.PrivateNetworks
					pn := make([]interface{}, len(prNet))
					for i, j := range prNet {
						pnItem := make(map[string]interface{})
						spn := make([]interface{}, 1)
						spnItem := make(map[string]interface{})

						spnItem["id"] = j.Id
						if j.Ips != nil {
							ips := make([]interface{}, len(*j.Ips))
							for k, l := range *j.Ips {
								ips[k] = l
							}
							spnItem["ips"] = ips
						}
						if j.Dhcp != nil {
							spnItem["dhcp"] = *j.Dhcp
						}
						if j.StatusDescription != nil {
							spnItem["status_description"] = *j.StatusDescription
						}
						spn[0] = spnItem
						pnItem["server_private_network"] = spn
						pn[i] = pnItem
					}
					pncItem["private_networks"] = pn
				}
				pnc[0] = pncItem
				nciMap["private_network_configuration"] = pnc
			}

			if netConf.IpBlocksConfiguration != nil {

				ipBlocksConf := *netConf.IpBlocksConfiguration

				if ipBlocksConf.IpBlocks != nil {

					ibc := nciMap["ip_blocks_configuration"]

					if ibc == nil || len(ibc.([]interface{})) == 0 {
						ibc = make([]interface{}, 1)
						ibci := make(map[string]interface{})
						ibc.([]interface{})[0] = ibci
					}

					ibci := ibc.([]interface{})[0]
					ibcInput := ibci.(map[string]interface{})

					ipBlocks := *ipBlocksConf.IpBlocks
					ib := make([]interface{}, len(ipBlocks))
					for i, j := range ipBlocks {
						ibItem := make(map[string]interface{})
						sib := make([]interface{}, 1)
						sibItem := make(map[string]interface{})

						sibItem["id"] = j.Id
						if j.VlanId != nil {
							sibItem["vlan_id"] = *j.VlanId
						}
						sib[0] = sibItem
						ibItem["server_ip_block"] = sib
						ib[i] = ibItem
					}
					ibcInput["ip_blocks"] = ib
				}
			}
			//return ncInput
		}
	}
	return ncInput
}
